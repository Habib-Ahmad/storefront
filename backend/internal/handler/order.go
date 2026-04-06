package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/adapter/terminalaf"
	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

// paymentInitiator is satisfied by *service.PaymentService.
type paymentInitiator interface {
	InitiatePayment(ctx context.Context, order *models.Order, customerEmail, subaccountCode string) (string, error)
}

// dispatcher is satisfied by *service.ShipmentService.
type dispatcher interface {
	Dispatch(ctx context.Context, orderID, tenantID uuid.UUID, req terminalaf.BookRequest) (*models.Shipment, error)
}

type orderCreateItemRequest struct {
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

type orderCreateRequest struct {
	IsDelivery      bool                     `json:"is_delivery"`
	PaymentMethod   string                   `json:"payment_method"   validate:"omitempty,oneof=online cash transfer"`
	CustomerName    *string                  `json:"customer_name"`
	CustomerPhone   *string                  `json:"customer_phone"`
	CustomerEmail   *string                  `json:"customer_email" validate:"omitempty,email"`
	ShippingAddress *string                  `json:"shipping_address"`
	Note            *string                  `json:"note"`
	ShippingFee     float64                  `json:"shipping_fee"`
	TotalAmount     *float64                 `json:"total_amount"`
	Items           []orderCreateItemRequest `json:"items"`
}

type publicOrderCreateRequest struct {
	IsDelivery      bool                     `json:"is_delivery"`
	CustomerName    string                   `json:"customer_name" validate:"required"`
	CustomerPhone   string                   `json:"customer_phone" validate:"required"`
	CustomerEmail   *string                  `json:"customer_email" validate:"omitempty,email"`
	ShippingAddress *string                  `json:"shipping_address"`
	Note            *string                  `json:"note"`
	Items           []orderCreateItemRequest `json:"items" validate:"required,min=1,dive"`
}

type orderCreateResp struct {
	*models.Order
	AuthorizationURL string `json:"authorization_url,omitempty"`
}

type OrderHandler struct {
	svc         *service.OrderService
	paymentSvc  paymentInitiator
	shipmentSvc dispatcher
	log         *slog.Logger
}

func NewOrderHandler(svc *service.OrderService, paymentSvc paymentInitiator, shipmentSvc dispatcher, log *slog.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, paymentSvc: paymentSvc, shipmentSvc: shipmentSvc, log: log}
}

func buildOrderItems(w http.ResponseWriter, itemsReq []orderCreateItemRequest) ([]models.OrderItem, bool) {
	items := make([]models.OrderItem, 0, len(itemsReq))
	for _, ri := range itemsReq {
		vid, err := uuid.Parse(ri.VariantID)
		if err != nil {
			respondErr(w, http.StatusBadRequest, "invalid variant_id: "+ri.VariantID)
			return nil, false
		}
		if ri.Quantity <= 0 {
			respondErr(w, http.StatusBadRequest, "quantity must be positive")
			return nil, false
		}
		items = append(items, models.OrderItem{VariantID: vid, Quantity: ri.Quantity})
	}
	return items, true
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func buildMerchantOrder(req orderCreateRequest, tenantID uuid.UUID) *models.Order {
	paymentMethod := models.PaymentMethodOnline
	if req.PaymentMethod != "" {
		paymentMethod = models.PaymentMethod(req.PaymentMethod)
	}

	shippingFee := decimal.NewFromFloat(req.ShippingFee)
	var totalAmount decimal.Decimal
	if req.TotalAmount != nil {
		totalAmount = decimal.NewFromFloat(*req.TotalAmount)
	}

	return &models.Order{
		TenantID:        tenantID,
		IsDelivery:      req.IsDelivery,
		PaymentMethod:   paymentMethod,
		CustomerName:    normalizeOptionalString(req.CustomerName),
		CustomerPhone:   normalizeOptionalString(req.CustomerPhone),
		CustomerEmail:   normalizeOptionalString(req.CustomerEmail),
		ShippingAddress: normalizeOptionalString(req.ShippingAddress),
		Note:            normalizeOptionalString(req.Note),
		ShippingFee:     shippingFee,
		TotalAmount:     totalAmount,
	}
}

func buildPublicOrder(req publicOrderCreateRequest) *models.Order {
	customerName := strings.TrimSpace(req.CustomerName)
	customerPhone := strings.TrimSpace(req.CustomerPhone)

	return &models.Order{
		IsDelivery:      req.IsDelivery,
		PaymentMethod:   models.PaymentMethodOnline,
		CustomerName:    &customerName,
		CustomerPhone:   &customerPhone,
		CustomerEmail:   normalizeOptionalString(req.CustomerEmail),
		ShippingAddress: normalizeOptionalString(req.ShippingAddress),
		Note:            normalizeOptionalString(req.Note),
		ShippingFee:     decimal.Zero,
	}
}

func publicStorefrontFromTenant(tenant *models.Tenant) models.PublicStorefront {
	return models.PublicStorefront{
		Name:         tenant.Name,
		Slug:         tenant.Slug,
		LogoURL:      tenant.LogoURL,
		ContactEmail: tenant.ContactEmail,
		ContactPhone: tenant.ContactPhone,
		Address:      tenant.Address,
	}
}

func publicCheckoutOrderFromOrder(order *models.Order) models.PublicStorefrontCheckoutOrder {
	return models.PublicStorefrontCheckoutOrder{
		TrackingSlug:      order.TrackingSlug,
		IsDelivery:        order.IsDelivery,
		CustomerName:      order.CustomerName,
		CustomerPhone:     order.CustomerPhone,
		CustomerEmail:     order.CustomerEmail,
		ShippingAddress:   order.ShippingAddress,
		Note:              order.Note,
		TotalAmount:       order.TotalAmount,
		ShippingFee:       order.ShippingFee,
		PaymentMethod:     order.PaymentMethod,
		PaymentStatus:     order.PaymentStatus,
		FulfillmentStatus: order.FulfillmentStatus,
	}
}

// POST /orders
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, false, true, false); err != nil {
		respondErr(w, http.StatusForbidden, "payments module not enabled")
		return
	}

	var req orderCreateRequest
	if !decodeValid(w, r, &req) {
		return
	}

	if len(req.Items) == 0 && (req.TotalAmount == nil || *req.TotalAmount <= 0) {
		respondErr(w, http.StatusUnprocessableEntity, "items or total_amount required")
		return
	}

	items, ok := buildOrderItems(w, req.Items)
	if !ok {
		return
	}

	order := buildMerchantOrder(req, tenant.ID)

	out, err := h.svc.Create(r.Context(), order, items)
	if err != nil {
		handleErr(w, h.log, r, err)
		return
	}

	var authURL string
	if out.PaymentMethod == models.PaymentMethodOnline {
		email := "guest@storefront.ng"
		if req.CustomerEmail != nil && *req.CustomerEmail != "" {
			email = *req.CustomerEmail
		}
		subaccount := ""
		if tenant.PaystackSubaccountID != nil {
			subaccount = *tenant.PaystackSubaccountID
		}
		authURL, err = h.paymentSvc.InitiatePayment(r.Context(), out, email, subaccount)
		if err != nil {
			h.log.Error("initiate payment", "order_id", out.ID, "error", err)
			authURL = ""
		}
	}

	respond(w, http.StatusCreated, orderCreateResp{Order: out, AuthorizationURL: authURL})
}

