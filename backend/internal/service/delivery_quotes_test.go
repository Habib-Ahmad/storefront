package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/adapter/shipbubble"
	"storefront/backend/internal/apperr"
	"storefront/backend/internal/db"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

type stubQuoteTenantRepo struct {
	tenant *models.Tenant
}

func (s *stubQuoteTenantRepo) Create(_ context.Context, t *models.Tenant) error {
	t.ID = uuid.New()
	return nil
}
func (s *stubQuoteTenantRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Tenant, error) {
	return s.tenant, nil
}
func (s *stubQuoteTenantRepo) GetBySlug(_ context.Context, slug string) (*models.Tenant, error) {
	if s.tenant == nil || s.tenant.Slug != slug {
		return nil, pgx.ErrNoRows
	}
	return s.tenant, nil
}
func (s *stubQuoteTenantRepo) Update(_ context.Context, _ *models.Tenant) error { return nil }
func (s *stubQuoteTenantRepo) SoftDelete(_ context.Context, _ uuid.UUID) error  { return nil }
func (s *stubQuoteTenantRepo) WithTx(_ db.DBTX) repository.TenantRepository     { return s }

type stubQuoteProductRepo struct {
	product *models.Product
	variant *models.ProductVariant
}

func (s *stubQuoteProductRepo) Create(_ context.Context, p *models.Product) error {
	p.ID = uuid.New()
	return nil
}
func (s *stubQuoteProductRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.Product, error) {
	if s.product == nil || s.product.TenantID != tenantID || s.product.ID != id {
		return nil, pgx.ErrNoRows
	}
	return s.product, nil
}
func (s *stubQuoteProductRepo) ListByTenant(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Product, error) {
	return nil, nil
}
func (s *stubQuoteProductRepo) ListPublicByTenant(_ context.Context, _ uuid.UUID) ([]models.PublicStorefrontProduct, error) {
	return nil, nil
}
func (s *stubQuoteProductRepo) CountByTenant(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}
func (s *stubQuoteProductRepo) Update(_ context.Context, _ *models.Product) error  { return nil }
func (s *stubQuoteProductRepo) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return nil }
func (s *stubQuoteProductRepo) CreateVariant(_ context.Context, v *models.ProductVariant) error {
	v.ID = uuid.New()
	return nil
}
func (s *stubQuoteProductRepo) GetVariantByID(_ context.Context, id uuid.UUID) (*models.ProductVariant, error) {
	if s.variant == nil || s.variant.ID != id {
		return nil, pgx.ErrNoRows
	}
	return s.variant, nil
}
func (s *stubQuoteProductRepo) ListVariants(_ context.Context, _ uuid.UUID) ([]models.ProductVariant, error) {
	return nil, nil
}
func (s *stubQuoteProductRepo) UpdateVariant(_ context.Context, _ *models.ProductVariant) error {
	return nil
}
func (s *stubQuoteProductRepo) DecrementStock(_ context.Context, _ uuid.UUID, _ int) error {
	return nil
}
func (s *stubQuoteProductRepo) RestoreStock(_ context.Context, _ uuid.UUID, _ int) error { return nil }
func (s *stubQuoteProductRepo) SoftDeleteVariant(_ context.Context, _ uuid.UUID) error   { return nil }
func (s *stubQuoteProductRepo) AddImage(_ context.Context, _ *models.ProductImage) error { return nil }
func (s *stubQuoteProductRepo) ListImagesByProduct(_ context.Context, _ uuid.UUID) ([]models.ProductImage, error) {
	return nil, nil
}
func (s *stubQuoteProductRepo) UpdateImage(_ context.Context, _ *models.ProductImage) error {
	return nil
}
func (s *stubQuoteProductRepo) DeleteImage(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubQuoteProductRepo) WithTx(_ db.DBTX) repository.ProductRepository    { return s }

