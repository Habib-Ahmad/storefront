package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/adapter/shipbubble"
	"storefront/backend/internal/apperr"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var ErrDeliveryOptionUnavailable = apperr.Conflict("selected delivery option is no longer available")

type DeliveryQuoteProvider interface {
	ValidateAddress(ctx context.Context, req shipbubble.ValidateAddressRequest) (*shipbubble.ValidatedAddress, error)
	GetPackageCategories(ctx context.Context) ([]shipbubble.PackageCategory, error)
	GetPackageBoxes(ctx context.Context) ([]shipbubble.PackageBox, error)
	FetchRates(ctx context.Context, req shipbubble.RateRequest) (*shipbubble.RateResponse, error)
}

type DeliveryQuoteService struct {
	storefronts *StorefrontService
	products    repository.ProductRepository
	provider    DeliveryQuoteProvider
}

func NewDeliveryQuoteService(storefronts *StorefrontService, products repository.ProductRepository, provider DeliveryQuoteProvider) *DeliveryQuoteService {
	return &DeliveryQuoteService{storefronts: storefronts, products: products, provider: provider}
}

func (s *DeliveryQuoteService) QuotePublic(ctx context.Context, slug string, req models.PublicStorefrontDeliveryQuoteRequest) (*models.PublicStorefrontDeliveryQuoteResponse, error) {
	storefront, options, debug, err := s.fetchPublicQuotes(ctx, slug, req)
	if err != nil {
		return nil, err
	}
	return &models.PublicStorefrontDeliveryQuoteResponse{
		Storefront: storefront,
		Options:    options,
		Debug:      debug,
	}, nil
}

func (s *DeliveryQuoteService) ResolvePublicSelection(ctx context.Context, slug string, req models.PublicStorefrontDeliveryQuoteRequest, selection models.PublicStorefrontDeliveryQuoteSelection) (decimal.Decimal, error) {
	_, options, _, err := s.fetchPublicQuotes(ctx, slug, req)
	if err != nil {
		return decimal.Zero, err
	}
	for _, option := range options {
		if option.CourierID == strings.TrimSpace(selection.CourierID) && option.ServiceCode == strings.TrimSpace(selection.ServiceCode) {
			if selection.ServiceType == "" || option.ServiceType == strings.TrimSpace(selection.ServiceType) {
				return option.Amount, nil
			}
		}
	}
	return decimal.Zero, ErrDeliveryOptionUnavailable
}

