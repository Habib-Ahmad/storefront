package handler

import (
	"log/slog"
	"net/http"

	"storefront/backend/internal/repository"
)

type TierHandler struct {
	tiers repository.TierRepository
	log   *slog.Logger
}

func NewTierHandler(tiers repository.TierRepository, log *slog.Logger) *TierHandler {
	return &TierHandler{tiers: tiers, log: log}
}

// GET /tiers
func (h *TierHandler) List(w http.ResponseWriter, r *http.Request) {
	tiers, err := h.tiers.List(r.Context())
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	respond(w, http.StatusOK, tiers)
}
