package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"storefront/backend/internal/models"
)

// Authenticate validates a Supabase-issued HS256 JWT and injects the user ID
// and role into the request context. Rejects requests with missing or invalid tokens.
func Authenticate(secret string) func(http.Handler) http.Handler {
	keyFunc := func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := bearerToken(r)
			if tokenStr == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			var claims jwt.MapClaims
			_, err := jwt.ParseWithClaims(tokenStr, &claims, keyFunc,
				jwt.WithValidMethods([]string{"HS256"}),
				jwt.WithExpirationRequired(),
			)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			sub, err := claims.GetSubject()
			if err != nil || sub == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			role := models.UserRoleStaff
			if r, ok := claims["role"].(string); ok && r != "" {
				role = models.UserRole(r)
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
			ctx = context.WithValue(ctx, ctxKeyUserRole, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}
