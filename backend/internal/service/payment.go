package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	paystack "storefront/backend/internal/adapter/paystack"
	"storefront/backend/internal/apperr"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var ErrPaymentVerificationFailed = apperr.New(http.StatusBadGateway, "payment verification failed")
var ErrPaymentAmountMismatch = apperr.New(http.StatusConflict, "payment amount does not match order total")
var ErrPaymentReferenceMismatch = apperr.New(http.StatusConflict, "payment reference does not match order")

// PaystackClient is the subset of paystack.Client used by PaymentService.
type PaystackClient interface {
	InitializeTransaction(ctx context.Context, req paystack.InitializeRequest) (*paystack.InitializeResponse, error)
	VerifyTransaction(ctx context.Context, reference string) (*paystack.VerifyResponse, error)
}

type PaymentService struct {
	paystack  PaystackClient
	orders    repository.OrderRepository
	products  repository.ProductRepository
	walletSvc *WalletService
	pool      TxBeginner
}

type orderLockingRepository interface {
	GetByIDInternalForUpdate(ctx context.Context, id uuid.UUID) (*models.Order, error)
}

func NewPaymentService(
	paystackClient PaystackClient,
	orders repository.OrderRepository,
	products repository.ProductRepository,
	walletSvc *WalletService,
) *PaymentService {
	return &PaymentService{paystack: paystackClient, orders: orders, products: products, walletSvc: walletSvc}
}

func (s *PaymentService) SetPool(pool TxBeginner) { s.pool = pool }

// InitiatePayment creates a Paystack transaction and returns the redirect URL.
// The Paystack reference equals the order UUID so webhook callbacks can reverse-map it.
func (s *PaymentService) InitiatePayment(ctx context.Context, order *models.Order, customerEmail, subaccountCode, callbackURL string) (string, error) {
	req := paystack.InitializeRequest{
		Email:          customerEmail,
		Amount:         order.TotalAmount.Add(order.ShippingFee),
		Reference:      order.ID.String(),
		CallbackURL:    callbackURL,
		SubaccountCode: subaccountCode,
		Metadata:       map[string]any{"order_id": order.ID},
	}
	resp, err := s.paystack.InitializeTransaction(ctx, req)
	if err != nil {
		return "", fmt.Errorf("initiate payment: %w", err)
	}
	return resp.AuthorizationURL, nil
}