func (s *DeliveryQuoteService) fetchPublicQuotes(ctx context.Context, slug string, req models.PublicStorefrontDeliveryQuoteRequest) (models.PublicStorefront, []models.PublicStorefrontDeliveryQuoteOption, *models.PublicStorefrontDeliveryQuoteDebug, error) {
	if s.provider == nil {
		return models.PublicStorefront{}, nil, nil, apperr.Conflict("delivery is not available right now")
	}
	if len(req.Items) == 0 {
		return models.PublicStorefront{}, nil, nil, apperr.Unprocessable("items is required")
	}
	if strings.TrimSpace(req.CustomerName) == "" {
		return models.PublicStorefront{}, nil, nil, apperr.Unprocessable("customer_name is required")
	}
	if strings.TrimSpace(req.CustomerPhone) == "" {
		return models.PublicStorefront{}, nil, nil, apperr.Unprocessable("customer_phone is required")
	}
	if strings.TrimSpace(req.ShippingAddress) == "" {
		return models.PublicStorefront{}, nil, nil, apperr.Unprocessable("shipping_address is required")
	}

	tenant, storefront, err := s.storefronts.getPublishedStorefront(ctx, slug)
	if err != nil {
		return models.PublicStorefront{}, nil, nil, err
	}
	if !tenant.ActiveModules.Logistics {
		return models.PublicStorefront{}, nil, nil, apperr.Forbidden("delivery is not enabled for this storefront")
	}

	senderPhone, senderPhoneFallback := quoteSenderPhone(tenant, req)
	senderAddressText, senderAddressFallback := quoteSenderAddress(tenant, req)
	assumptions := []string{}
	if senderPhoneFallback {
		assumptions = append(assumptions, "using customer phone as temporary sender contact until storefront logistics profile is completed")
	}
	if senderAddressFallback {
		assumptions = append(assumptions, "using customer delivery address as temporary sender address until storefront logistics profile is completed")
	}

	senderEmail := fallbackEmail(tenant.ContactEmail, tenant.Slug)
	receiverEmail := fallbackEmail(req.CustomerEmail, "customer")

	senderAddress, err := s.provider.ValidateAddress(ctx, shipbubble.ValidateAddressRequest{
		Name:    tenant.Name,
		Email:   senderEmail,
		Phone:   senderPhone,
		Address: senderAddressText,
	})
	if err != nil {
		return models.PublicStorefront{}, nil, nil, mapQuoteAddressValidationError("sender", err)
	}
	receiverAddress, err := s.provider.ValidateAddress(ctx, shipbubble.ValidateAddressRequest{
		Name:    strings.TrimSpace(req.CustomerName),
		Email:   receiverEmail,
		Phone:   strings.TrimSpace(req.CustomerPhone),
		Address: strings.TrimSpace(req.ShippingAddress),
	})
	if err != nil {
		return models.PublicStorefront{}, nil, nil, mapQuoteAddressValidationError("receiver", err)
	}

	categories, err := s.provider.GetPackageCategories(ctx)
	if err != nil {
		return models.PublicStorefront{}, nil, nil, fmt.Errorf("get package categories: %w", err)
	}
	boxes, err := s.provider.GetPackageBoxes(ctx)
	if err != nil {
		return models.PublicStorefront{}, nil, nil, fmt.Errorf("get package boxes: %w", err)
	}

	packageItems, productCategories, totalWeight, packageAssumptions, err := s.buildPackageItems(ctx, tenant.ID, req.Items)
	if err != nil {
		return models.PublicStorefront{}, nil, nil, err
	}
	assumptions = append(assumptions, packageAssumptions...)
	category := choosePackageCategory(categories, productCategories)
	if category.ID == 0 {
		return models.PublicStorefront{}, nil, nil, apperr.Conflict("delivery categories are not available right now")
	}
	box := choosePackageBox(boxes, totalWeight)
	if box.Name == "" {
		return models.PublicStorefront{}, nil, nil, apperr.Conflict("delivery packaging is not available right now")
	}
	if !box.MaxWeight.IsZero() {
		assumptions = append(assumptions, fmt.Sprintf("using %s package dimensions until product shipping dimensions are stored", box.Name))
	}

	rates, err := s.provider.FetchRates(ctx, shipbubble.RateRequest{
		SenderAddressCode:    senderAddress.AddressCode,
		ReceiverAddressCode:  receiverAddress.AddressCode,
		PickupDate:           pickupDate(time.Now()),
		CategoryID:           category.ID,
		PackageItems:         packageItems,
		DeliveryInstructions: strings.TrimSpace(derefString(req.DeliveryInstructions)),
		PackageDimension: shipbubble.PackageDimension{
			Length: box.Length,
			Width:  box.Width,
			Height: box.Height,
		},
		IsInvoiceRequired: false,
	})
	if err != nil {
		return models.PublicStorefront{}, nil, nil, fmt.Errorf("fetch delivery quotes: %w", err)
	}

	if len(rates.Options) == 0 {
		return models.PublicStorefront{}, nil, nil, apperr.Conflict("no delivery options are available for this address yet")
	}

	fastestID := ""
	if rates.Fastest != nil {
		fastestID = quoteOptionID(rates.Fastest.CourierID, rates.Fastest.ServiceCode, rates.Fastest.ServiceType)
	}
	cheapestID := ""
	if rates.Cheapest != nil {
		cheapestID = quoteOptionID(rates.Cheapest.CourierID, rates.Cheapest.ServiceCode, rates.Cheapest.ServiceType)
	}

	options := make([]models.PublicStorefrontDeliveryQuoteOption, 0, len(rates.Options))
	for _, option := range rates.Options {
		options = append(options, models.PublicStorefrontDeliveryQuoteOption{
			ID:             quoteOptionID(option.CourierID, option.ServiceCode, option.ServiceType),
			CourierID:      option.CourierID,
			CourierName:    option.CourierName,
			ServiceCode:    option.ServiceCode,
			ServiceType:    option.ServiceType,
			Amount:         option.Total,
			Currency:       option.Currency,
			PickupETA:      option.PickupETA,
			DeliveryETA:    option.DeliveryETA,
			TrackingLabel:  option.Tracking.Label,
			TrackingLevel:  option.TrackingLevel,
			IsFastest:      fastestID != "" && fastestID == quoteOptionID(option.CourierID, option.ServiceCode, option.ServiceType),
			IsCheapest:     cheapestID != "" && cheapestID == quoteOptionID(option.CourierID, option.ServiceCode, option.ServiceType),
			ProviderFields: option.Raw,
		})
	}

	debug := &models.PublicStorefrontDeliveryQuoteDebug{
		SenderAddressCode:   senderAddress.AddressCode,
		ReceiverAddressCode: receiverAddress.AddressCode,
		CategoryID:          category.ID,
		CategoryName:        category.Name,
		PackageBox:          box.Name,
		EstimatedWeightKG:   totalWeight,
		Assumptions:         assumptions,
		RawResponse:         rates.RawResponse,
	}

	return storefront, options, debug, nil
}

