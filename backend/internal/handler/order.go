package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
	log *slog.Logger
}

func NewOrderHandler(svc *service.OrderService, log *slog.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, log: log}
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

	order := &models.Order{
		TenantID:        tenant.ID,
		IsDelivery:      req.IsDelivery,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		CustomerEmail:   req.CustomerEmail,
		ShippingAddress: req.ShippingAddress,
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
	respond(w, http.StatusCreated, out)
}

// GET /orders/{id}
func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid order id")
		return
	}
	// DB fetch wired in full handler pass
	respondErr(w, http.StatusNotImplemented, "not yet implemented")
}
