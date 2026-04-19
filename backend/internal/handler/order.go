package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

// paymentInitiator is satisfied by *service.PaymentService.
type paymentInitiator interface {
	InitiatePayment(ctx context.Context, order *models.Order, customerEmail, subaccountCode, callbackURL string) (string, error)
	HandleChargeSuccess(ctx context.Context, reference string) error
	HandleChargeFailed(ctx context.Context, reference string) error
}

type dispatcher interface {
	Dispatch(ctx context.Context, orderID, tenantID uuid.UUID, req service.DispatchShipmentRequest) (*models.Shipment, error)
	QuoteDispatchOptions(ctx context.Context, orderID, tenantID uuid.UUID) ([]models.DispatchShipmentOption, error)
}

type publicDeliveryQuoter interface {
	QuotePublic(ctx context.Context, slug string, req models.PublicStorefrontDeliveryQuoteRequest) (*models.PublicStorefrontDeliveryQuoteResponse, error)
	ResolvePublicSelection(ctx context.Context, slug string, req models.PublicStorefrontDeliveryQuoteRequest, selection models.PublicStorefrontDeliveryQuoteSelection) (decimal.Decimal, error)
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
	IsDelivery      bool                                           `json:"is_delivery"`
	CustomerName    *string                                        `json:"customer_name"`
	CustomerPhone   string                                         `json:"customer_phone" validate:"required"`
	CustomerEmail   *string                                        `json:"customer_email" validate:"omitempty,email"`
	CheckoutID      string                                         `json:"checkout_id" validate:"required,uuid"`
	ShippingAddress *string                                        `json:"shipping_address"`
	DeliveryOption  *models.PublicStorefrontDeliveryQuoteSelection `json:"delivery_option"`
	Note            *string                                        `json:"note"`
	Items           []orderCreateItemRequest                       `json:"items" validate:"required,min=1,dive"`
}

type orderCreateResp struct {
	*models.Order
	AuthorizationURL string `json:"authorization_url,omitempty"`
}

type paymentResumeResp struct {
	AuthorizationURL string `json:"authorization_url"`
}

type trackingResp struct {
	TrackingSlug      string                   `json:"tracking_slug"`
	IsDelivery        bool                     `json:"is_delivery"`
	StorefrontSlug    string                   `json:"storefront_slug,omitempty"`
	CustomerName      *string                  `json:"customer_name,omitempty"`
	PaymentStatus     models.PaymentStatus     `json:"payment_status"`
	FulfillmentStatus models.FulfillmentStatus `json:"fulfillment_status"`
}

type OrderHandler struct {
	svc            *service.OrderService
	paymentSvc     paymentInitiator
	shipmentSvc    dispatcher
	deliveryQuotes publicDeliveryQuoter
	publicAppURL   string
	log            *slog.Logger
}

func NewOrderHandler(svc *service.OrderService, paymentSvc paymentInitiator, publicAppURL string, log *slog.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, paymentSvc: paymentSvc, publicAppURL: publicAppURL, log: log}
}

func (h *OrderHandler) SetDeliveryQuoteService(svc publicDeliveryQuoter) {
	h.deliveryQuotes = svc
}

