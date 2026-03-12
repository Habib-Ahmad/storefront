package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrVariantNotFound = errors.New("variant not found")
	ErrSoldOut         = errors.New("variant is sold out")
)

type ProductService struct {
	products repository.ProductRepository
}

func NewProductService(products repository.ProductRepository) *ProductService {
	return &ProductService{products: products}
}

// Create creates a product and auto-creates a default variant if none are provided.
func (s *ProductService) Create(ctx context.Context, p *models.Product, variants []models.ProductVariant) (*models.Product, error) {
	if err := s.products.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	if len(variants) == 0 {
		// Single-variant rule: auto-create a hidden default variant.
		v := defaultVariant(p.ID)
		if err := s.products.CreateVariant(ctx, v); err != nil {
			return nil, fmt.Errorf("create default variant: %w", err)
		}
		return p, nil
	}

	for i := range variants {
		variants[i].ProductID = p.ID
		if err := s.products.CreateVariant(ctx, &variants[i]); err != nil {
			return nil, fmt.Errorf("create variant: %w", err)
		}
	}
	return p, nil
}

// DecrementStock reduces stock by qty. Blocks on sold-out (stock_qty == 0).
// Nil stock_qty (infinite) is a no-op.
func (s *ProductService) DecrementStock(ctx context.Context, variantID uuid.UUID, qty int) error {
	v, err := s.products.GetVariantByID(ctx, variantID)
	if err != nil {
		return ErrVariantNotFound
	}

	// nil = infinite stock, nothing to decrement
	if v.StockQty == nil {
		return nil
	}
	if *v.StockQty == 0 {
		return ErrSoldOut
	}
	if *v.StockQty < qty {
		return ErrSoldOut
	}

	newQty := *v.StockQty - qty
	v.StockQty = &newQty
	return s.products.UpdateVariant(ctx, v)
}

func (s *ProductService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, []models.ProductVariant, error) {
	p, err := s.products.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, nil, ErrProductNotFound
	}
	variants, err := s.products.ListVariants(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("list variants: %w", err)
	}
	return p, variants, nil
}

func (s *ProductService) Update(ctx context.Context, p *models.Product) error {
	existing, err := s.products.GetByID(ctx, p.TenantID, p.ID)
	if err != nil {
		return ErrProductNotFound
	}
	existing.Name = p.Name
	existing.Description = p.Description
	existing.Category = p.Category
	existing.IsAvailable = p.IsAvailable
	return s.products.Update(ctx, existing)
}

func (s *ProductService) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	variants, err := s.products.ListVariants(ctx, id)
	if err != nil {
		return fmt.Errorf("list variants for cascade: %w", err)
	}
	for _, v := range variants {
		if err := s.products.SoftDeleteVariant(ctx, v.ID); err != nil {
			return fmt.Errorf("cascade delete variant %s: %w", v.ID, err)
		}
	}
	return s.products.SoftDelete(ctx, tenantID, id)
}

func (s *ProductService) List(ctx context.Context, tenantID uuid.UUID) ([]models.Product, error) {
	return s.products.ListByTenant(ctx, tenantID)
}

func (s *ProductService) CreateVariant(ctx context.Context, tenantID uuid.UUID, v *models.ProductVariant) error {
	if _, err := s.products.GetByID(ctx, tenantID, v.ProductID); err != nil {
		return ErrProductNotFound
	}
	return s.products.CreateVariant(ctx, v)
}

func (s *ProductService) ListVariants(ctx context.Context, tenantID, productID uuid.UUID) ([]models.ProductVariant, error) {
	if _, err := s.products.GetByID(ctx, tenantID, productID); err != nil {
		return nil, ErrProductNotFound
	}
	return s.products.ListVariants(ctx, productID)
}

func (s *ProductService) UpdateVariant(ctx context.Context, v *models.ProductVariant) error {
	existing, err := s.products.GetVariantByID(ctx, v.ID)
	if err != nil {
		return ErrVariantNotFound
	}
	existing.SKU = v.SKU
	existing.Attributes = v.Attributes
	existing.Price = v.Price
	existing.StockQty = v.StockQty
	return s.products.UpdateVariant(ctx, existing)
}

func (s *ProductService) DeleteVariant(ctx context.Context, id uuid.UUID) error {
	if _, err := s.products.GetVariantByID(ctx, id); err != nil {
		return ErrVariantNotFound
	}
	return s.products.SoftDeleteVariant(ctx, id)
}

func defaultVariant(productID uuid.UUID) *models.ProductVariant {
	attrs, _ := json.Marshal(map[string]string{})
	return &models.ProductVariant{
		ProductID:  productID,
		SKU:        "DEFAULT-" + productID.String()[:8],
		Attributes: json.RawMessage(attrs),
		Price:      decimal.Zero,
		IsDefault:  true,
	}
}
