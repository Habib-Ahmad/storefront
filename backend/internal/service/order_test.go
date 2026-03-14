package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

func phone(s string) *string { return &s }
func addr(s string) *string  { return &s }
func name(s string) *string  { return &s }

func TestCreateOrder_DeliveryMissingPhone(t *testing.T) {
	svc := service.NewOrderService(&mockOrderRepo{}, &mockProductRepo{})
	order := &models.Order{IsDelivery: true, ShippingAddress: addr("123 Lagos")}
	_, err := svc.Create(context.Background(), order, nil)
	if err == nil {
		t.Fatal("expected validation error for missing phone")
	}
}

func TestCreateOrder_DeliveryMissingAddress(t *testing.T) {
	svc := service.NewOrderService(&mockOrderRepo{}, &mockProductRepo{})
	order := &models.Order{IsDelivery: true, CustomerPhone: phone("08012345678")}
	_, err := svc.Create(context.Background(), order, nil)
	if err == nil {
		t.Fatal("expected validation error for missing address")
	}
}

func TestCreateOrder_DeliveryValid(t *testing.T) {
	variantID := uuid.New()
	price := decimal.NewFromInt(2500)
	repo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, Price: price, StockQty: nil}}

	svc := service.NewOrderService(&mockOrderRepo{}, repo)
	order := &models.Order{
		TenantID:        uuid.New(),
		IsDelivery:      true,
		CustomerPhone:   phone("08012345678"),
		ShippingAddress: addr("123 Lagos"),
		CustomerName:    name("Ade"),
	}
	items := []models.OrderItem{{VariantID: variantID, Quantity: 1}}
	out, err := svc.Create(context.Background(), order, items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PaymentStatus != models.PaymentStatusPending {
		t.Fatal("payment_status should start as pending")
	}
	if out.FulfillmentStatus != models.FulfillmentStatusProcessing {
		t.Fatal("fulfillment_status should start as processing")
	}
}

func TestCreateOrder_PickupNoValidation(t *testing.T) {
	variantID := uuid.New()
	repo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(1000), StockQty: nil}}
	svc := service.NewOrderService(&mockOrderRepo{}, repo)

	order := &models.Order{TenantID: uuid.New(), IsDelivery: false, CustomerName: name("Bola")}
	_, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 1}})
	if err != nil {
		t.Fatalf("pickup order should not require phone/address: %v", err)
	}
}

func TestCreateOrder_SoldOutVariant(t *testing.T) {
	variantID := uuid.New()
	zero := 0
	repo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, StockQty: &zero}}
	svc := service.NewOrderService(&mockOrderRepo{}, repo)

	order := &models.Order{TenantID: uuid.New(), IsDelivery: false, CustomerName: name("Chidi")}
	_, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 1}})
	if err == nil {
		t.Fatal("expected sold-out error")
	}
}

func TestCreateOrder_InsufficientStock(t *testing.T) {
	// stock=3, order qty=10: should reject even though not fully sold out
	variantID := uuid.New()
	qty := 3
	repo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(500), StockQty: &qty}}
	svc := service.NewOrderService(&mockOrderRepo{}, repo)

	order := &models.Order{TenantID: uuid.New(), CustomerName: name("Chidi")}
	_, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 10}})
	if err == nil {
		t.Fatal("expected error: requested quantity exceeds available stock")
	}
	if !errors.Is(err, service.ErrSoldOut) {
		t.Fatalf("expected ErrSoldOut, got: %v", err)
	}
}

func TestCreateOrder_PriceSnapshotOnItem(t *testing.T) {
	// Spec §ERD: price_at_sale is an immutable snapshot of the variant price at time of sale.
	variantID := uuid.New()
	price := decimal.NewFromInt(2500)
	orderRepo := &mockOrderRepo{}
	productRepo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, Price: price, StockQty: nil}}
	svc := service.NewOrderService(orderRepo, productRepo)

	order := &models.Order{TenantID: uuid.New(), CustomerName: name("Ade")}
	_, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orderRepo.items) == 0 {
		t.Fatal("no items captured")
	}
	if !orderRepo.items[0].PriceAtSale.Equal(price) {
		t.Fatalf("price_at_sale: want %s, got %s", price, orderRepo.items[0].PriceAtSale)
	}
}

func TestCreateOrder_TotalAmount(t *testing.T) {
	// total_amount = sum of (price_at_sale * quantity) across all items
	variantID := uuid.New()
	price := decimal.NewFromInt(1000)
	orderRepo := &mockOrderRepo{}
	productRepo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, Price: price, StockQty: nil}}
	svc := service.NewOrderService(orderRepo, productRepo)

	order := &models.Order{TenantID: uuid.New(), CustomerName: name("Bola")}
	out, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 3}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := decimal.NewFromInt(3000) // 1000 * 3
	if !out.TotalAmount.Equal(expected) {
		t.Fatalf("total_amount: want %s, got %s", expected, out.TotalAmount)
	}
}

