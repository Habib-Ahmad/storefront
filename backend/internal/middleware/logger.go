package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// RequestLogger returns a chi-compatible middleware that logs every request
// with method, path, status, response size, duration, and request ID.
func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, r)

			status := ww.Status()
			args := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).String(),
				"request_id", chimw.GetReqID(r.Context()),
			}
			switch {
			case status >= 500:
				log.Error("request", args...)
			case status >= 400:
				log.Warn("request", args...)
			default:
				log.Info("request", args...)
			}
		})
	}
}