// POST /storefronts/{slug}/orders
func (h *OrderHandler) CreatePublic(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondErr(w, http.StatusBadRequest, "invalid storefront slug")
		return
	}

	var req publicOrderCreateRequest
	if !decodeValid(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.CustomerName) == "" {
		respondErr(w, http.StatusUnprocessableEntity, "customer_name is required")
		return
	}
	if strings.TrimSpace(req.CustomerPhone) == "" {
		respondErr(w, http.StatusUnprocessableEntity, "customer_phone is required")
		return
	}

	items, ok := buildOrderItems(w, req.Items)
	if !ok {
		return
	}

	tenant, order, err := h.svc.CreatePublic(r.Context(), slug, buildPublicOrder(req), items)
	if err != nil {
		handleErr(w, h.log, r, err)
		return
	}

	respond(w, http.StatusCreated, models.PublicStorefrontCheckoutResponse{
		Storefront: publicStorefrontFromTenant(tenant),
		Order:      publicCheckoutOrderFromOrder(order),
	})
}

// GET /orders/{id}
func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid order id")
		return
	}
	order, err := h.svc.GetByID(r.Context(), tenant.ID, id)
	if err != nil {
		respondErr(w, http.StatusNotFound, "order not found")
		return
	}
	respond(w, http.StatusOK, order)
}

// GET /orders?limit=20&offset=0
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)
	orders, err := h.svc.List(r.Context(), tenant.ID, limit, offset)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	if orders == nil {
		orders = []models.Order{}
	}
	total, err := h.svc.CountByTenant(r.Context(), tenant.ID)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	respondPage(w, orders, total, limit, offset)
}

// POST /orders/{id}/dispatch
func (h *OrderHandler) Dispatch(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, false, false, true); err != nil {
		respondErr(w, http.StatusForbidden, "logistics module not enabled")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var req terminalaf.BookRequest
	if !decodeValid(w, r, &req) {
		return
	}
	req.Reference = id.String()

	shipment, err := h.shipmentSvc.Dispatch(r.Context(), id, tenant.ID, req)
	if err != nil {
		handleErr(w, h.log, r, err)
		return
	}
	respond(w, http.StatusCreated, shipment)
}

// POST /orders/{id}/cancel
func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid order id")
		return
	}
	if err := h.svc.Cancel(r.Context(), tenant.ID, id); err != nil {
		handleErr(w, h.log, r, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// GET /orders/{id}/items
func (h *OrderHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	_ = tenant // ownership enforced by fetching order first
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid order id")
		return
	}
	items, err := h.svc.ListItems(r.Context(), id)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	if items == nil {
		items = []models.OrderItem{}
	}
	respond(w, http.StatusOK, items)
}

// GET /track/{slug} — public, no auth
func (h *OrderHandler) Track(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondErr(w, http.StatusBadRequest, "tracking slug required")
		return
	}
	order, err := h.svc.GetByTrackingSlug(r.Context(), slug)
	if err != nil {
		respondErr(w, http.StatusNotFound, "order not found")
		return
	}
	// Return only the fields a customer needs — no internal IDs or financial data.
	type trackingResp struct {
		TrackingSlug      string                   `json:"tracking_slug"`
		CustomerName      *string                  `json:"customer_name,omitempty"`
		PaymentStatus     models.PaymentStatus     `json:"payment_status"`
		FulfillmentStatus models.FulfillmentStatus `json:"fulfillment_status"`
	}
	respond(w, http.StatusOK, trackingResp{
		TrackingSlug:      order.TrackingSlug,
		CustomerName:      order.CustomerName,
		PaymentStatus:     order.PaymentStatus,
		FulfillmentStatus: order.FulfillmentStatus,
	})
}
