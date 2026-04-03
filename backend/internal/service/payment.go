package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	paystack "storefront/backend/internal/adapter/paystack"
	"storefront/backend/internal/apperr"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var ErrPaymentVerificationFailed = apperr.New(http.StatusBadGateway, "payment verification failed")
var ErrPaymentAmountMismatch = apperr.New(http.StatusConflict, "payment amount does not match order total")

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
func (s *PaymentService) InitiatePayment(ctx context.Context, order *models.Order, customerEmail, subaccountCode string) (string, error) {
	req := paystack.InitializeRequest{
		Email:          customerEmail,
		Amount:         order.TotalAmount.Add(order.ShippingFee),
		Reference:      order.ID.String(),
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

	orderID, err := uuid.Parse(reference)
	if err != nil {
		return fmt.Errorf("invalid reference: %w", err)
	}

	order, err := s.orders.GetByIDInternal(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	if order.PaymentStatus == models.PaymentStatusPaid {
		return nil
	}

	expectedAmount := order.TotalAmount.Add(order.ShippingFee)
	if !resp.Amount.Equal(expectedAmount) {
		return ErrPaymentAmountMismatch
	}

	// Atomic: update payment status + credit wallet in a single DB transaction.
	if s.pool != nil {
		dbTx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer dbTx.Rollback(ctx) //nolint:errcheck

		if err := s.orders.WithTx(dbTx).UpdatePaymentStatus(ctx, order.TenantID, orderID, models.PaymentStatusPaid); err != nil {
			return fmt.Errorf("update payment status: %w", err)
		}

		if !order.IsDelivery {
			if err := s.orders.WithTx(dbTx).UpdateFulfillmentStatus(ctx, order.TenantID, orderID, models.FulfillmentStatusCompleted); err != nil {
				return fmt.Errorf("update fulfillment status: %w", err)
			}
		}

		if order.IsDelivery {
			if _, err := s.walletSvc.CreditWithTx(ctx, dbTx, order.TenantID, resp.Amount, &orderID); err != nil {
				return fmt.Errorf("credit wallet: %w", err)
			}
		} else {
			if _, err := s.walletSvc.CreditAvailableWithTx(ctx, dbTx, order.TenantID, resp.Amount, &orderID); err != nil {
				return fmt.Errorf("credit wallet: %w", err)
			}
		}

		return dbTx.Commit(ctx)
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

	if order.IsDelivery {
		if _, err := s.walletSvc.Credit(ctx, order.TenantID, resp.Amount, &orderID); err != nil {
			return fmt.Errorf("credit wallet: %w", err)
		}
	} else {
		if _, err := s.walletSvc.CreditAvailable(ctx, order.TenantID, resp.Amount, &orderID); err != nil {
			return fmt.Errorf("credit wallet: %w", err)
		}
	}

	return nil
}

// HandleChargeFailed marks the order as payment-failed and restores stock.
func (s *PaymentService) HandleChargeFailed(ctx context.Context, reference string) error {
	orderID, err := uuid.Parse(reference)
	if err != nil {
		return fmt.Errorf("invalid reference: %w", err)
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

	items, err := s.orders.ListItems(ctx, orderID)
	if err != nil {
		return fmt.Errorf("list items for restock: %w", err)
	}
	for _, item := range items {
		_ = s.products.RestoreStock(ctx, item.VariantID, item.Quantity)
	}

	return nil
}