func (s *DeliveryQuoteService) buildPackageItems(ctx context.Context, tenantID uuid.UUID, req []models.PublicStorefrontDeliveryQuoteRequestItem) ([]shipbubble.PackageItem, []string, decimal.Decimal, []string, error) {
	items := make([]shipbubble.PackageItem, 0, len(req))
	productCategories := make([]string, 0, len(req))
	totalWeight := decimal.Zero
	assumptions := []string{"estimating package weight from product categories until shipping metadata is stored per product"}

	for _, item := range req {
		if item.Quantity <= 0 {
			return nil, nil, decimal.Zero, nil, apperr.Unprocessable("item quantity must be greater than zero")
		}

		variant, err := s.products.GetVariantByID(ctx, item.VariantID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil, decimal.Zero, nil, ErrStorefrontNotFound
			}
			return nil, nil, decimal.Zero, nil, fmt.Errorf("get product variant: %w", err)
		}

		product, err := s.products.GetByID(ctx, tenantID, variant.ProductID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil, decimal.Zero, nil, ErrStorefrontNotFound
			}
			return nil, nil, decimal.Zero, nil, fmt.Errorf("get product for delivery quote: %w", err)
		}
		if !product.IsAvailable || (variant.StockQty != nil && *variant.StockQty < item.Quantity) {
			return nil, nil, decimal.Zero, nil, apperr.Conflict("one or more items are no longer available")
		}

		category := strings.TrimSpace(derefString(product.Category))
		weight := estimateUnitWeight(category)
		totalWeight = totalWeight.Add(weight.Mul(decimal.NewFromInt(int64(item.Quantity))))
		productCategories = append(productCategories, category)

		items = append(items, shipbubble.PackageItem{
			Name:        product.Name,
			Description: strings.TrimSpace(derefString(product.Description)),
			UnitWeight:  weight,
			UnitAmount:  variant.Price.String(),
			Quantity:    strconv.Itoa(item.Quantity),
		})
	}

	if totalWeight.IsZero() {
		totalWeight = decimal.RequireFromString("0.25")
	}

	return items, productCategories, totalWeight, assumptions, nil
}

func fallbackEmail(email *string, label string) string {
	if email != nil && strings.TrimSpace(*email) != "" {
		return strings.TrimSpace(*email)
	}
	return fmt.Sprintf("%s@storefront.local", sanitizeEmailLabel(label))
}

func quoteSenderPhone(tenant *models.Tenant, req models.PublicStorefrontDeliveryQuoteRequest) (string, bool) {
	if tenant.ContactPhone != nil && strings.TrimSpace(*tenant.ContactPhone) != "" {
		return strings.TrimSpace(*tenant.ContactPhone), false
	}
	return strings.TrimSpace(req.CustomerPhone), true
}

