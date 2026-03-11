package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondErr(w http.ResponseWriter, status int, msg string) {
	respond(w, status, map[string]string{"error": msg})
}

// serverErr logs the real error at ERROR level then sends a generic 500 to the client.
// This prevents leaking internal details while keeping diagnostics in the logs.
func serverErr(w http.ResponseWriter, log *slog.Logger, r *http.Request, err error) {
	log.Error("internal error",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)
	respondErr(w, http.StatusInternalServerError, "internal server error")
}
