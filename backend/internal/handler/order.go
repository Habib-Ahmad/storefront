package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

// paymentInitiator is satisfied by *service.PaymentService.
type paymentInitiator interface {
	InitiatePayment(ctx context.Context, order *models.Order, customerEmail, subaccountCode string) (string, error)
}

type OrderHandler struct {
	svc        *service.OrderService
	paymentSvc paymentInitiator
	log        *slog.Logger
}

func NewOrderHandler(svc *service.OrderService, paymentSvc paymentInitiator, log *slog.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, paymentSvc: paymentSvc, log: log}
}

// POST /orders
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())

	var req struct {
		IsDelivery      bool               `json:"is_delivery"`
		CustomerName    string             `json:"customer_name"    validate:"required"`
		CustomerPhone   *string            `json:"customer_phone"`
		CustomerEmail   *string            `json:"customer_email"`
		ShippingAddress *string            `json:"shipping_address"`
		ShippingFee     float64            `json:"shipping_fee"`
		Items           []models.OrderItem `json:"items"            validate:"required,min=1"`
	}
	if !decodeValid(w, r, &req) {
		return
	}

	shippingFee := decimal.NewFromFloat(req.ShippingFee)
	order := &models.Order{
		TenantID:        tenant.ID,
		IsDelivery:      req.IsDelivery,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		CustomerEmail:   req.CustomerEmail,
		ShippingAddress: req.ShippingAddress,
		ShippingFee:     shippingFee,
	}

	out, err := h.svc.Create(r.Context(), order, req.Items)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDeliveryFieldsMissing):
			respondErr(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, service.ErrSoldOut):
			respondErr(w, http.StatusConflict, err.Error())
		default:
			serverErr(w, h.log, r, err)
		}
		return
	}

	email := "guest@storefront.ng"
	if req.CustomerEmail != nil && *req.CustomerEmail != "" {
		email = *req.CustomerEmail
	}
	subaccount := ""
	if tenant.PaystackSubaccountID != nil {
		subaccount = *tenant.PaystackSubaccountID
	}
	authURL, err := h.paymentSvc.InitiatePayment(r.Context(), out, email, subaccount)
	if err != nil {
		h.log.Error("initiate payment", "order_id", out.ID, "error", err)
		authURL = ""
	}

	type orderCreateResp struct {
		*models.Order
		AuthorizationURL string `json:"authorization_url,omitempty"`
	}
	respond(w, http.StatusCreated, orderCreateResp{Order: out, AuthorizationURL: authURL})
}

// GET /orders/{id}
func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid order id")
		return
	}
	order, err := h.svc.GetByID(r.Context(), id)
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
	respond(w, http.StatusOK, orders)
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
		CustomerName      string                   `json:"customer_name"`
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