func quoteSenderAddress(tenant *models.Tenant, req models.PublicStorefrontDeliveryQuoteRequest) (string, bool) {
	if tenant.Address != nil && strings.TrimSpace(*tenant.Address) != "" {
		return strings.TrimSpace(*tenant.Address), false
	}
	return strings.TrimSpace(req.ShippingAddress), true
}

func mapQuoteAddressValidationError(role string, err error) error {
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(message, "address/validate") || strings.Contains(message, "validate address") || strings.Contains(message, "couldn't validate the provided address") {
		if role == "sender" {
			return apperr.Unprocessable("the store pickup address is incomplete. Ask the store admin to add a clear street, city, state, and country in logistics setup")
		}
		return apperr.Unprocessable("we couldn't validate this delivery address. Enter a clear street, city, state, and country")
	}

	return fmt.Errorf("validate %s address: %w", role, err)
}

func sanitizeEmailLabel(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "delivery"
	}
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_':
			builder.WriteRune(r)
		}
	}
	if builder.Len() == 0 {
		return "delivery"
	}
	return builder.String()
}

func pickupDate(now time.Time) string {
	if now.Hour() >= 18 {
		now = now.Add(24 * time.Hour)
	}
	return now.Format("2006-01-02")
}

func quoteOptionID(courierID, serviceCode, serviceType string) string {
	return strings.Join([]string{strings.TrimSpace(courierID), strings.TrimSpace(serviceCode), strings.TrimSpace(serviceType)}, ":")
}

func choosePackageCategory(categories []shipbubble.PackageCategory, productCategories []string) shipbubble.PackageCategory {
	if len(categories) == 0 {
		return shipbubble.PackageCategory{}
	}

	preferred := []string{}
	for _, category := range productCategories {
		normalized := strings.ToLower(strings.TrimSpace(category))
		switch {
		case strings.Contains(normalized, "food"):
			preferred = append(preferred, "food")
		case strings.Contains(normalized, "jewel"):
			preferred = append(preferred, "jewelry")
		case strings.Contains(normalized, "elect"):
			preferred = append(preferred, "electronic")
		case strings.Contains(normalized, "fashion"), strings.Contains(normalized, "cloth"), strings.Contains(normalized, "wear"):
			preferred = append(preferred, "fashion")
		default:
			preferred = append(preferred, "accessories")
		}
	}

	for _, want := range preferred {
		for _, category := range categories {
			if strings.Contains(strings.ToLower(category.Name), want) {
				return category
			}
		}
	}

	for _, category := range categories {
		if strings.Contains(strings.ToLower(category.Name), "accessories") {
			return category
		}
	}
	return categories[0]
}

func choosePackageBox(boxes []shipbubble.PackageBox, totalWeight decimal.Decimal) shipbubble.PackageBox {
	if len(boxes) == 0 {
		return shipbubble.PackageBox{}
	}
	selected := boxes[0]
	for _, box := range boxes {
		if box.MaxWeight.GreaterThanOrEqual(totalWeight) {
			if selected.MaxWeight.IsZero() || box.MaxWeight.LessThan(selected.MaxWeight) || selected.MaxWeight.LessThan(totalWeight) {
				selected = box
			}
			continue
		}
		if selected.MaxWeight.LessThan(totalWeight) && box.MaxWeight.GreaterThan(selected.MaxWeight) {
			selected = box
		}
	}
	return selected
}

func estimateUnitWeight(category string) decimal.Decimal {
	normalized := strings.ToLower(strings.TrimSpace(category))
	switch {
	case strings.Contains(normalized, "food"):
		return decimal.RequireFromString("1.00")
	case strings.Contains(normalized, "elect"):
		return decimal.RequireFromString("1.50")
	case strings.Contains(normalized, "jewel"):
		return decimal.RequireFromString("0.10")
	case strings.Contains(normalized, "fashion"), strings.Contains(normalized, "cloth"), strings.Contains(normalized, "wear"):
		return decimal.RequireFromString("0.35")
	default:
		return decimal.RequireFromString("0.25")
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
