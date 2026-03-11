package middleware

import (
	"net/http"

	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

// ResolveTenant loads the tenant for the authenticated user and injects it into ctx.
// Must run after Authenticate. Rejects suspended tenants.
func ResolveTenant(users repository.UserRepository, tenants repository.TenantRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserIDFromCtx(r.Context())
			if userID.String() == "00000000-0000-0000-0000-000000000000" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := users.GetByID(r.Context(), userID)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tenant, err := tenants.GetByID(r.Context(), user.TenantID)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if tenant.Status == models.TenantStatusSuspended {
				http.Error(w, "account suspended", http.StatusForbidden)
				return
			}

			ctx := r.Context()
			ctx = setTenant(ctx, tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
