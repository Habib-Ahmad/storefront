package middleware

import (
	"context"
	"crypto/ecdsa"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"storefront/backend/internal/models"
)

// Authenticate validates a Supabase-issued JWT (ES256 or HS256) and injects
// the user ID and role into the request context.
func Authenticate(keyFunc jwt.Keyfunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := bearerToken(r)
			if tokenStr == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			var claims jwt.MapClaims
			_, err := jwt.ParseWithClaims(tokenStr, &claims, keyFunc,
				jwt.WithValidMethods([]string{"ES256", "HS256"}),
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

// NewKeyFunc builds a jwt.Keyfunc that supports ES256 (via ECDSA public key)
// and HS256 (via shared secret) for backward compatibility with legacy Supabase projects.
func NewKeyFunc(ecKey *ecdsa.PublicKey, hmacSecret string) jwt.Keyfunc {
	return func(t *jwt.Token) (any, error) {
		switch t.Method.(type) {
		case *jwt.SigningMethodECDSA:
			if ecKey == nil {
				return nil, jwt.ErrSignatureInvalid
			}
			return ecKey, nil
		case *jwt.SigningMethodHMAC:
			if hmacSecret == "" {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(hmacSecret), nil
		default:
			return nil, jwt.ErrSignatureInvalid
		}
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}
