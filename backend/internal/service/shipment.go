package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

var ErrShipmentSelectionRequired = apperr.Unprocessable("courier_id and service_code are required")

type ShipmentProvider interface {
	ValidateAddress(ctx context.Context, req shipbubble.ValidateAddressRequest) (*shipbubble.ValidatedAddress, error)
	GetPackageCategories(ctx context.Context) ([]shipbubble.PackageCategory, error)
	GetPackageBoxes(ctx context.Context) ([]shipbubble.PackageBox, error)
	FetchRates(ctx context.Context, req shipbubble.RateRequest) (*shipbubble.RateResponse, error)
	CreateShipment(ctx context.Context, req shipbubble.CreateShipmentRequest) (*shipbubble.ShipmentRecord, error)
}

type DispatchShipmentRequest struct {
	CourierID   string
	ServiceCode string
	ServiceType string
}

type ShipmentService struct {
	provider  ShipmentProvider
	shipments repository.ShipmentRepository
	orders    repository.OrderRepository
	products  repository.ProductRepository
	tenants   repository.TenantRepository
	walletSvc *WalletService
}

func NewShipmentService(
	provider ShipmentProvider,
	shipments repository.ShipmentRepository,
	orders repository.OrderRepository,
	products repository.ProductRepository,
	tenants repository.TenantRepository,
	walletSvc *WalletService,
) *ShipmentService {
	return &ShipmentService{
		provider:  provider,
		shipments: shipments,
		orders:    orders,
		products:  products,
		tenants:   tenants,
		walletSvc: walletSvc,
	}
}

func (s *ShipmentService) Dispatch(ctx context.Context, orderID, tenantID uuid.UUID, req DispatchShipmentRequest) (*models.Shipment, error) {
	if s.provider == nil {
		return nil, apperr.Conflict("delivery is not available right now")
	}
	if strings.TrimSpace(req.CourierID) == "" || strings.TrimSpace(req.ServiceCode) == "" {
		return nil, ErrShipmentSelectionRequired
	}

	order, err := s.orders.GetByID(ctx, tenantID, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	if !order.IsDelivery {
		return nil, apperr.Unprocessable("pickup orders cannot be dispatched")
	}
	if order.PaymentStatus != models.PaymentStatusPaid {
		return nil, apperr.Conflict("only paid delivery orders can be dispatched")
	}
	if order.FulfillmentStatus != models.FulfillmentStatusProcessing {
		return nil, apperr.Unprocessable("order is not in a dispatchable state")
	}

	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant: %w", err)
	}

	shipment, err := s.bookShipment(ctx, tenant, order, req)
	if err != nil {
		return nil, err
	}

	if err := s.orders.UpdateFulfillmentStatus(ctx, tenantID, orderID, models.FulfillmentStatusShipped); err != nil {
		return nil, fmt.Errorf("update fulfillment status: %w", err)
	}

	return shipment, nil
}

func (s *ShipmentService) QuoteDispatchOptions(ctx context.Context, orderID, tenantID uuid.UUID) ([]models.DispatchShipmentOption, error) {
	if s.provider == nil {
		return nil, apperr.Conflict("delivery is not available right now")
	}

	order, err := s.orders.GetByID(ctx, tenantID, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	if !order.IsDelivery {
		return nil, apperr.Unprocessable("pickup orders cannot be dispatched")
	}
	if order.PaymentStatus != models.PaymentStatusPaid {
		return nil, apperr.Conflict("only paid delivery orders can be dispatched")
	}
	if order.FulfillmentStatus != models.FulfillmentStatusProcessing {
		return nil, apperr.Unprocessable("order is not in a dispatchable state")
	}

	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant: %w", err)
	}

	rates, err := s.quoteRates(ctx, tenant, order, "")
	if err != nil {
		return nil, err
	}

	return dispatchOptionsFromRates(rates), nil
}

