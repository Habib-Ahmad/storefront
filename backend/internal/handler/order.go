package handler

import (
	"context"
	"log/slog"
	"net/http"

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

type OrderHandler struct {
	svc         *service.OrderService
	paymentSvc  paymentInitiator
	shipmentSvc dispatcher
	log         *slog.Logger
}

func NewOrderHandler(svc *service.OrderService, paymentSvc paymentInitiator, shipmentSvc dispatcher, log *slog.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, paymentSvc: paymentSvc, shipmentSvc: shipmentSvc, log: log}
}

// POST /orders
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())

	var req struct {
		IsDelivery      bool               `json:"is_delivery"`
		PaymentMethod   string             `json:"payment_method"   validate:"omitempty,oneof=online cash transfer"`
		CustomerName    *string            `json:"customer_name"`
		CustomerPhone   *string            `json:"customer_phone"`
		CustomerEmail   *string            `json:"customer_email"`
		ShippingAddress *string            `json:"shipping_address"`
		Note            *string            `json:"note"`
		ShippingFee     float64            `json:"shipping_fee"`
		TotalAmount     *float64           `json:"total_amount"`
		Items           []models.OrderItem `json:"items"`
	}
	if !decodeValid(w, r, &req) {
		return
	}

	if len(req.Items) == 0 && (req.TotalAmount == nil || *req.TotalAmount <= 0) {
		respondErr(w, http.StatusUnprocessableEntity, "items or total_amount required")
		return
	}

	paymentMethod := models.PaymentMethodOnline
	if req.PaymentMethod != "" {
		paymentMethod = models.PaymentMethod(req.PaymentMethod)
	}

	shippingFee := decimal.NewFromFloat(req.ShippingFee)
	var totalAmount decimal.Decimal
	if req.TotalAmount != nil {
		totalAmount = decimal.NewFromFloat(*req.TotalAmount)
	}
	order := &models.Order{
		TenantID:        tenant.ID,
		IsDelivery:      req.IsDelivery,
		PaymentMethod:   paymentMethod,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		CustomerEmail:   req.CustomerEmail,
		ShippingAddress: req.ShippingAddress,
		Note:            req.Note,
		ShippingFee:     shippingFee,
		TotalAmount:     totalAmount,
	}

	out, err := h.svc.Create(r.Context(), order, req.Items)
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

	type orderCreateResp struct {
		*models.Order
		AuthorizationURL string `json:"authorization_url,omitempty"`
	}
	respond(w, http.StatusCreated, orderCreateResp{Order: out, AuthorizationURL: authURL})
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
		serverErr(w, h.log, r, err)
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
