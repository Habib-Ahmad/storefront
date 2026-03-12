package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	paystack "storefront/backend/internal/adapter/paystack"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var (
	ErrPaymentVerificationFailed = errors.New("payment verification failed")
	ErrAlreadyPaid               = errors.New("order already paid")
)

// PaystackClient is the subset of paystack.Client used by PaymentService.
type PaystackClient interface {
	InitializeTransaction(ctx context.Context, req paystack.InitializeRequest) (*paystack.InitializeResponse, error)
	VerifyTransaction(ctx context.Context, reference string) (*paystack.VerifyResponse, error)
}

type PaymentService struct {
	paystack  PaystackClient
	orders    repository.OrderRepository
	walletSvc *WalletService
	tiers     repository.TierRepository
	tenants   repository.TenantRepository
}

func NewPaymentService(
	paystackClient PaystackClient,
	orders repository.OrderRepository,
	walletSvc *WalletService,
	tiers repository.TierRepository,
	tenants repository.TenantRepository,
) *PaymentService {
	return &PaymentService{paystack: paystackClient, orders: orders, walletSvc: walletSvc, tiers: tiers, tenants: tenants}
}

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

	// Idempotency guard — Paystack retries charge.success up to 3×.
	if order.PaymentStatus == models.PaymentStatusPaid {
		return ErrAlreadyPaid
	}

	if err := s.orders.UpdatePaymentStatus(ctx, order.TenantID, orderID, models.PaymentStatusPaid); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	// walletSvc.Credit looks up the wallet by tenant ID.
	// Deduct commission before crediting the merchant.
	tenant, err := s.tenants.GetByID(ctx, order.TenantID)
	if err != nil {
		return fmt.Errorf("get tenant: %w", err)
	}
	tier, err := s.tiers.GetByID(ctx, tenant.TierID)
	if err != nil {
		return fmt.Errorf("get tier: %w", err)
	}

	commission := resp.Amount.Mul(tier.CommissionRate)
	netAmount := resp.Amount.Sub(commission)

	if _, err := s.walletSvc.Credit(ctx, order.TenantID, netAmount, &orderID); err != nil {
		return fmt.Errorf("credit wallet: %w", err)
	}

	if commission.IsPositive() {
		if _, err := s.walletSvc.RecordCommission(ctx, order.TenantID, commission, &orderID); err != nil {
			return fmt.Errorf("record commission: %w", err)
		}
	}

	return nil
}
