package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"storefront/backend/internal/service"
)

type StorefrontHandler struct {
	svc *service.StorefrontService
	log *slog.Logger
}

func NewStorefrontHandler(svc *service.StorefrontService, log *slog.Logger) *StorefrontHandler {
	return &StorefrontHandler{svc: svc, log: log}
}

func (h *StorefrontHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondErr(w, http.StatusBadRequest, "invalid storefront slug")
		return
	}

	catalog, err := h.svc.GetPublicBySlug(r.Context(), slug)
	if err != nil {
		handleErr(w, h.log, r, err)
		return
	}

	respond(w, http.StatusOK, catalog)
}

func (h *StorefrontHandler) GetPublicProduct(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondErr(w, http.StatusBadRequest, "invalid storefront slug")
		return
	}

	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.svc.GetPublicProductBySlug(r.Context(), slug, productID)
	if err != nil {
		handleErr(w, h.log, r, err)
		return
	}

	respond(w, http.StatusOK, product)
}
