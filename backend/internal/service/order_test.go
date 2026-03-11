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
		CustomerName:    "Ade",
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

	order := &models.Order{TenantID: uuid.New(), IsDelivery: false, CustomerName: "Bola"}
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

	order := &models.Order{TenantID: uuid.New(), IsDelivery: false, CustomerName: "Chidi"}
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

	order := &models.Order{TenantID: uuid.New(), CustomerName: "Chidi"}
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

	order := &models.Order{TenantID: uuid.New(), CustomerName: "Ade"}
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

	order := &models.Order{TenantID: uuid.New(), CustomerName: "Bola"}
	out, err := svc.Create(context.Background(), order, []models.OrderItem{{VariantID: variantID, Quantity: 3}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := decimal.NewFromInt(3000) // 1000 * 3
	if !out.TotalAmount.Equal(expected) {
		t.Fatalf("total_amount: want %s, got %s", expected, out.TotalAmount)
	}
}