func (h *OrderHandler) SetShipmentService(svc dispatcher) {
	h.shipmentSvc = svc
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

func paymentEmail(customerEmail *string) string {
	if customerEmail == nil {
		return "guest@storefront.ng"
	}
	trimmed := strings.TrimSpace(*customerEmail)
	if trimmed == "" {
		return "guest@storefront.ng"
	}
	return trimmed
}

func buildPublicPaymentCallbackURL(baseURL, trackingSlug string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	basePath := strings.TrimRight(parsed.Path, "/")
	parsed.Path = basePath + "/order/" + trackingSlug
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String()
}

func paystackSubaccount(tenant *models.Tenant) string {
	if tenant == nil || tenant.PaystackSubaccountID == nil {
		return ""
	}
	return *tenant.PaystackSubaccountID
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

func buildPublicDeliveryQuoteRequest(req publicOrderCreateRequest) models.PublicStorefrontDeliveryQuoteRequest {
	return models.PublicStorefrontDeliveryQuoteRequest{
		CustomerName:         fallbackPublicCustomerName(req.CustomerName),
		CustomerPhone:        strings.TrimSpace(req.CustomerPhone),
		CustomerEmail:        normalizeOptionalString(req.CustomerEmail),
		ShippingAddress:      strings.TrimSpace(derefOptionalString(req.ShippingAddress)),
		DeliveryInstructions: normalizeOptionalString(req.Note),
		Items:                buildPublicDeliveryQuoteItems(req.Items),
	}
}

func buildPublicDeliveryQuoteItems(items []orderCreateItemRequest) []models.PublicStorefrontDeliveryQuoteRequestItem {
	quoteItems := make([]models.PublicStorefrontDeliveryQuoteRequestItem, 0, len(items))
	for _, item := range items {
		variantID, err := uuid.Parse(item.VariantID)
		if err != nil {
			continue
		}
		quoteItems = append(quoteItems, models.PublicStorefrontDeliveryQuoteRequestItem{
			VariantID: variantID,
			Quantity:  item.Quantity,
		})
	}
	return quoteItems
}

func derefOptionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func fallbackPublicCustomerName(value *string) string {
	name := strings.TrimSpace(derefOptionalString(value))
	if name == "" {
		return "Guest customer"
	}
	return name
}

func buildPublicOrder(req publicOrderCreateRequest, checkoutID uuid.UUID, shippingFee decimal.Decimal) *models.Order {
	customerPhone := strings.TrimSpace(req.CustomerPhone)
	shippingAddress := normalizeOptionalString(req.ShippingAddress)
	if !req.IsDelivery {
		shippingAddress = nil
		shippingFee = decimal.Zero
	}

	return &models.Order{
		IsDelivery:       req.IsDelivery,
		PaymentMethod:    models.PaymentMethodOnline,
		PublicCheckoutID: &checkoutID,
		CustomerName:     normalizeOptionalString(req.CustomerName),
		CustomerPhone:    &customerPhone,
		CustomerEmail:    normalizeOptionalString(req.CustomerEmail),
		ShippingAddress:  shippingAddress,
		Note:             normalizeOptionalString(req.Note),
		ShippingFee:      shippingFee,
	}
}

func parseOrderListView(raw string) models.OrderListView {
	switch models.OrderListView(strings.TrimSpace(raw)) {
	case models.OrderListViewAll:
		return models.OrderListViewAll
	case models.OrderListViewCancelled:
		return models.OrderListViewCancelled
	case models.OrderListViewActive:
		return models.OrderListViewActive
	case models.OrderListViewActionable:
		return models.OrderListViewActionable
	default:
		return models.OrderListViewActionable
	}
}

func publicStorefrontFromTenant(tenant *models.Tenant) models.PublicStorefront {
	return service.PublicStorefrontFromTenant(tenant)
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

func trackingResponseFromOrder(order *models.Order) trackingResp {
	return trackingResp{
		TrackingSlug:      order.TrackingSlug,
		IsDelivery:        order.IsDelivery,
		CustomerName:      order.CustomerName,
		PaymentStatus:     order.PaymentStatus,
		FulfillmentStatus: order.FulfillmentStatus,
	}
}

func (h *OrderHandler) publicTrackingResponse(ctx context.Context, order *models.Order) trackingResp {
	resp := trackingResponseFromOrder(order)
	tenant, err := h.svc.GetTenantByID(ctx, order.TenantID)
	if err != nil {
		if h.log != nil {
			h.log.Warn("load tenant for public tracking", "tenant_id", order.TenantID, "error", err)
		}
		return resp
	}
	if tenant != nil {
		resp.StorefrontSlug = tenant.Slug
	}
	return resp
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
	if out.PaymentMethod == models.PaymentMethodOnline && h.paymentSvc != nil {
		subaccount := ""
		if tenant.PaystackSubaccountID != nil {
			subaccount = *tenant.PaystackSubaccountID
		}
		authURL, err = h.paymentSvc.InitiatePayment(r.Context(), out, paymentEmail(req.CustomerEmail), subaccount, "")
		if err != nil {
			h.log.Error("initiate payment", "order_id", out.ID, "error", err)
			if failErr := h.paymentSvc.HandleChargeFailed(r.Context(), out.ID.String()); failErr != nil {
				h.log.Error("rollback merchant payment init failure", "order_id", out.ID, "error", failErr)
			}
			respondErr(w, http.StatusBadGateway, "could not start payment")
			return
		}
	}

	respond(w, http.StatusCreated, orderCreateResp{Order: out, AuthorizationURL: authURL})
}

// POST /storefronts/{slug}/delivery-quotes
func (h *OrderHandler) QuotePublicDelivery(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondErr(w, http.StatusBadRequest, "invalid storefront slug")
		return
	}
	if h.deliveryQuotes == nil {
		serverErr(w, h.log, r, errors.New("delivery quote service not configured"))
		return
	}

	var req models.PublicStorefrontDeliveryQuoteRequest
	if !decodeValid(w, r, &req) {
		return
	}

	quotes, err := h.deliveryQuotes.QuotePublic(r.Context(), slug, req)
	if err != nil {
		var providerErr *service.PublicDeliveryQuoteProviderError
		if errors.As(err, &providerErr) {
			h.log.Warn("public delivery quotes unavailable",
				"slug", slug,
				"operation", providerErr.Operation,
				"error", providerErr.Err.Error(),
			)
			respondErr(w, http.StatusConflict, "delivery is temporarily unavailable right now. Try again later or choose pickup")
			return
		}
		handleErr(w, h.log, r, err)
		return
	}

	respond(w, http.StatusOK, quotes)
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
	if strings.TrimSpace(req.CustomerPhone) == "" {
		respondErr(w, http.StatusUnprocessableEntity, "customer_phone is required")
		return
	}
	if req.IsDelivery && strings.TrimSpace(derefOptionalString(req.ShippingAddress)) == "" {
		respondErr(w, http.StatusUnprocessableEntity, "shipping_address is required for delivery orders")
		return
	}
	checkoutID, err := uuid.Parse(req.CheckoutID)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid checkout_id")
		return
	}

	items, ok := buildOrderItems(w, req.Items)
	if !ok {
		return
	}

	shippingFee := decimal.Zero
	if req.IsDelivery {
		if h.deliveryQuotes == nil {
			serverErr(w, h.log, r, errors.New("delivery quote service not configured"))
			return
		}
		if req.DeliveryOption == nil {
			respondErr(w, http.StatusUnprocessableEntity, "delivery_option is required for delivery orders")
			return
		}
		shippingFee, err = h.deliveryQuotes.ResolvePublicSelection(r.Context(), slug, buildPublicDeliveryQuoteRequest(req), *req.DeliveryOption)
		if err != nil {
			handleErr(w, h.log, r, err)
			return
		}
	}

	tenant, order, reusedExisting, err := h.svc.CreatePublic(r.Context(), slug, buildPublicOrder(req, checkoutID, shippingFee), items)
	if err != nil {
		handleErr(w, h.log, r, err)
		return
	}

	var authURL string
	if !reusedExisting && order.PaymentMethod == models.PaymentMethodOnline && h.paymentSvc != nil {
		subaccount := ""
		if tenant.PaystackSubaccountID != nil {
			subaccount = *tenant.PaystackSubaccountID
		}

		authURL, err = h.paymentSvc.InitiatePayment(
			r.Context(),
			order,
			paymentEmail(order.CustomerEmail),
			subaccount,
			buildPublicPaymentCallbackURL(h.publicAppURL, order.TrackingSlug),
		)
		if err != nil {
			h.log.Error("initiate public payment", "order_id", order.ID, "error", err)
			if failErr := h.paymentSvc.HandleChargeFailed(r.Context(), order.ID.String()); failErr != nil {
				h.log.Error("rollback public payment init failure", "order_id", order.ID, "error", failErr)
			}
			respondErr(w, http.StatusBadGateway, "could not start payment")
			return
		}
	}

	status := http.StatusCreated
	if reusedExisting {
		status = http.StatusOK
	}

	respond(w, status, models.PublicStorefrontCheckoutResponse{
		Storefront:       publicStorefrontFromTenant(tenant),
		Order:            publicCheckoutOrderFromOrder(order),
		AuthorizationURL: authURL,
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

// GET /orders?limit=20&offset=0&view=actionable
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)
	view := parseOrderListView(r.URL.Query().Get("view"))
	orders, err := h.svc.List(r.Context(), tenant.ID, view, limit, offset)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	if orders == nil {
		orders = []models.Order{}
	}
	total, err := h.svc.CountByTenant(r.Context(), tenant.ID, view)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	respondPage(w, orders, total, limit, offset)
}

// GET /orders/{id}/dispatch-options
func (h *OrderHandler) DispatchOptions(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, false, false, true); err != nil {
		respondErr(w, http.StatusForbidden, "logistics module not enabled")
		return
	}
	if h.shipmentSvc == nil {
		serverErr(w, h.log, r, errors.New("shipment service not configured"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid order id")
		return
	}

	options, err := h.shipmentSvc.QuoteDispatchOptions(r.Context(), id, tenant.ID)
	if err != nil {
		handleErr(w, h.log, r, err)
		return
	}

	respond(w, http.StatusOK, options)
}

// POST /orders/{id}/dispatch
func (h *OrderHandler) Dispatch(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, false, false, true); err != nil {
		respondErr(w, http.StatusForbidden, "logistics module not enabled")
		return
	}
	if h.shipmentSvc == nil {
		serverErr(w, h.log, r, errors.New("shipment service not configured"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var req struct {
		CourierID   string `json:"courier_id" validate:"required"`
		ServiceCode string `json:"service_code" validate:"required"`
		ServiceType string `json:"service_type"`
	}
	if !decodeValid(w, r, &req) {
		return
	}

	shipment, err := h.shipmentSvc.Dispatch(r.Context(), id, tenant.ID, service.DispatchShipmentRequest{
		CourierID:   req.CourierID,
		ServiceCode: req.ServiceCode,
		ServiceType: req.ServiceType,
	})
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

// POST /orders/{id}/resume-payment
func (h *OrderHandler) ResumePayment(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.svc.GetByID(r.Context(), tenant.ID, id)
	if err != nil {
		respondErr(w, http.StatusNotFound, "order not found")
		return
	}
	if err := service.EnsurePaymentResumable(order); err != nil {
		handleErr(w, h.log, r, err)
		return
	}
	if h.paymentSvc == nil {
		serverErr(w, h.log, r, errors.New("payment service not configured"))
		return
	}

	authorizationURL, err := h.paymentSvc.InitiatePayment(
		r.Context(),
		order,
		paymentEmail(order.CustomerEmail),
		paystackSubaccount(tenant),
		"",
	)
	if err != nil {
		h.log.Error("resume merchant payment", "order_id", order.ID, "error", err)
		respondErr(w, http.StatusBadGateway, "could not start payment")
		return
	}

	respond(w, http.StatusOK, paymentResumeResp{AuthorizationURL: authorizationURL})
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
	respond(w, http.StatusOK, h.publicTrackingResponse(r.Context(), order))
}

// POST /track/{slug}/confirm-payment — public, no auth
func (h *OrderHandler) ConfirmPaymentPublic(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondErr(w, http.StatusBadRequest, "tracking slug required")
		return
	}
	if h.paymentSvc == nil {
		serverErr(w, h.log, r, errors.New("payment service not configured"))
		return
	}

	var req struct {
		Reference string `json:"reference"`
		Trxref    string `json:"trxref"`
	}
	if !decodeValid(w, r, &req) {
		return
	}
	reference := strings.TrimSpace(req.Reference)
	if reference == "" {
		reference = strings.TrimSpace(req.Trxref)
	}
	if reference == "" {
		respondErr(w, http.StatusBadRequest, "payment reference required")
		return
	}

	order, err := h.svc.GetByTrackingSlug(r.Context(), slug)
	if err != nil {
		respondErr(w, http.StatusNotFound, "order not found")
		return
	}
	if reference != order.ID.String() {
		respondErr(w, http.StatusBadRequest, "invalid payment reference")
		return
	}

	statusCode := http.StatusOK
	if order.PaymentStatus == models.PaymentStatusPending {
		if err := h.paymentSvc.HandleChargeSuccess(r.Context(), reference); err != nil {
			if errors.Is(err, service.ErrPaymentVerificationFailed) {
				statusCode = http.StatusAccepted
			} else {
				handleErr(w, h.log, r, err)
				return
			}
		}
	}

	refreshed, err := h.svc.GetByTrackingSlug(r.Context(), slug)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}

	respond(w, statusCode, h.publicTrackingResponse(r.Context(), refreshed))
}

// POST /track/{slug}/resume-payment — public, no auth
func (h *OrderHandler) ResumePaymentPublic(w http.ResponseWriter, r *http.Request) {
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
	if err := service.EnsurePaymentResumable(order); err != nil {
		handleErr(w, h.log, r, err)
		return
	}
	if h.paymentSvc == nil {
		serverErr(w, h.log, r, errors.New("payment service not configured"))
		return
	}

	tenant, err := h.svc.GetTenantByID(r.Context(), order.TenantID)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}

	authorizationURL, err := h.paymentSvc.InitiatePayment(
		r.Context(),
		order,
		paymentEmail(order.CustomerEmail),
		paystackSubaccount(tenant),
		buildPublicPaymentCallbackURL(h.publicAppURL, order.TrackingSlug),
	)
	if err != nil {
		h.log.Error("resume public payment", "order_id", order.ID, "error", err)
		respondErr(w, http.StatusBadGateway, "could not start payment")
		return
	}

	respond(w, http.StatusOK, paymentResumeResp{AuthorizationURL: authorizationURL})
}
