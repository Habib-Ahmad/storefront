package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"storefront/backend/internal/apperr"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		if name != "" {
			return name
		}
		return fld.Name
	})
	return v
}

const maxRequestBody = 1 << 20 // 1 MB

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondErr(w http.ResponseWriter, status int, msg string) {
	respond(w, status, map[string]string{"error": msg})
}

// respondValidationErr returns a 422 with a map of field → message pairs.
func respondValidationErr(w http.ResponseWriter, err error) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		respondErr(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	fields := make(map[string]string, len(ve))
	for _, fe := range ve {
		fields[fe.Field()] = validationMsg(fe)
	}
	respond(w, http.StatusUnprocessableEntity, map[string]any{"errors": fields})
}

func validationMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "gte":
		return fmt.Sprintf("must be %s or greater", fe.Param())
	default:
		return fmt.Sprintf("failed %s validation", fe.Tag())
	}
}

// decodeValid decodes the JSON body into dst and validates struct tags.
// On failure it writes the appropriate error response and returns false.
func decodeValid(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	if err := validate.Struct(dst); err != nil {
		respondValidationErr(w, err)
		return false
	}
	return true
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

func handleErr(w http.ResponseWriter, log *slog.Logger, r *http.Request, err error) {
	if status, msg := apperr.HTTPError(err); status != 0 {
		respondErr(w, status, msg)
		return
	}
	serverErr(w, log, r, err)
}

type pageResponse struct {
	Data    any `json:"data"`
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

func respondPage(w http.ResponseWriter, data any, total, limit, offset int) {
	page := offset/limit + 1
	respond(w, http.StatusOK, pageResponse{Data: data, Total: total, Page: page, PerPage: limit})
}
