package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	paystack "storefront/backend/internal/adapter/paystack"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

// ── Paystack client mock ──────────────────────────────────────────────────────

type mockPaystackClient struct {
	initResp *paystack.InitializeResponse
	initErr  error
	verResp  *paystack.VerifyResponse
	verErr   error
	initReq  *paystack.InitializeRequest
}

func (m *mockPaystackClient) InitializeTransaction(_ context.Context, req paystack.InitializeRequest) (*paystack.InitializeResponse, error) {
	m.initReq = &req
	return m.initResp, m.initErr
}

func (m *mockPaystackClient) VerifyTransaction(_ context.Context, _ string) (*paystack.VerifyResponse, error) {
	return m.verResp, m.verErr
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newPaymentSvc(ps *mockPaystackClient, orders *mockOrderRepo, wallet *models.Wallet) *service.PaymentService {
	walletSvc := service.NewWalletService(
		&mockWalletRepo{wallet: wallet},
		&mockTxRepo{},
		&mockTenantRepo{},
		testHMACSecret,
	)
	return service.NewPaymentService(ps, orders, &mockProductRepo{}, walletSvc)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestInitiatePayment_ReturnsAuthURL(t *testing.T) {
	orderID := uuid.New()
	order := &models.Order{
		ID:          orderID,
		TotalAmount: decimal.NewFromInt(5000),
		ShippingFee: decimal.NewFromInt(500),
	}
	ps := &mockPaystackClient{
		initResp: &paystack.InitializeResponse{AuthorizationURL: "https://paystack.com/pay/xyz"},
	}
	svc := newPaymentSvc(ps, &mockOrderRepo{order: order}, &models.Wallet{ID: uuid.New()})

	url, err := svc.InitiatePayment(
		context.Background(),
		order,
		"buyer@example.com",
		"SUB_123",
		"https://storefront.test/track/abc123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://paystack.com/pay/xyz" {
		t.Fatalf("expected auth URL, got %s", url)
	}
	if ps.initReq == nil {
		t.Fatal("expected initialize request to be captured")
	}
	if ps.initReq.CallbackURL != "https://storefront.test/track/abc123" {
		t.Fatalf("expected callback URL to be forwarded, got %s", ps.initReq.CallbackURL)
	}
}

func TestInitiatePayment_AdapterError(t *testing.T) {
	order := &models.Order{ID: uuid.New()}
	ps := &mockPaystackClient{initErr: errors.New("paystack down")}
	svc := newPaymentSvc(ps, &mockOrderRepo{order: order}, &models.Wallet{ID: uuid.New()})

	_, err := svc.InitiatePayment(context.Background(), order, "buyer@example.com", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHandleChargeSuccess_CreditsWallet(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	order := &models.Order{
		ID:                orderID,
		TenantID:          tenantID,
		IsDelivery:        true,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPending,
		FulfillmentStatus: models.FulfillmentStatusProcessing,
		TotalAmount:       decimal.NewFromInt(5000),
	}
	wallet := &models.Wallet{ID: uuid.New(), TenantID: tenantID}

	ps := &mockPaystackClient{
		verResp: &paystack.VerifyResponse{
			Status:    "success",
			Amount:    decimal.NewFromInt(5000),
			Reference: orderID.String(),
		},
	}
	txRepo := &mockTxRepo{}
	walletSvc := service.NewWalletService(
		&mockWalletRepo{wallet: wallet},
		txRepo,
		&mockTenantRepo{},
		testHMACSecret,
	)
	paymentSvc := service.NewPaymentService(ps, &mockOrderRepo{order: order}, &mockProductRepo{}, walletSvc)

	err := paymentSvc.HandleChargeSuccess(context.Background(), orderID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txRepo.created == nil {
		t.Fatal("expected wallet credit transaction to be created")
	}
}

func TestHandleChargeSuccess_PickupOrder_CompletesAndCreditsAvailable(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	order := &models.Order{
		ID:                orderID,
		TenantID:          tenantID,
		IsDelivery:        false,
		PaymentStatus:     models.PaymentStatusPending,
		FulfillmentStatus: models.FulfillmentStatusProcessing,
		TotalAmount:       decimal.NewFromInt(5000),
	}
	wallet := &models.Wallet{ID: uuid.New(), TenantID: tenantID}
	ps := &mockPaystackClient{
		verResp: &paystack.VerifyResponse{
			Status:    "success",
			Amount:    decimal.NewFromInt(5000),
			Reference: orderID.String(),
		},
	}
	walletRepo := &mockWalletRepo{wallet: wallet}
	txRepo := &mockTxRepo{}
	walletSvc := service.NewWalletService(walletRepo, txRepo, &mockTenantRepo{}, testHMACSecret)
	paymentSvc := service.NewPaymentService(ps, &mockOrderRepo{order: order}, &mockProductRepo{}, walletSvc)

	err := paymentSvc.HandleChargeSuccess(context.Background(), orderID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.FulfillmentStatus != models.FulfillmentStatusCompleted {
		t.Fatalf("pickup order should complete after payment, got %s", order.FulfillmentStatus)
	}
	if wallet.AvailableBalance.String() != "5000" {
		t.Fatalf("pickup payment should credit available balance, got %s", wallet.AvailableBalance)
	}
	if wallet.PendingBalance.String() != "0" {
		t.Fatalf("pickup payment should not leave pending balance, got %s", wallet.PendingBalance)
	}
}

func TestHandleChargeSuccess_VerificationFails(t *testing.T) {
	ps := &mockPaystackClient{
		verResp: &paystack.VerifyResponse{Status: "failed"},
	}
	svc := newPaymentSvc(ps, &mockOrderRepo{order: &models.Order{ID: uuid.New()}}, &models.Wallet{ID: uuid.New()})

	err := svc.HandleChargeSuccess(context.Background(), uuid.New().String())
	if !errors.Is(err, service.ErrPaymentVerificationFailed) {
		t.Fatalf("expected ErrPaymentVerificationFailed, got %v", err)
	}
}

func TestHandleChargeSuccess_InvalidReference(t *testing.T) {
	ps := &mockPaystackClient{
		verResp: &paystack.VerifyResponse{Status: "success"},
	}
	svc := newPaymentSvc(ps, &mockOrderRepo{}, &models.Wallet{ID: uuid.New()})

	err := svc.HandleChargeSuccess(context.Background(), "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid reference, got nil")
	}
}

func TestHandleChargeSuccess_AmountMismatch(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	order := &models.Order{
		ID:                orderID,
		TenantID:          tenantID,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPending,
		FulfillmentStatus: models.FulfillmentStatusProcessing,
		TotalAmount:       decimal.NewFromInt(5000),
		ShippingFee:       decimal.NewFromInt(500),
	}
	ps := &mockPaystackClient{
		verResp: &paystack.VerifyResponse{
			Status:    "success",
			Amount:    decimal.NewFromInt(3000), // mismatch: order expects 5500
			Reference: orderID.String(),
		},
	}
	svc := newPaymentSvc(ps, &mockOrderRepo{order: order}, &models.Wallet{ID: uuid.New(), TenantID: tenantID})

	err := svc.HandleChargeSuccess(context.Background(), orderID.String())
	if !errors.Is(err, service.ErrPaymentAmountMismatch) {
		t.Fatalf("expected ErrPaymentAmountMismatch, got %v", err)
	}
}

func TestHandleChargeFailed_CancelsAndRestocks(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	variantID := uuid.New()
	orderRepo := &mockOrderRepo{
		order: &models.Order{
			ID:                orderID,
			TenantID:          tenantID,
			PaymentMethod:     models.PaymentMethodOnline,
			PaymentStatus:     models.PaymentStatusPending,
			FulfillmentStatus: models.FulfillmentStatusProcessing,
		},
		items: []models.OrderItem{{VariantID: variantID, Quantity: 1}},
	}
	productRepo := &mockProductRepo{}
	walletSvc := service.NewWalletService(
		&mockWalletRepo{wallet: &models.Wallet{ID: uuid.New(), TenantID: tenantID}},
		&mockTxRepo{},
		&mockTenantRepo{},
		testHMACSecret,
	)
	svc := service.NewPaymentService(&mockPaystackClient{}, orderRepo, productRepo, walletSvc)

	err := svc.HandleChargeFailed(context.Background(), orderID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orderRepo.paymentStatus != models.PaymentStatusFailed {
		t.Fatalf("expected failed payment status, got %s", orderRepo.paymentStatus)
	}
	if orderRepo.fulfillmentStatus != models.FulfillmentStatusCancelled {
		t.Fatalf("expected cancelled fulfillment status, got %s", orderRepo.fulfillmentStatus)
	}
	if productRepo.restocked[variantID] != 1 {
		t.Fatalf("expected restock of 1, got %d", productRepo.restocked[variantID])
	}
}