type stubQuoteProvider struct {
	validatedAddress *shipbubble.ValidatedAddress
	validateErr      error
	categories       []shipbubble.PackageCategory
	boxes            []shipbubble.PackageBox
	rateResponse     *shipbubble.RateResponse
	lastRateRequest  shipbubble.RateRequest
	validateRequests []shipbubble.ValidateAddressRequest
	validateCalls    int
	rateCalls        int
}

func (s *stubQuoteProvider) ValidateAddress(_ context.Context, req shipbubble.ValidateAddressRequest) (*shipbubble.ValidatedAddress, error) {
	s.validateRequests = append(s.validateRequests, req)
	s.validateCalls++
	if s.validateErr != nil {
		return nil, s.validateErr
	}
	if s.validatedAddress != nil {
		copy := *s.validatedAddress
		copy.AddressCode += int64(s.validateCalls - 1)
		return &copy, nil
	}
	return &shipbubble.ValidatedAddress{AddressCode: int64(s.validateCalls)}, nil
}

func (s *stubQuoteProvider) GetPackageCategories(_ context.Context) ([]shipbubble.PackageCategory, error) {
	return s.categories, nil
}

func (s *stubQuoteProvider) GetPackageBoxes(_ context.Context) ([]shipbubble.PackageBox, error) {
	return s.boxes, nil
}

func (s *stubQuoteProvider) FetchRates(_ context.Context, req shipbubble.RateRequest) (*shipbubble.RateResponse, error) {
	s.rateCalls++
	s.lastRateRequest = req
	if s.rateResponse == nil {
		return nil, errors.New("missing rate response")
	}
	return s.rateResponse, nil
}

