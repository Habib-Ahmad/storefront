package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

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
	tenant, err := s.tenants.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStorefrontNotFound
		}
		return nil, fmt.Errorf("get storefront by slug: %w", err)
	}

	if tenant.Status != models.TenantStatusActive || !tenant.StorefrontPublished {
		return nil, ErrStorefrontNotFound
	}

	products, err := s.products.ListPublicByTenant(ctx, tenant.ID)
	if err != nil {
		return nil, fmt.Errorf("list public storefront products: %w", err)
	}
	if products == nil {
		products = []models.PublicStorefrontProduct{}
	}

	return &models.PublicStorefrontCatalog{
		Storefront: models.PublicStorefront{
			Name:         tenant.Name,
			Slug:         tenant.Slug,
			LogoURL:      tenant.LogoURL,
			ContactEmail: tenant.ContactEmail,
			ContactPhone: tenant.ContactPhone,
			Address:      tenant.Address,
		},
		Products: products,
	}, nil
}