func TestCreateOrder_CashSale_PaidImmediately(t *testing.T) {
	variantID := uuid.New()
	repo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(500), StockQty: nil}}
	svc := service.NewOrderService(&mockOrderRepo{}, repo)

	order := &models.Order{TenantID: uuid.New(), PaymentMethod: models.PaymentMethodCash}
	out, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 2}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PaymentStatus != models.PaymentStatusPaid {
		t.Fatalf("cash sale should be paid immediately, got %s", out.PaymentStatus)
	}
}

func TestCreateOrder_TransferSale_PaidImmediately(t *testing.T) {
	variantID := uuid.New()
	repo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(1000), StockQty: nil}}
	svc := service.NewOrderService(&mockOrderRepo{}, repo)

	order := &models.Order{TenantID: uuid.New(), PaymentMethod: models.PaymentMethodTransfer}
	out, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PaymentStatus != models.PaymentStatusPaid {
		t.Fatalf("transfer sale should be paid immediately, got %s", out.PaymentStatus)
	}
}

func TestCreateOrder_CashSale_CreditsWallet(t *testing.T) {
	variantID := uuid.New()
	tenantID := uuid.New()
	walletID := uuid.New()

	productRepo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(2000), StockQty: nil}}
	orderRepo := &mockOrderRepo{}
	txRepo := &mockTxRepo{}
	walletRepo := &mockWalletRepo{wallet: &models.Wallet{ID: walletID, TenantID: tenantID}}
	tenantRepo := &mockTenantRepo{tenant: &models.Tenant{ID: tenantID, TierID: uuid.New()}}
	tierRepo := &mockTierRepo{tier: &models.Tier{CommissionRate: decimal.NewFromFloat(0.05)}}

	walletSvc := service.NewWalletService(walletRepo, txRepo, tenantRepo, testHMACSecret)
	walletSvc.SetTierRepo(tierRepo)

	svc := service.NewOrderService(orderRepo, productRepo)
	svc.SetWalletService(walletSvc)
	svc.SetTenantRepo(tenantRepo)
	svc.SetTierRepo(tierRepo)

	order := &models.Order{TenantID: tenantID, PaymentMethod: models.PaymentMethodCash}
	_, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if txRepo.created == nil {
		t.Fatal("expected wallet transaction to be created")
	}
	// 2000 * 0.05 = 100 commission; net credit = 1900
	// But commission is recorded as a separate entry, so the credit is 1900
	// and then a commission debit of -100.
	// allCreated should have 2 entries.
	if len(txRepo.allCreated) != 2 {
		t.Fatalf("expected 2 transactions (credit + commission), got %d", len(txRepo.allCreated))
	}
	credit := txRepo.allCreated[0]
	if !credit.Amount.Equal(decimal.NewFromInt(1900)) {
		t.Fatalf("net credit: want 1900, got %s", credit.Amount)
	}
	comm := txRepo.allCreated[1]
	if !comm.Amount.Equal(decimal.NewFromInt(-100)) {
		t.Fatalf("commission: want -100, got %s", comm.Amount)
	}
}