// HandleChargeSuccess verifies the Paystack charge event and credits the tenant wallet.
// Delivery orders settle to pending balance until delivery is confirmed.
// No-delivery orders settle directly to available balance and become completed.
// reference is the order UUID string used as Paystack reference at initialization.
func (s *PaymentService) HandleChargeSuccess(ctx context.Context, reference string) error {
	resp, err := s.paystack.VerifyTransaction(ctx, reference)
	if err != nil {
		return fmt.Errorf("verify transaction: %w", err)
	}
	if resp.Status != "success" {
		return ErrPaymentVerificationFailed
	}
	if resp.Reference != reference {
		return ErrPaymentReferenceMismatch
	}

	orderID, err := uuid.Parse(reference)
	if err != nil {
		return fmt.Errorf("invalid reference: %w", err)
	}

	// Atomic: update payment status + credit wallet in a single DB transaction.
	if s.pool != nil {
		dbTx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer dbTx.Rollback(ctx) //nolint:errcheck

		txOrders := s.orders.WithTx(dbTx)
		lockingOrders, ok := txOrders.(orderLockingRepository)
		if !ok {
			return fmt.Errorf("order repository does not support row locking")
		}

		order, err := lockingOrders.GetByIDInternalForUpdate(ctx, orderID)
		if err != nil {
			return fmt.Errorf("get order: %w", err)
		}

		if order.PaymentStatus == models.PaymentStatusPaid {
			return nil
		}
		if order.PaymentStatus != models.PaymentStatusPending {
			return nil
		}

		expectedAmount := order.TotalAmount.Add(order.ShippingFee)
		if !resp.Amount.Equal(expectedAmount) {
			return ErrPaymentAmountMismatch
		}

		if err := txOrders.UpdatePaymentStatus(ctx, order.TenantID, orderID, models.PaymentStatusPaid); err != nil {
			return fmt.Errorf("update payment status: %w", err)
		}

		if !order.IsDelivery {
			if err := txOrders.UpdateFulfillmentStatus(ctx, order.TenantID, orderID, models.FulfillmentStatusCompleted); err != nil {
				return fmt.Errorf("update fulfillment status: %w", err)
			}
		}

		if _, err := s.walletSvc.CreditSaleWithTx(
			ctx,
			dbTx,
			order.TenantID,
			order.TotalAmount.Add(order.ShippingFee),
			order.TotalAmount,
			order.IsDelivery,
			&orderID,
		); err != nil {
			return fmt.Errorf("credit wallet: %w", err)
		}

		return dbTx.Commit(ctx)
	}

	order, err := s.orders.GetByIDInternal(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	if order.PaymentStatus == models.PaymentStatusPaid {
		return nil
	}
	if order.PaymentStatus != models.PaymentStatusPending {
		return nil
	}

	expectedAmount := order.TotalAmount.Add(order.ShippingFee)
	if !resp.Amount.Equal(expectedAmount) {
		return ErrPaymentAmountMismatch
	}

	// Fallback (no pool): non-atomic path for tests / simple setups.
	if err := s.orders.UpdatePaymentStatus(ctx, order.TenantID, orderID, models.PaymentStatusPaid); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	if !order.IsDelivery {
		if err := s.orders.UpdateFulfillmentStatus(ctx, order.TenantID, orderID, models.FulfillmentStatusCompleted); err != nil {
			return fmt.Errorf("update fulfillment status: %w", err)
		}
	}

	if _, err := s.walletSvc.CreditSale(
		ctx,
		order.TenantID,
		order.TotalAmount.Add(order.ShippingFee),
		order.TotalAmount,
		order.IsDelivery,
		&orderID,
	); err != nil {
		return fmt.Errorf("credit wallet: %w", err)
	}

	return nil
}

// HandleChargeFailed marks the order as payment-failed, cancels fulfillment, and restores stock.
func (s *PaymentService) HandleChargeFailed(ctx context.Context, reference string) error {
	orderID, err := uuid.Parse(reference)
	if err != nil {
		return fmt.Errorf("invalid reference: %w", err)
	}

	if s.pool != nil {
		dbTx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer dbTx.Rollback(ctx) //nolint:errcheck

		txOrders := s.orders.WithTx(dbTx)
		lockingOrders, ok := txOrders.(orderLockingRepository)
		if !ok {
			return fmt.Errorf("order repository does not support row locking")
		}

		order, err := lockingOrders.GetByIDInternalForUpdate(ctx, orderID)
		if err != nil {
			return fmt.Errorf("get order: %w", err)
		}

		if order.PaymentStatus != models.PaymentStatusPending {
			return nil
		}

		if err := txOrders.UpdatePaymentStatus(ctx, order.TenantID, orderID, models.PaymentStatusFailed); err != nil {
			return fmt.Errorf("update payment status: %w", err)
		}
		if err := txOrders.UpdateFulfillmentStatus(ctx, order.TenantID, orderID, models.FulfillmentStatusCancelled); err != nil {
			return fmt.Errorf("update fulfillment status: %w", err)
		}

		items, err := txOrders.ListItems(ctx, orderID)
		if err != nil {
			return fmt.Errorf("list items for restock: %w", err)
		}
		for _, item := range items {
			if err := s.products.WithTx(dbTx).RestoreStock(ctx, item.VariantID, item.Quantity); err != nil {
				return fmt.Errorf("restore stock for variant %s: %w", item.VariantID, err)
			}
		}

		return dbTx.Commit(ctx)
	}

	order, err := s.orders.GetByIDInternal(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	if order.PaymentStatus != models.PaymentStatusPending {
		return nil
	}

	if err := s.orders.UpdatePaymentStatus(ctx, order.TenantID, orderID, models.PaymentStatusFailed); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	if err := s.orders.UpdateFulfillmentStatus(ctx, order.TenantID, orderID, models.FulfillmentStatusCancelled); err != nil {
		return fmt.Errorf("update fulfillment status: %w", err)
	}

	items, err := s.orders.ListItems(ctx, orderID)
	if err != nil {
		return fmt.Errorf("list items for restock: %w", err)
	}
	for _, item := range items {
		_ = s.products.RestoreStock(ctx, item.VariantID, item.Quantity)
	}

	return nil
}

// ExpirePendingOrder fails a stale unpaid order using the same rollback path as a failed charge.
func (s *PaymentService) ExpirePendingOrder(ctx context.Context, orderID uuid.UUID) error {
	return s.HandleChargeFailed(ctx, orderID.String())
}

// SweepExpiredPendingOrders expires stale online orders in small batches.
func (s *PaymentService) SweepExpiredPendingOrders(ctx context.Context, pool TxQueryer, ttl time.Duration, batchSize int) (int, error) {
	if ttl <= 0 {
		return 0, nil
	}
	if batchSize <= 0 {
		batchSize = 100
	}

	cutoff := time.Now().Add(-ttl)
	rows, err := pool.Query(ctx, `
		SELECT id
		FROM orders
		WHERE payment_method = 'online'
		  AND payment_status = 'pending'
		  AND fulfillment_status = 'processing'
		  AND created_at <= $1
		ORDER BY created_at ASC
		LIMIT $2`, cutoff, batchSize)
	if err != nil {
		return 0, fmt.Errorf("list stale pending orders: %w", err)
	}
	defer rows.Close()

	var expired int
	for rows.Next() {
		var orderID uuid.UUID
		if err := rows.Scan(&orderID); err != nil {
			return expired, fmt.Errorf("scan stale pending order: %w", err)
		}
		if err := s.ExpirePendingOrder(ctx, orderID); err != nil {
			return expired, fmt.Errorf("expire pending order %s: %w", orderID, err)
		}
		expired++
	}

	return expired, rows.Err()
}
