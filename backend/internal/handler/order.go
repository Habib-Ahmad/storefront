package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// POST /orders
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())

	var req struct {
		IsDelivery      bool               `json:"is_delivery"`
		CustomerName    string             `json:"customer_name"`
		CustomerPhone   *string            `json:"customer_phone"`
		CustomerEmail   *string            `json:"customer_email"`
		ShippingAddress *string            `json:"shipping_address"`
		ShippingFee     float64            `json:"shipping_fee"`
		Items           []models.OrderItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CustomerName == "" {
		respondErr(w, http.StatusBadRequest, "customer_name is required")
		return
	}
	if len(req.Items) == 0 {
		respondErr(w, http.StatusBadRequest, "order must contain at least one item")
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
			respondErr(w, http.StatusInternalServerError, "create order failed")
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
