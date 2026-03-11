package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

func TestCreateProduct_AutoDefaultVariant(t *testing.T) {
	repo := &mockProductRepo{}
	svc := service.NewProductService(repo)

	p := &models.Product{TenantID: uuid.New(), Name: "T-Shirt", IsAvailable: true}
	_, err := svc.Create(context.Background(), p, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.variantCreated == nil {
		t.Fatal("expected auto default variant to be created")
	}
	if !repo.variantCreated.IsDefault {
		t.Fatal("auto-created variant should have is_default=true")
	}
}

func TestCreateProduct_ExplicitVariants_NoDefault(t *testing.T) {
	repo := &mockProductRepo{}
	svc := service.NewProductService(repo)

	p := &models.Product{TenantID: uuid.New(), Name: "Hoodie", IsAvailable: true}
	variants := []models.ProductVariant{
		{SKU: "HS-M", Price: decimal.NewFromInt(5000)},
		{SKU: "HS-L", Price: decimal.NewFromInt(5000)},
	}
	_, err := svc.Create(context.Background(), p, variants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The last variantCreated will be the second explicit variant, not a default.
	if repo.variantCreated != nil && repo.variantCreated.IsDefault {
		t.Fatal("should not auto-create default variant when explicit variants are provided")
	}
}

func TestDecrementStock_InfiniteStock(t *testing.T) {
	// nil StockQty = infinite; decrement should be a no-op
	repo := &mockProductRepo{variant: &models.ProductVariant{ID: uuid.New(), StockQty: nil}}
	svc := service.NewProductService(repo)

	if err := svc.DecrementStock(context.Background(), uuid.New(), 3); err != nil {
		t.Fatalf("unexpected error for infinite stock: %v", err)
	}
	// variant should not have been updated
	if repo.variant.StockQty != nil {
		t.Fatal("infinite stock variant should not be mutated")
	}
}

func TestDecrementStock_SoldOut(t *testing.T) {
	zero := 0
	repo := &mockProductRepo{variant: &models.ProductVariant{ID: uuid.New(), StockQty: &zero}}
	svc := service.NewProductService(repo)

	err := svc.DecrementStock(context.Background(), uuid.New(), 1)
	if err == nil {
		t.Fatal("expected sold-out error")
	}
}

func TestDecrementStock_Sufficient(t *testing.T) {
	qty := 10
	repo := &mockProductRepo{variant: &models.ProductVariant{ID: uuid.New(), StockQty: &qty}}
	svc := service.NewProductService(repo)

	if err := svc.DecrementStock(context.Background(), uuid.New(), 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.variant.StockQty == nil || *repo.variant.StockQty != 7 {
		t.Fatalf("expected stock_qty=7, got %v", repo.variant.StockQty)
	}
}