func (s *ShipmentService) HandleStatusUpdate(ctx context.Context, carrierRef, status string, payload []byte) error {
	carrierRef = strings.TrimSpace(carrierRef)
	if carrierRef == "" {
		return apperr.Unprocessable("shipment reference is required")
	}

	shipment, err := s.shipments.GetByCarrierRef(ctx, carrierRef)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NotFound("shipment not found")
		}
		return fmt.Errorf("get shipment by carrier ref: %w", err)
	}

	if len(payload) > 0 {
		if err := s.shipments.AppendCarrierEvent(ctx, shipment.TenantID, shipment.ID, payload); err != nil {
			return fmt.Errorf("append carrier event: %w", err)
		}
	}

	order, err := s.orders.GetByIDInternal(ctx, shipment.OrderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	switch normalizeShipbubbleStatus(status) {
	case models.ShipmentStatusDelivered:
		return s.handleDelivered(ctx, order, shipment)
	case models.ShipmentStatusFailed:
		return s.handleFailed(ctx, order, shipment)
	case models.ShipmentStatusPickedUp, models.ShipmentStatusInTransit, models.ShipmentStatusQueued:
		return s.handleInFlight(ctx, order, shipment, normalizeShipbubbleStatus(status))
	default:
		return nil
	}
}

func (s *ShipmentService) bookShipment(ctx context.Context, tenant *models.Tenant, order *models.Order, req DispatchShipmentRequest) (*models.Shipment, error) {
	rates, err := s.quoteRates(ctx, tenant, order, req.ServiceType)
	if err != nil {
		return nil, err
	}

	selection := pickRateOption(rates.Options, req)
	if selection == nil {
		return nil, ErrDeliveryOptionUnavailable
	}

	created, err := s.provider.CreateShipment(ctx, shipbubble.CreateShipmentRequest{
		RequestToken: rates.RequestToken,
		ServiceCode:  selection.ServiceCode,
		CourierID:    selection.CourierID,
	})
	if err != nil {
		return nil, fmt.Errorf("create shipment: %w", err)
	}

	history, err := marshalCarrierHistory(created.Raw)
	if err != nil {
		return nil, fmt.Errorf("marshal shipment history: %w", err)
	}
	carrierRef := strings.TrimSpace(created.OrderID)
	trackingNumber := nullableTrimmedString(created.Courier.TrackingCode)
	shipment := &models.Shipment{
		OrderID:        order.ID,
		TenantID:       tenant.ID,
		Status:         normalizeShipbubbleStatus(created.Status),
		CarrierRef:     &carrierRef,
		TrackingNumber: trackingNumber,
		CarrierHistory: history,
	}
	if err := s.shipments.UpsertDispatch(ctx, shipment); err != nil {
		return nil, fmt.Errorf("save shipment: %w", err)
	}

	return shipment, nil
}

func (s *ShipmentService) quoteRates(ctx context.Context, tenant *models.Tenant, order *models.Order, serviceType string) (*shipbubble.RateResponse, error) {
	senderPhone := strings.TrimSpace(derefString(tenant.ContactPhone))
	if senderPhone == "" {
		return nil, apperr.Unprocessable("store pickup phone is required before dispatching orders")
	}
	senderAddressText := strings.TrimSpace(derefString(tenant.Address))
	if senderAddressText == "" {
		return nil, apperr.Unprocessable("store pickup address is required before dispatching orders")
	}
	receiverPhone := strings.TrimSpace(derefString(order.CustomerPhone))
	if receiverPhone == "" {
		return nil, ErrDeliveryFieldsMissing
	}
	receiverAddressText := strings.TrimSpace(derefString(order.ShippingAddress))
	if receiverAddressText == "" {
		return nil, ErrDeliveryFieldsMissing
	}

	senderAddress, err := s.provider.ValidateAddress(ctx, shipbubble.ValidateAddressRequest{
		Name:    shipbubbleSenderName(tenant.Name, tenant.Slug),
		Email:   fallbackEmail(tenant.ContactEmail, tenant.Slug),
		Phone:   senderPhone,
		Address: senderAddressText,
	})
	if err != nil {
		return nil, mapShipmentAddressValidationError("sender", err)
	}
	receiverAddress, err := s.provider.ValidateAddress(ctx, shipbubble.ValidateAddressRequest{
		Name:    shipbubbleReceiverName(derefString(order.CustomerName)),
		Email:   fallbackEmail(order.CustomerEmail, "customer"),
		Phone:   receiverPhone,
		Address: receiverAddressText,
	})
	if err != nil {
		return nil, mapShipmentAddressValidationError("receiver", err)
	}

	packageItems, productCategories, totalWeight, err := s.buildPackageItems(ctx, tenant.ID, order)
	if err != nil {
		return nil, err
	}
	categories, err := s.provider.GetPackageCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("get package categories: %w", err)
	}
	boxes, err := s.provider.GetPackageBoxes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get package boxes: %w", err)
	}
	category := choosePackageCategory(categories, productCategories)
	if category.ID == 0 {
		return nil, apperr.Conflict("delivery categories are not available right now")
	}
	box := choosePackageBox(boxes, totalWeight)
	if box.Name == "" {
		return nil, apperr.Conflict("delivery packaging is not available right now")
	}

	rates, err := s.provider.FetchRates(ctx, shipbubble.RateRequest{
		SenderAddressCode:   senderAddress.AddressCode,
		ReceiverAddressCode: receiverAddress.AddressCode,
		PickupDate:          pickupDate(time.Now()),
		CategoryID:          category.ID,
		PackageItems:        packageItems,
		ServiceType:         strings.TrimSpace(serviceType),
		PackageDimension: shipbubble.PackageDimension{
			Length: box.Length,
			Width:  box.Width,
			Height: box.Height,
		},
		IsInvoiceRequired: false,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch dispatch rates: %w", err)
	}
	if len(rates.Options) == 0 {
		return nil, apperr.Conflict("no courier options are available for this delivery order yet")
	}

	return rates, nil
}

