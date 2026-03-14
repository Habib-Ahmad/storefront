package handler

import (
	"log/slog"
	"net/http"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/service"
)

type UserHandler struct {
	svc *service.UserService
	log *slog.Logger
}

func NewUserHandler(svc *service.UserService, log *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: log}
}

// GET /users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		respondErr(w, http.StatusNotFound, "user not found")
		return
	}
	respond(w, http.StatusOK, user)
}

// PUT /users/me
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	var req struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		Phone     *string `json:"phone"`
	}
	if !decodeValid(w, r, &req) {
		return
	}
	if err := h.svc.UpdateProfile(r.Context(), userID, req.FirstName, req.LastName, req.Phone); err != nil {
		handleErr(w, h.log, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
