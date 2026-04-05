package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

func TestGetPublicBySlug_ReturnsPublishedStorefront(t *testing.T) {
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:                  tenantID,
		Name:                "Funke Fabrics",
		Slug:                "funke-fabrics",
		StorefrontPublished: true,
		Status:              models.TenantStatusActive,
	}
	products := []models.PublicStorefrontProduct{{
		ID:      uuid.New(),
		Name:    "Ankara Set",
		Price:   decimal.NewFromInt(24500),
		InStock: true,
	}}

	svc := service.NewStorefrontService(
		&mockTenantRepo{tenant: tenant},
		&mockProductRepo{publicProducts: products},
	)

	out, err := svc.GetPublicBySlug(context.Background(), tenant.Slug)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Storefront.Name != tenant.Name {
		t.Fatalf("expected storefront name %q, got %q", tenant.Name, out.Storefront.Name)
	}
	if len(out.Products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(out.Products))
	}
	if !out.Products[0].Price.Equal(products[0].Price) {
		t.Fatalf("expected price %s, got %s", products[0].Price, out.Products[0].Price)
	}
}

func TestGetPublicBySlug_HidesUnpublishedStorefront(t *testing.T) {
	svc := service.NewStorefrontService(
		&mockTenantRepo{tenant: &models.Tenant{
			ID:                  uuid.New(),
			Slug:                "hidden-store",
			StorefrontPublished: false,
			Status:              models.TenantStatusActive,
		}},
		&mockProductRepo{},
	)

	_, err := svc.GetPublicBySlug(context.Background(), "hidden-store")
	if !errors.Is(err, service.ErrStorefrontNotFound) {
		t.Fatalf("expected ErrStorefrontNotFound, got %v", err)
	}
}

func TestGetPublicBySlug_HidesSuspendedStorefront(t *testing.T) {
	svc := service.NewStorefrontService(
		&mockTenantRepo{tenant: &models.Tenant{
			ID:                  uuid.New(),
			Slug:                "suspended-store",
			StorefrontPublished: true,
			Status:              models.TenantStatusSuspended,
		}},
		&mockProductRepo{},
	)

	_, err := svc.GetPublicBySlug(context.Background(), "suspended-store")
	if !errors.Is(err, service.ErrStorefrontNotFound) {
		t.Fatalf("expected ErrStorefrontNotFound, got %v", err)
	}
}

func TestGetPublicBySlug_NotFound(t *testing.T) {
	svc := service.NewStorefrontService(
		&mockTenantRepo{err: pgx.ErrNoRows},
		&mockProductRepo{},
	)

	_, err := svc.GetPublicBySlug(context.Background(), "missing-store")
	if !errors.Is(err, service.ErrStorefrontNotFound) {
		t.Fatalf("expected ErrStorefrontNotFound, got %v", err)
	}
}