func dispatchOptionsFromRates(rates *shipbubble.RateResponse) []models.DispatchShipmentOption {
	fastestID := ""
	if rates.Fastest != nil {
		fastestID = quoteOptionID(rates.Fastest.CourierID, rates.Fastest.ServiceCode, rates.Fastest.ServiceType)
	}
	cheapestID := ""
	if rates.Cheapest != nil {
		cheapestID = quoteOptionID(rates.Cheapest.CourierID, rates.Cheapest.ServiceCode, rates.Cheapest.ServiceType)
	}

	options := make([]models.DispatchShipmentOption, 0, len(rates.Options))
	for _, option := range rates.Options {
		id := quoteOptionID(option.CourierID, option.ServiceCode, option.ServiceType)
		options = append(options, models.DispatchShipmentOption{
			ID:            id,
			CourierID:     option.CourierID,
			CourierName:   option.CourierName,
			ServiceCode:   option.ServiceCode,
			ServiceType:   option.ServiceType,
			Amount:        option.Total.String(),
			Currency:      option.Currency,
			PickupETA:     option.PickupETA,
			DeliveryETA:   option.DeliveryETA,
			TrackingLabel: option.Tracking.Label,
			TrackingLevel: option.TrackingLevel,
			IsFastest:     fastestID != "" && fastestID == id,
			IsCheapest:    cheapestID != "" && cheapestID == id,
			ProviderData:  option.Raw,
		})
	}

	return options
}

func (s *ShipmentService) buildPackageItems(ctx context.Context, tenantID uuid.UUID, order *models.Order) ([]shipbubble.PackageItem, []string, decimal.Decimal, error) {
	items, err := s.orders.ListItems(ctx, order.ID)
	if err != nil {
		return nil, nil, decimal.Zero, fmt.Errorf("list order items: %w", err)
	}
	if len(items) == 0 {
		amount := order.TotalAmount
		if amount.IsZero() {
			amount = order.TotalAmount.Add(order.ShippingFee)
		}
		return []shipbubble.PackageItem{{
			Name:        "Storefront order",
			Description: fmt.Sprintf("Order %s", order.ID.String()),
			UnitWeight:  decimal.RequireFromString("0.25"),
			UnitAmount:  amount.String(),
			Quantity:    "1",
		}}, []string{"accessories"}, decimal.RequireFromString("0.25"), nil
	}

	packageItems := make([]shipbubble.PackageItem, 0, len(items))
	productCategories := make([]string, 0, len(items))
	totalWeight := decimal.Zero

	for _, item := range items {
		variant, err := s.products.GetVariantByID(ctx, item.VariantID)
		if err != nil {
			return nil, nil, decimal.Zero, fmt.Errorf("get order variant: %w", err)
		}
		product, err := s.products.GetByID(ctx, tenantID, variant.ProductID)
		if err != nil {
			return nil, nil, decimal.Zero, fmt.Errorf("get order product: %w", err)
		}

		category := strings.TrimSpace(derefString(product.Category))
		weight := estimateUnitWeight(category)
		totalWeight = totalWeight.Add(weight.Mul(decimal.NewFromInt(int64(item.Quantity))))
		productCategories = append(productCategories, category)

		packageItems = append(packageItems, shipbubble.PackageItem{
			Name:        derefString(item.ProductName),
			Description: strings.TrimSpace(derefString(product.Description)),
			UnitWeight:  weight,
			UnitAmount:  item.PriceAtSale.String(),
			Quantity:    fmt.Sprintf("%d", item.Quantity),
		})
	}

	if totalWeight.IsZero() {
		totalWeight = decimal.RequireFromString("0.25")
	}

	return packageItems, productCategories, totalWeight, nil
}

