package service_test

import (
	"context"
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