func TestCreateOrder_QuickSale_NoItems(t *testing.T) {
	orderRepo := &mockOrderRepo{}
	svc := service.NewOrderService(orderRepo, &mockProductRepo{})

	total := decimal.NewFromInt(5000)
	order := &models.Order{TenantID: uuid.New(), PaymentMethod: models.PaymentMethodCash, TotalAmount: total}
	out, err := svc.Create(context.Background(), order, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.TotalAmount.Equal(total) {
		t.Fatalf("total_amount: want %s, got %s", total, out.TotalAmount)
	}
	if out.PaymentStatus != models.PaymentStatusPaid {
		t.Fatalf("quick cash sale should be paid, got %s", out.PaymentStatus)
	}
}

// ── 6c: Order lifecycle tests ─────────────────────────────────────────────────

func TestCancelOrder_PaidOrder_RefundsAndRestocks(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	variantID := uuid.New()

	orderRepo := &mockOrderRepo{
		order: &models.Order{
			ID:                orderID,
			TenantID:          tenantID,
			FulfillmentStatus: models.FulfillmentStatusProcessing,
			PaymentStatus:     models.PaymentStatusPaid,
			TotalAmount:       decimal.NewFromInt(3000),
		},
		items: []models.OrderItem{
			{VariantID: variantID, Quantity: 2},
		},
	}
	productRepo := &mockProductRepo{}
	txRepo := &mockTxRepo{}
	walletRepo := &mockWalletRepo{wallet: &models.Wallet{ID: uuid.New(), TenantID: tenantID}}
	walletSvc := service.NewWalletService(walletRepo, txRepo, &mockTenantRepo{}, testHMACSecret)

	svc := service.NewOrderService(orderRepo, productRepo)
	svc.SetWalletService(walletSvc)

	err := svc.Cancel(context.Background(), tenantID, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orderRepo.fulfillmentStatus != models.FulfillmentStatusCancelled {
		t.Fatalf("expected cancelled, got %s", orderRepo.fulfillmentStatus)
	}
	if orderRepo.paymentStatus != models.PaymentStatusRefunded {
		t.Fatalf("expected refunded, got %s", orderRepo.paymentStatus)
	}
	if productRepo.restocked[variantID] != 2 {
		t.Fatalf("expected restock of 2, got %d", productRepo.restocked[variantID])
	}
	if txRepo.created == nil {
		t.Fatal("expected wallet refund transaction")
	}
}

func TestCancelOrder_PendingPayment_NoRefund(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	variantID := uuid.New()

	orderRepo := &mockOrderRepo{
		order: &models.Order{
			ID:                orderID,
			TenantID:          tenantID,
			FulfillmentStatus: models.FulfillmentStatusProcessing,
			PaymentStatus:     models.PaymentStatusPending,
			TotalAmount:       decimal.NewFromInt(2000),
		},
		items: []models.OrderItem{
			{VariantID: variantID, Quantity: 1},
		},
	}
	productRepo := &mockProductRepo{}
	svc := service.NewOrderService(orderRepo, productRepo)

	err := svc.Cancel(context.Background(), tenantID, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orderRepo.fulfillmentStatus != models.FulfillmentStatusCancelled {
		t.Fatalf("expected cancelled, got %s", orderRepo.fulfillmentStatus)
	}
	if productRepo.restocked[variantID] != 1 {
		t.Fatalf("expected restock of 1, got %d", productRepo.restocked[variantID])
	}
	// payment status should NOT be updated (no refund needed for unpaid order)
	if orderRepo.paymentStatus == models.PaymentStatusRefunded {
		t.Fatal("pending order should not be refunded")
	}
}

func TestCancelOrder_ShippedOrder_Fails(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()

	orderRepo := &mockOrderRepo{
		order: &models.Order{
			ID:                orderID,
			TenantID:          tenantID,
			FulfillmentStatus: models.FulfillmentStatusShipped,
			PaymentStatus:     models.PaymentStatusPaid,
		},
	}
	svc := service.NewOrderService(orderRepo, &mockProductRepo{})

	err := svc.Cancel(context.Background(), tenantID, orderID)
	if !errors.Is(err, service.ErrOrderNotCancellable) {
		t.Fatalf("expected ErrOrderNotCancellable, got %v", err)
	}
}

// ── 6d: Offline payment path tests ───────────────────────────────────────────

func TestCreateOrder_TransferSale_CreditsWallet(t *testing.T) {
	variantID := uuid.New()
	tenantID := uuid.New()
	walletID := uuid.New()

	productRepo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(1500), StockQty: nil}}
	orderRepo := &mockOrderRepo{}
	txRepo := &mockTxRepo{}
	walletRepo := &mockWalletRepo{wallet: &models.Wallet{ID: walletID, TenantID: tenantID}}
	tenantRepo := &mockTenantRepo{tenant: &models.Tenant{ID: tenantID, TierID: uuid.New()}}
	tierRepo := &mockTierRepo{tier: &models.Tier{CommissionRate: decimal.NewFromFloat(0.10)}}

	walletSvc := service.NewWalletService(walletRepo, txRepo, tenantRepo, testHMACSecret)
	walletSvc.SetTierRepo(tierRepo)

	svc := service.NewOrderService(orderRepo, productRepo)
	svc.SetWalletService(walletSvc)
	svc.SetTenantRepo(tenantRepo)
	svc.SetTierRepo(tierRepo)

	order := &models.Order{TenantID: tenantID, PaymentMethod: models.PaymentMethodTransfer}
	_, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 2}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(txRepo.allCreated) != 2 {
		t.Fatalf("expected 2 transactions (credit + commission), got %d", len(txRepo.allCreated))
	}
	// 1500 * 2 = 3000; commission = 300 (10%); net credit = 2700
	credit := txRepo.allCreated[0]
	if !credit.Amount.Equal(decimal.NewFromInt(2700)) {
		t.Fatalf("net credit: want 2700, got %s", credit.Amount)
	}
	comm := txRepo.allCreated[1]
	if !comm.Amount.Equal(decimal.NewFromInt(-300)) {
		t.Fatalf("commission: want -300, got %s", comm.Amount)
	}
}

func TestCreateOrder_OnlineSale_NoWalletCredit(t *testing.T) {
	variantID := uuid.New()
	tenantID := uuid.New()

	productRepo := &mockProductRepo{variant: &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(1000), StockQty: nil}}
	orderRepo := &mockOrderRepo{}
	txRepo := &mockTxRepo{}
	walletRepo := &mockWalletRepo{wallet: &models.Wallet{ID: uuid.New(), TenantID: tenantID}}
	walletSvc := service.NewWalletService(walletRepo, txRepo, &mockTenantRepo{}, testHMACSecret)

	svc := service.NewOrderService(orderRepo, productRepo)
	svc.SetWalletService(walletSvc)

	order := &models.Order{TenantID: tenantID, PaymentMethod: models.PaymentMethodOnline}
	out, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PaymentStatus != models.PaymentStatusPending {
		t.Fatalf("online sale should be pending, got %s", out.PaymentStatus)
	}
	if txRepo.created != nil {
		t.Fatal("online sale should not create wallet transaction at order time")
	}
}