func (s *ShipmentService) handleDelivered(ctx context.Context, order *models.Order, shipment *models.Shipment) error {
	if err := s.shipments.UpdateStatus(ctx, order.TenantID, shipment.ID, models.ShipmentStatusDelivered); err != nil {
		return fmt.Errorf("update shipment status: %w", err)
	}
	if order.FulfillmentStatus == models.FulfillmentStatusDelivered {
		return nil
	}
	if err := s.orders.UpdateFulfillmentStatus(ctx, order.TenantID, order.ID, models.FulfillmentStatusDelivered); err != nil {
		return fmt.Errorf("update fulfillment status: %w", err)
	}
	if order.PaymentMethod != models.PaymentMethodOnline || s.walletSvc == nil {
		return nil
	}
	amount := order.TotalAmount.Add(order.ShippingFee)
	if err := s.walletSvc.ReleasePending(ctx, order.TenantID, amount, &order.ID); err != nil {
		return fmt.Errorf("release pending: %w", err)
	}
	return nil
}

func (s *ShipmentService) handleFailed(ctx context.Context, order *models.Order, shipment *models.Shipment) error {
	if order.FulfillmentStatus == models.FulfillmentStatusDelivered {
		return nil
	}
	if err := s.shipments.UpdateStatus(ctx, order.TenantID, shipment.ID, models.ShipmentStatusFailed); err != nil {
		return fmt.Errorf("update shipment status: %w", err)
	}
	if err := s.orders.UpdateFulfillmentStatus(ctx, order.TenantID, order.ID, models.FulfillmentStatusProcessing); err != nil {
		return fmt.Errorf("reset fulfillment status: %w", err)
	}
	return nil
}

func (s *ShipmentService) handleInFlight(ctx context.Context, order *models.Order, shipment *models.Shipment, status models.ShipmentStatus) error {
	if order.FulfillmentStatus == models.FulfillmentStatusDelivered {
		return nil
	}
	if err := s.shipments.UpdateStatus(ctx, order.TenantID, shipment.ID, status); err != nil {
		return fmt.Errorf("update shipment status: %w", err)
	}
	if order.FulfillmentStatus != models.FulfillmentStatusShipped {
		if err := s.orders.UpdateFulfillmentStatus(ctx, order.TenantID, order.ID, models.FulfillmentStatusShipped); err != nil {
			return fmt.Errorf("update fulfillment status: %w", err)
		}
	}
	return nil
}

func pickRateOption(options []shipbubble.RateOption, req DispatchShipmentRequest) *shipbubble.RateOption {
	courierID := strings.TrimSpace(req.CourierID)
	serviceCode := strings.TrimSpace(req.ServiceCode)
	serviceType := strings.TrimSpace(req.ServiceType)
	for i := range options {
		option := &options[i]
		if option.CourierID != courierID || option.ServiceCode != serviceCode {
			continue
		}
		if serviceType != "" && option.ServiceType != serviceType {
			continue
		}
		return option
	}
	return nil
}

func normalizeShipbubbleStatus(status string) models.ShipmentStatus {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "picked_up":
		return models.ShipmentStatusPickedUp
	case "in_transit":
		return models.ShipmentStatusInTransit
	case "completed":
		return models.ShipmentStatusDelivered
	case "cancelled":
		return models.ShipmentStatusFailed
	default:
		return models.ShipmentStatusQueued
	}
}

func mapShipmentAddressValidationError(role string, err error) error {
	return mapShipbubbleValidationError(
		role,
		"the store pickup address is incomplete. Update the logistics setup with a clear street, city, state, and country before dispatching orders",
		"the customer delivery address is incomplete. Update the order with a clear street, city, state, and country before dispatching it",
		err,
	)
}

func marshalCarrierHistory(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return json.RawMessage("[]"), nil
	}
	return json.Marshal([]json.RawMessage{raw})
}

func nullableTrimmedString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
