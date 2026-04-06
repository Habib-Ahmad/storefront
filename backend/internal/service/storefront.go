package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/apperr"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var ErrStorefrontNotFound = apperr.NotFound("storefront not found")

type StorefrontService struct {
	tenants  repository.TenantRepository
	products repository.ProductRepository
}

func NewStorefrontService(tenants repository.TenantRepository, products repository.ProductRepository) *StorefrontService {
	return &StorefrontService{tenants: tenants, products: products}
}

func (s *StorefrontService) GetPublicBySlug(ctx context.Context, slug string) (*models.PublicStorefrontCatalog, error) {
	tenant, storefront, err := s.getPublishedStorefront(ctx, slug)
	if err != nil {
		return nil, err
	}

	products, err := s.products.ListPublicByTenant(ctx, tenant.ID)
	if err != nil {
		return nil, fmt.Errorf("list public storefront products: %w", err)
	}
	if products == nil {
		products = []models.PublicStorefrontProduct{}
	}

	return &models.PublicStorefrontCatalog{
		Storefront: storefront,
		Products:   products,
	}, nil
}

func (s *StorefrontService) GetPublicProductBySlug(ctx context.Context, slug string, productID uuid.UUID) (*models.PublicStorefrontProductDetail, error) {
	tenant, storefront, err := s.getPublishedStorefront(ctx, slug)
	if err != nil {
		return nil, err
	}

	product, err := s.products.GetByID(ctx, tenant.ID, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStorefrontNotFound
		}
		return nil, fmt.Errorf("get public product: %w", err)
	}
	if !product.IsAvailable {
		return nil, ErrStorefrontNotFound
	}

	variants, err := s.products.ListVariants(ctx, product.ID)
	if err != nil {
		return nil, fmt.Errorf("list public product variants: %w", err)
	}
	images, err := s.products.ListImagesByProduct(ctx, product.ID)
	if err != nil {
		return nil, fmt.Errorf("list public product images: %w", err)
	}

	publicVariants := make([]models.PublicStorefrontProductVariant, 0, len(variants))
	inStock := false
	startingPrice := decimal.Zero
	for index, variant := range variants {
		variantInStock := variant.StockQty == nil || *variant.StockQty > 0
		if variantInStock {
			inStock = true
		}
		if index == 0 || variant.Price.LessThan(startingPrice) {
			startingPrice = variant.Price
		}
		publicVariants = append(publicVariants, models.PublicStorefrontProductVariant{
			ID:         variant.ID,
			Attributes: variant.Attributes,
			Price:      variant.Price,
			InStock:    variantInStock,
			IsDefault:  variant.IsDefault,
		})
	}

	publicImages := make([]models.PublicStorefrontProductImage, 0, len(images))
	var primaryImageURL *string
	for index, image := range images {
		if index == 0 || image.IsPrimary {
			primaryImageURL = &image.URL
		}
		publicImages = append(publicImages, models.PublicStorefrontProductImage{
			ID:        image.ID,
			URL:       image.URL,
			SortOrder: image.SortOrder,
			IsPrimary: image.IsPrimary,
		})
	}

	return &models.PublicStorefrontProductDetail{
		Storefront: storefront,
		Product: models.PublicStorefrontProduct{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Category:    product.Category,
			ImageURL:    primaryImageURL,
			Price:       startingPrice,
			InStock:     inStock,
		},
		Variants: publicVariants,
		Images:   publicImages,
	}, nil
}

func (s *StorefrontService) getPublishedStorefront(ctx context.Context, slug string) (*models.Tenant, models.PublicStorefront, error) {
	tenant, err := s.tenants.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.PublicStorefront{}, ErrStorefrontNotFound
		}
		return nil, models.PublicStorefront{}, fmt.Errorf("get storefront by slug: %w", err)
	}

	if tenant.Status != models.TenantStatusActive || !tenant.StorefrontPublished {
		return nil, models.PublicStorefront{}, ErrStorefrontNotFound
	}

	return tenant, models.PublicStorefront{
		Name:         tenant.Name,
		Slug:         tenant.Slug,
		LogoURL:      tenant.LogoURL,
		ContactEmail: tenant.ContactEmail,
		ContactPhone: tenant.ContactPhone,
		Address:      tenant.Address,
	}, nil
}
