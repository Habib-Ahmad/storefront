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

func TestCreateProduct_BlankSingleVariantSKU_UsesDefaultLabel(t *testing.T) {
	repo := &mockProductRepo{}
	svc := service.NewProductService(repo)

	p := &models.Product{TenantID: uuid.New(), Name: "Hoodie", IsAvailable: true}
	variants := []models.ProductVariant{{SKU: "", Price: decimal.NewFromInt(5000)}}
	_, err := svc.Create(context.Background(), p, variants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.variantCreated == nil {
		t.Fatal("expected variant to be created")
	}
	if repo.variantCreated.SKU != "Default" {
		t.Fatalf("expected default SKU label, got %q", repo.variantCreated.SKU)
	}
	if !repo.variantCreated.IsDefault {
		t.Fatal("expected blank single variant to become the default variant")
	}
	if !repo.variantCreated.Price.Equal(decimal.NewFromInt(5000)) {
		t.Fatalf("expected price to be preserved, got %s", repo.variantCreated.Price)
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

func TestDecrementStock_InsufficientQty(t *testing.T) {
	// stock=2, order qty=5: should block even though not sold out
	qty := 2
	repo := &mockProductRepo{variant: &models.ProductVariant{ID: uuid.New(), StockQty: &qty}}
	svc := service.NewProductService(repo)

	err := svc.DecrementStock(context.Background(), uuid.New(), 5)
	if err == nil {
		t.Fatal("expected error when requested qty exceeds stock")
	}
	if !errors.Is(err, service.ErrSoldOut) {
		t.Fatalf("expected ErrSoldOut, got: %v", err)
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

func TestGetByID_Found(t *testing.T) {
	productID := uuid.New()
	tenantID := uuid.New()
	repo := &mockProductRepo{product: &models.Product{ID: productID, TenantID: tenantID, Name: "T-Shirt"}}
	svc := service.NewProductService(repo)

	p, variants, err := svc.GetByID(context.Background(), tenantID, productID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != productID {
		t.Fatalf("expected product %s, got %s", productID, p.ID)
	}
	_ = variants // just ensure no error
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockProductRepo{err: errors.New("no rows")}
	svc := service.NewProductService(repo)

	_, _, err := svc.GetByID(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, service.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got: %v", err)
	}
}

func TestUpdate_OK(t *testing.T) {
	productID := uuid.New()
	tenantID := uuid.New()
	repo := &mockProductRepo{product: &models.Product{ID: productID, TenantID: tenantID, Name: "Old"}}
	svc := service.NewProductService(repo)

	p := &models.Product{ID: productID, TenantID: tenantID, Name: "New", IsAvailable: true}
	if err := svc.Update(context.Background(), p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := &mockProductRepo{err: errors.New("no rows")}
	svc := service.NewProductService(repo)

	err := svc.Update(context.Background(), &models.Product{ID: uuid.New(), TenantID: uuid.New(), Name: "X"})
	if !errors.Is(err, service.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got: %v", err)
	}
}

func TestSoftDelete_CascadesVariants(t *testing.T) {
	repo := &mockProductRepo{product: &models.Product{ID: uuid.New()}}
	svc := service.NewProductService(repo)

	if err := svc.SoftDelete(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateVariant_OK(t *testing.T) {
	productID := uuid.New()
	tenantID := uuid.New()
	repo := &mockProductRepo{product: &models.Product{ID: productID, TenantID: tenantID}}
	svc := service.NewProductService(repo)

	v := &models.ProductVariant{ProductID: productID, SKU: "M-RED", Price: decimal.NewFromInt(5000)}
	if err := svc.CreateVariant(context.Background(), tenantID, v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID == uuid.Nil {
		t.Fatal("variant ID should be assigned")
	}
}

func TestCreateVariant_ProductNotFound(t *testing.T) {
	repo := &mockProductRepo{err: errors.New("no rows")}
	svc := service.NewProductService(repo)

	v := &models.ProductVariant{ProductID: uuid.New(), SKU: "X", Price: decimal.NewFromInt(100)}
	err := svc.CreateVariant(context.Background(), uuid.New(), v)
	if !errors.Is(err, service.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got: %v", err)
	}
}

func TestUpdateVariant_OK(t *testing.T) {
	productID := uuid.New()
	tenantID := uuid.New()
	variantID := uuid.New()
	repo := &mockProductRepo{
		product: &models.Product{ID: productID, TenantID: tenantID},
		variant: &models.ProductVariant{ID: variantID, ProductID: productID, SKU: "OLD"},
	}
	svc := service.NewProductService(repo)

	v := &models.ProductVariant{ID: variantID, SKU: "NEW", Price: decimal.NewFromInt(6000)}
	if err := svc.UpdateVariant(context.Background(), tenantID, v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateVariant_NotFound(t *testing.T) {
	repo := &mockProductRepo{err: errors.New("no rows")}
	svc := service.NewProductService(repo)

	v := &models.ProductVariant{ID: uuid.New(), SKU: "X", Price: decimal.NewFromInt(100)}
	err := svc.UpdateVariant(context.Background(), uuid.New(), v)
	if !errors.Is(err, service.ErrVariantNotFound) {
		t.Fatalf("expected ErrVariantNotFound, got: %v", err)
	}
}

func TestDeleteVariant_OK(t *testing.T) {
	productID := uuid.New()
	tenantID := uuid.New()
	variantID := uuid.New()
	repo := &mockProductRepo{
		product: &models.Product{ID: productID, TenantID: tenantID},
		variant: &models.ProductVariant{ID: variantID, ProductID: productID},
	}
	svc := service.NewProductService(repo)

	if err := svc.DeleteVariant(context.Background(), tenantID, variantID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteVariant_NotFound(t *testing.T) {
	repo := &mockProductRepo{err: errors.New("no rows")}
	svc := service.NewProductService(repo)

	err := svc.DeleteVariant(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, service.ErrVariantNotFound) {
		t.Fatalf("expected ErrVariantNotFound, got: %v", err)
	}
}
