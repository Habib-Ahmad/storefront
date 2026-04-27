package middleware

import (
	"net/http"

	"storefront/backend/internal/authz"
)

func RequirePermission(permission authz.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := UserRoleFromCtx(r.Context())
			if !authz.HasPermission(role, permission) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
