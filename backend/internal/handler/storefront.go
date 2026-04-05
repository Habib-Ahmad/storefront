package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

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
