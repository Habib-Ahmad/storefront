package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/apperr"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var (
	ErrProductNotFound    = apperr.NotFound("product not found")
	ErrVariantNotFound    = apperr.NotFound("variant not found")
	ErrImageNotFound      = apperr.NotFound("image not found")
	ErrSoldOut            = apperr.Conflict("variant is sold out")
	ErrDuplicateSKU       = apperr.Conflict("a variant with this SKU already exists for this product")
	ErrDuplicateSortOrder = apperr.Conflict("an image with this sort order already exists for this product")
)

type ProductService struct {
	products repository.ProductRepository
}

func NewProductService(products repository.ProductRepository) *ProductService {
	return &ProductService{products: products}
}

var emptyJSON = json.RawMessage([]byte(`{}`))

// normalizeVariant defaults nil Attributes to {} so the NOT NULL column is satisfied.
func normalizeVariant(v *models.ProductVariant) {
	if len(v.Attributes) == 0 {
		v.Attributes = emptyJSON
	}
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
		normalizeVariant(&variants[i])
		if err := s.products.CreateVariant(ctx, &variants[i]); err != nil {
			if apperr.IsUniqueViolation(err) {
				return nil, ErrDuplicateSKU
			}
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

func (s *ProductService) GetImagesByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]models.ProductImage, error) {
	if _, err := s.products.GetByID(ctx, tenantID, productID); err != nil {
		return nil, ErrProductNotFound
	}
	return s.products.ListImagesByProduct(ctx, productID)
}

func (s *ProductService) GetImage(ctx context.Context, tenantID, productID, imageID uuid.UUID) (*models.ProductImage, error) {
	if _, err := s.products.GetByID(ctx, tenantID, productID); err != nil {
		return nil, ErrProductNotFound
	}
	images, err := s.products.ListImagesByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	for _, img := range images {
		if img.ID == imageID {
			return &img, nil
		}
	}
	return nil, ErrImageNotFound
}

func (s *ProductService) AddImage(ctx context.Context, tenantID uuid.UUID, img *models.ProductImage) error {
	if _, err := s.products.GetByID(ctx, tenantID, img.ProductID); err != nil {
		return ErrProductNotFound
	}
	if err := s.products.AddImage(ctx, img); err != nil {
		if apperr.IsUniqueViolation(err) {
			return ErrDuplicateSortOrder
		}
		return err
	}
	return nil
}

func (s *ProductService) DeleteImage(ctx context.Context, tenantID, productID, imageID uuid.UUID) error {
	if _, err := s.products.GetByID(ctx, tenantID, productID); err != nil {
		return ErrProductNotFound
	}
	images, err := s.products.ListImagesByProduct(ctx, productID)
	if err != nil {
		return err
	}
	found := false
	for _, img := range images {
		if img.ID == imageID {
			found = true
			break
		}
	}
	if !found {
		return ErrImageNotFound
	}
	return s.products.DeleteImage(ctx, imageID)
}

func (s *ProductService) UpdateImage(ctx context.Context, tenantID uuid.UUID, img *models.ProductImage) error {
	if _, err := s.products.GetByID(ctx, tenantID, img.ProductID); err != nil {
		return ErrProductNotFound
	}
	if err := s.products.UpdateImage(ctx, img); err != nil {
		if apperr.IsUniqueViolation(err) {
			return ErrDuplicateSortOrder
		}
		return err
	}
	return nil
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

func (s *ProductService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]models.Product, error) {
	return s.products.ListByTenant(ctx, tenantID, limit, offset)
}

func (s *ProductService) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return s.products.CountByTenant(ctx, tenantID)
}

func (s *ProductService) CreateVariant(ctx context.Context, tenantID uuid.UUID, v *models.ProductVariant) error {
	if _, err := s.products.GetByID(ctx, tenantID, v.ProductID); err != nil {
		return ErrProductNotFound
	}
	normalizeVariant(v)
	if err := s.products.CreateVariant(ctx, v); err != nil {
		if apperr.IsUniqueViolation(err) {
			return ErrDuplicateSKU
		}
		return err
	}
	return nil
}

func (s *ProductService) ListVariants(ctx context.Context, tenantID, productID uuid.UUID) ([]models.ProductVariant, error) {
	if _, err := s.products.GetByID(ctx, tenantID, productID); err != nil {
		return nil, ErrProductNotFound
	}
	return s.products.ListVariants(ctx, productID)
}

func (s *ProductService) UpdateVariant(ctx context.Context, tenantID uuid.UUID, v *models.ProductVariant) error {
	existing, err := s.products.GetVariantByID(ctx, v.ID)
	if err != nil {
		return ErrVariantNotFound
	}
	if _, err := s.products.GetByID(ctx, tenantID, existing.ProductID); err != nil {
		return ErrVariantNotFound
	}
	existing.SKU = v.SKU
	existing.Attributes = v.Attributes
	existing.Price = v.Price
	existing.StockQty = v.StockQty
	normalizeVariant(existing)
	return s.products.UpdateVariant(ctx, existing)
}

func (s *ProductService) DeleteVariant(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	existing, err := s.products.GetVariantByID(ctx, id)
	if err != nil {
		return ErrVariantNotFound
	}
	if _, err := s.products.GetByID(ctx, tenantID, existing.ProductID); err != nil {
		return ErrVariantNotFound
	}
	return s.products.SoftDeleteVariant(ctx, id)
}

func defaultVariant(productID uuid.UUID) *models.ProductVariant {
	attrs, _ := json.Marshal(map[string]string{})
	return &models.ProductVariant{
		ProductID:  productID,
		SKU:        "Default",
		Attributes: json.RawMessage(attrs),
		Price:      decimal.Zero,
		IsDefault:  true,
	}
}
