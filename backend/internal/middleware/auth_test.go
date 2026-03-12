package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"storefront/backend/internal/middleware"
)

const testSecret = "test-secret-at-least-32-bytes-long!!"

var testKeyFunc = middleware.NewKeyFunc(nil, testSecret)

func makeToken(t *testing.T, sub string, expiry time.Duration, secret string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(expiry).Unix(),
		"iat": time.Now().Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func TestAuthenticate_ValidToken(t *testing.T) {
	userID := uuid.New()
	token := makeToken(t, userID.String(), time.Hour, testSecret)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	var gotID uuid.UUID
	mw := middleware.Authenticate(testKeyFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = middleware.UserIDFromCtx(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotID != userID {
		t.Fatalf("expected userID %s, got %s", userID, gotID)
	}
}

func TestAuthenticate_MissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw := middleware.Authenticate(testKeyFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthenticate_ExpiredToken(t *testing.T) {
	token := makeToken(t, uuid.New().String(), -time.Hour, testSecret)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	mw := middleware.Authenticate(testKeyFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthenticate_TamperedToken(t *testing.T) {
	token := makeToken(t, uuid.New().String(), time.Hour, "wrong-secret-!!!!!!!!!!!!!!!!")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	mw := middleware.Authenticate(testKeyFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthenticate_InvalidSubject(t *testing.T) {
	token := makeToken(t, "not-a-uuid", time.Hour, testSecret)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	mw := middleware.Authenticate(testKeyFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
