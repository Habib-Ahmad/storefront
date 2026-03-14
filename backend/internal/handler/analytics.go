package handler

import (
	"log/slog"
	"net/http"
	"time"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/repository"
)

type AnalyticsHandler struct {
	repo repository.AnalyticsRepository
	log  *slog.Logger
}

func NewAnalyticsHandler(repo repository.AnalyticsRepository, log *slog.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{repo: repo, log: log}
}

// GET /analytics/summary?from=2025-01-01&to=2025-01-31
func (h *AnalyticsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := now

	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t.AddDate(0, 0, 1) // make "to" exclusive end-of-day
		}
	}

	summary, err := h.repo.Summary(r.Context(), tenant.ID, from, to)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	respond(w, http.StatusOK, summary)
}