func TestDeliveryQuoteService_QuotePublic_NormalizesRates(t *testing.T) {
	tenantID := uuid.New()
	variantID := uuid.New()
	tenantAddress := "12 Allen Avenue, Ikeja, Lagos, Nigeria"
	tenantPhone := "+2348012345678"
	tenantEmail := "hello@funkefabrics.com"
	providerRaw := json.RawMessage(`{"request_token":"quote-token"}`)

	storefronts := NewStorefrontService(&stubQuoteTenantRepo{tenant: &models.Tenant{
		ID:                  tenantID,
		Name:                "Funke Fabrics",
		Slug:                "funke-fabrics",
		StorefrontPublished: true,
		ContactEmail:        &tenantEmail,
		ContactPhone:        &tenantPhone,
		Address:             &tenantAddress,
		ActiveModules:       models.ActiveModules{Logistics: true},
		Status:              models.TenantStatusActive,
	}}, &stubQuoteProductRepo{})
	products := &stubQuoteProductRepo{
		product: &models.Product{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "Ankara Set",
			Description: ptrString("A bright two-piece set"),
			Category:    ptrString("Fashion"),
			IsAvailable: true,
		},
		variant: &models.ProductVariant{
			ID:        variantID,
			ProductID: uuid.Nil,
			Price:     decimal.NewFromInt(24500),
			IsDefault: true,
		},
	}
	products.variant.ProductID = products.product.ID
	provider := &stubQuoteProvider{
		validatedAddress: &shipbubble.ValidatedAddress{AddressCode: 1001},
		categories: []shipbubble.PackageCategory{
			{ID: 1, Name: "Accessories"},
			{ID: 2, Name: "Fashion wears"},
		},
		boxes: []shipbubble.PackageBox{
			{Name: "small box", Length: decimal.NewFromInt(10), Width: decimal.NewFromInt(10), Height: decimal.NewFromInt(10), MaxWeight: decimal.RequireFromString("0.50")},
			{Name: "medium box", Length: decimal.NewFromInt(16), Width: decimal.NewFromInt(12), Height: decimal.NewFromInt(10), MaxWeight: decimal.RequireFromString("2.00")},
		},
		rateResponse: &shipbubble.RateResponse{
			RequestToken: "quote-token",
			Options: []shipbubble.RateOption{{
				CourierID:     "123",
				CourierName:   "Kwik",
				ServiceCode:   "bike",
				ServiceType:   "dropoff",
				Currency:      "NGN",
				Total:         decimal.NewFromInt(3500),
				Tracking:      shipbubble.TrackingSummary{Label: "Full tracking"},
				TrackingLevel: 4,
				Raw:           json.RawMessage(`{"courier_name":"Kwik"}`),
			}},
			Fastest:     &shipbubble.RateOption{CourierID: "123", ServiceCode: "bike", ServiceType: "dropoff"},
			Cheapest:    &shipbubble.RateOption{CourierID: "123", ServiceCode: "bike", ServiceType: "dropoff"},
			RawResponse: providerRaw,
		},
	}

	svc := NewDeliveryQuoteService(storefronts, products, provider)
	resp, err := svc.QuotePublic(context.Background(), "funke-fabrics", models.PublicStorefrontDeliveryQuoteRequest{
		CustomerName:    "Chidi",
		CustomerPhone:   "08012345678",
		ShippingAddress: "23 Abuja",
		Items:           []models.PublicStorefrontDeliveryQuoteRequestItem{{VariantID: variantID, Quantity: 2}},
	})
	if err != nil {
		t.Fatalf("QuotePublic returned error: %v", err)
	}
	if len(resp.Options) != 1 {
		t.Fatalf("expected 1 quote option, got %d", len(resp.Options))
	}
	if resp.Options[0].CourierName != "Kwik" {
		t.Fatalf("unexpected courier name: %s", resp.Options[0].CourierName)
	}
	if !resp.Options[0].IsFastest || !resp.Options[0].IsCheapest {
		t.Fatal("expected the single option to be marked fastest and cheapest")
	}
	if provider.lastRateRequest.CategoryID != 2 {
		t.Fatalf("expected fashion category ID 2, got %d", provider.lastRateRequest.CategoryID)
	}
	if len(provider.lastRateRequest.PackageItems) != 1 {
		t.Fatalf("expected one package item, got %d", len(provider.lastRateRequest.PackageItems))
	}
	if provider.lastRateRequest.PackageItems[0].Quantity != "2" {
		t.Fatalf("expected quantity 2, got %s", provider.lastRateRequest.PackageItems[0].Quantity)
	}
	if provider.lastRateRequest.PackageDimension.Length.String() != "16" {
		t.Fatalf("expected medium box dimensions to be used, got %+v", provider.lastRateRequest.PackageDimension)
	}
	if resp.Debug == nil || resp.Debug.PackageBox != "medium box" {
		t.Fatalf("expected debug package box to be medium box, got %+v", resp.Debug)
	}
	if resp.Debug.EstimatedWeightKG.String() != "0.7" {
		t.Fatalf("expected estimated weight 0.7kg, got %s", resp.Debug.EstimatedWeightKG.String())
	}
}

