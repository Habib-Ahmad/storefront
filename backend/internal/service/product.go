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

func (s *ProductService) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.products.SoftDelete(ctx, tenantID, id)
}

func (s *ProductService) List(ctx context.Context, tenantID uuid.UUID) ([]models.Product, error) {
	return s.products.ListByTenant(ctx, tenantID)
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