func TestDeliveryQuoteService_ResolvePublicSelection_RejectsUnavailableOption(t *testing.T) {
	tenantID := uuid.New()
	variantID := uuid.New()
	tenantAddress := "12 Allen Avenue, Ikeja, Lagos, Nigeria"
	tenantPhone := "+2348012345678"
	tenantEmail := "hello@funkefabrics.com"
	products := &stubQuoteProductRepo{
		product: &models.Product{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "Ankara Set",
			Category:    ptrString("Fashion"),
			IsAvailable: true,
		},
		variant: &models.ProductVariant{
			ID:        variantID,
			Price:     decimal.NewFromInt(24500),
			IsDefault: true,
		},
	}
	products.variant.ProductID = products.product.ID
	provider := &stubQuoteProvider{
		validatedAddress: &shipbubble.ValidatedAddress{AddressCode: 1001},
		categories:       []shipbubble.PackageCategory{{ID: 2, Name: "Fashion wears"}},
		boxes:            []shipbubble.PackageBox{{Name: "medium box", Length: decimal.NewFromInt(16), Width: decimal.NewFromInt(12), Height: decimal.NewFromInt(10), MaxWeight: decimal.RequireFromString("2.00")}},
		rateResponse: &shipbubble.RateResponse{
			Options: []shipbubble.RateOption{{
				CourierID:   "123",
				CourierName: "Kwik",
				ServiceCode: "bike",
				ServiceType: "dropoff",
				Currency:    "NGN",
				Total:       decimal.NewFromInt(3500),
			}},
		},
	}
	svc := NewDeliveryQuoteService(
		NewStorefrontService(&stubQuoteTenantRepo{tenant: &models.Tenant{
			ID:                  tenantID,
			Name:                "Funke Fabrics",
			Slug:                "funke-fabrics",
			StorefrontPublished: true,
			ContactEmail:        &tenantEmail,
			ContactPhone:        &tenantPhone,
			Address:             &tenantAddress,
			ActiveModules:       models.ActiveModules{Logistics: true},
			Status:              models.TenantStatusActive,
		}}, products),
		products,
		provider,
	)

	_, err := svc.ResolvePublicSelection(context.Background(), "funke-fabrics", models.PublicStorefrontDeliveryQuoteRequest{
		CustomerName:    "Chidi",
		CustomerPhone:   "08012345678",
		ShippingAddress: "23 Abuja",
		Items:           []models.PublicStorefrontDeliveryQuoteRequestItem{{VariantID: variantID, Quantity: 1}},
	}, models.PublicStorefrontDeliveryQuoteSelection{CourierID: "999", ServiceCode: "van"})
	if !errors.Is(err, ErrDeliveryOptionUnavailable) {
		t.Fatalf("expected ErrDeliveryOptionUnavailable, got %v", err)
	}
}

func TestDeliveryQuoteService_QuotePublic_NormalizesProviderNames(t *testing.T) {
	tenantID := uuid.New()
	variantID := uuid.New()
	tenantAddress := "Plot 72 Ahmadu Bello Way, Abuja, Abuja, FCT, Nigeria"
	tenantPhone := "+2348011223344"
	tenantEmail := "user1@gmail.com"

	storefronts := NewStorefrontService(&stubQuoteTenantRepo{tenant: &models.Tenant{
		ID:                  tenantID,
		Name:                "User1 Store Updated",
		Slug:                "user1-store",
		StorefrontPublished: true,
		ContactEmail:        &tenantEmail,
		ContactPhone:        &tenantPhone,
		Address:             &tenantAddress,
		ActiveModules:       models.ActiveModules{Logistics: true},
		Status:              models.TenantStatusActive,
	}}, &stubQuoteProductRepo{})
	products := &stubQuoteProductRepo{
		product: &models.Product{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "Ankara Set",
			Category:    ptrString("Fashion"),
			IsAvailable: true,
		},
		variant: &models.ProductVariant{
			ID:        variantID,
			Price:     decimal.NewFromInt(24500),
			IsDefault: true,
		},
	}
	products.variant.ProductID = products.product.ID
	provider := &stubQuoteProvider{
		validatedAddress: &shipbubble.ValidatedAddress{AddressCode: 1001},
		categories:       []shipbubble.PackageCategory{{ID: 2, Name: "Fashion wears"}},
		boxes:            []shipbubble.PackageBox{{Name: "medium box", Length: decimal.NewFromInt(16), Width: decimal.NewFromInt(12), Height: decimal.NewFromInt(10), MaxWeight: decimal.RequireFromString("2.00")}},
		rateResponse: &shipbubble.RateResponse{
			RequestToken: "quote-token",
			Options: []shipbubble.RateOption{{
				CourierID:   "123",
				CourierName: "Kwik",
				ServiceCode: "bike",
				ServiceType: "dropoff",
				Currency:    "NGN",
				Total:       decimal.NewFromInt(3500),
				Tracking:    shipbubble.TrackingSummary{Label: "Full tracking"},
			}},
		},
	}

	svc := NewDeliveryQuoteService(storefronts, products, provider)
	_, err := svc.QuotePublic(context.Background(), "user1-store", models.PublicStorefrontDeliveryQuoteRequest{
		CustomerName:    "Ada1",
		CustomerPhone:   "08012345678",
		ShippingAddress: "23 Broad Street, Lagos, Lagos, Nigeria",
		Items:           []models.PublicStorefrontDeliveryQuoteRequestItem{{VariantID: variantID, Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("QuotePublic returned error: %v", err)
	}
	if len(provider.validateRequests) != 2 {
		t.Fatalf("expected 2 validation requests, got %d", len(provider.validateRequests))
	}
	if provider.validateRequests[0].Name != "User Store Updated" {
		t.Fatalf("expected normalized sender name, got %q", provider.validateRequests[0].Name)
	}
	if provider.validateRequests[1].Name != "Ada Customer" {
		t.Fatalf("expected normalized receiver name, got %q", provider.validateRequests[1].Name)
	}
}

func TestDeliveryQuoteService_QuotePublic_ReturnsProviderWalletFundingError(t *testing.T) {
	tenantID := uuid.New()
	variantID := uuid.New()
	tenantAddress := "12 Allen Avenue, Ikeja, Lagos, Nigeria"
	tenantPhone := "+2348012345678"
	tenantEmail := "hello@funkefabrics.com"

	storefronts := NewStorefrontService(&stubQuoteTenantRepo{tenant: &models.Tenant{
		ID:                  tenantID,
		Name:                "Funke Fabrics",
		Slug:                "funke-fabrics",
		StorefrontPublished: true,
		ContactEmail:        &tenantEmail,
		ContactPhone:        &tenantPhone,
		Address:             &tenantAddress,
		ActiveModules:       models.ActiveModules{Logistics: true},
		Status:              models.TenantStatusActive,
	}}, &stubQuoteProductRepo{})
	products := &stubQuoteProductRepo{
		product: &models.Product{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "Ankara Set",
			Category:    ptrString("Fashion"),
			IsAvailable: true,
		},
		variant: &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(24500), IsDefault: true},
	}
	products.variant.ProductID = products.product.ID
	provider := &stubQuoteProvider{
		validateErr: errors.New("shipbubble: validate address: Insufficient wallet balance, Please fund your wallet to validate new addresses"),
	}

	svc := NewDeliveryQuoteService(storefronts, products, provider)
	_, err := svc.QuotePublic(context.Background(), "funke-fabrics", models.PublicStorefrontDeliveryQuoteRequest{
		CustomerName:    "Chidi",
		CustomerPhone:   "08012345678",
		ShippingAddress: "23 Broad Street, Lagos, Lagos, Nigeria",
		Items:           []models.PublicStorefrontDeliveryQuoteRequestItem{{VariantID: variantID, Quantity: 1}},
	})
	if err == nil {
		t.Fatal("expected wallet funding error")
	}
	appErr, ok := err.(*apperr.Error)
	if !ok {
		t.Fatalf("expected apperr.Error, got %T", err)
	}
	if appErr.Status != 409 {
		t.Fatalf("expected conflict status, got %d", appErr.Status)
	}
	if appErr.Message != "delivery is temporarily unavailable because the shipping provider wallet needs funding" {
		t.Fatalf("unexpected error message: %q", appErr.Message)
	}
}

func TestDeliveryQuoteService_QuotePublic_RejectsIncompleteStorefrontLogisticsProfile(t *testing.T) {
	tenantID := uuid.New()
	variantID := uuid.New()
	products := &stubQuoteProductRepo{
		product: &models.Product{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "Ankara Set",
			Category:    ptrString("Fashion"),
			IsAvailable: true,
		},
		variant: &models.ProductVariant{
			ID:        variantID,
			Price:     decimal.NewFromInt(24500),
			IsDefault: true,
		},
	}
	products.variant.ProductID = products.product.ID
	provider := &stubQuoteProvider{
		validatedAddress: &shipbubble.ValidatedAddress{AddressCode: 1001},
		categories:       []shipbubble.PackageCategory{{ID: 2, Name: "Fashion wears"}},
		boxes:            []shipbubble.PackageBox{{Name: "medium box", Length: decimal.NewFromInt(16), Width: decimal.NewFromInt(12), Height: decimal.NewFromInt(10), MaxWeight: decimal.RequireFromString("2.00")}},
		rateResponse: &shipbubble.RateResponse{
			Options: []shipbubble.RateOption{{
				CourierID:   "123",
				CourierName: "Kwik",
				ServiceCode: "bike",
				ServiceType: "dropoff",
				Currency:    "NGN",
				Total:       decimal.NewFromInt(3500),
			}},
		},
	}
	svc := NewDeliveryQuoteService(
		NewStorefrontService(&stubQuoteTenantRepo{tenant: &models.Tenant{
			ID:                  tenantID,
			Name:                "Funke Fabrics",
			Slug:                "funke-fabrics",
			StorefrontPublished: true,
			ActiveModules:       models.ActiveModules{Logistics: true},
			Status:              models.TenantStatusActive,
		}}, products),
		products,
		provider,
	)

	_, err := svc.QuotePublic(context.Background(), "funke-fabrics", models.PublicStorefrontDeliveryQuoteRequest{
		CustomerName:    "Chidi",
		CustomerPhone:   "08012345678",
		ShippingAddress: "23 Abuja",
		Items:           []models.PublicStorefrontDeliveryQuoteRequestItem{{VariantID: variantID, Quantity: 1}},
	})
	if err == nil {
		t.Fatal("expected delivery readiness error")
	}
	status, message := apperr.HTTPError(err)
	if status != 409 {
		t.Fatalf("expected status 409, got %d", status)
	}
	if message != "Delivery is temporarily unavailable while the store completes its pickup profile." {
		t.Fatalf("unexpected message: %s", message)
	}
	if provider.validateCalls != 0 {
		t.Fatalf("expected no address validations, got %d", provider.validateCalls)
	}
}

func TestDeliveryQuoteService_QuotePublic_ReturnsValidationErrorForInvalidReceiverAddress(t *testing.T) {
	tenantID := uuid.New()
	variantID := uuid.New()
	tenantAddress := "16 Owerri Street, Gwarinpa, Abuja, FCT, Nigeria"
	tenantPhone := "+2348012345678"
	tenantEmail := "hello@funkefabrics.com"
	products := &stubQuoteProductRepo{
		product: &models.Product{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "Ankara Set",
			Category:    ptrString("Fashion"),
			IsAvailable: true,
		},
		variant: &models.ProductVariant{
			ID:        variantID,
			Price:     decimal.NewFromInt(24500),
			IsDefault: true,
		},
	}
	products.variant.ProductID = products.product.ID
	provider := &stubQuoteProvider{
		validateErr: errors.New(`shipbubble POST /shipping/address/validate: status 400: {"status":"failed","message":"Sorry, we couldn't validate the provided address. Please provide a clear and accurate address including the city, state and country of your address"}`),
	}
	svc := NewDeliveryQuoteService(
		NewStorefrontService(&stubQuoteTenantRepo{tenant: &models.Tenant{
			ID:                  tenantID,
			Name:                "Funke Fabrics",
			Slug:                "funke-fabrics",
			StorefrontPublished: true,
			ContactEmail:        &tenantEmail,
			ContactPhone:        &tenantPhone,
			Address:             &tenantAddress,
			ActiveModules:       models.ActiveModules{Logistics: true},
			Status:              models.TenantStatusActive,
		}}, products),
		products,
		provider,
	)

	_, err := svc.QuotePublic(context.Background(), "funke-fabrics", models.PublicStorefrontDeliveryQuoteRequest{
		CustomerName:    "Chidi",
		CustomerPhone:   "08012345678",
		ShippingAddress: "23 Abuja",
		Items:           []models.PublicStorefrontDeliveryQuoteRequestItem{{VariantID: variantID, Quantity: 1}},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	status, message := apperr.HTTPError(err)
	if status != 422 {
		t.Fatalf("expected status 422, got %d", status)
	}
	if message != "the store pickup address could not be validated. Ask the store admin to review logistics setup" {
		t.Fatalf("unexpected message: %s", message)
	}
}

func ptrString(value string) *string {
	return &value
}
