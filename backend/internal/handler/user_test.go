package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"storefront/backend/internal/handler"
	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
	"storefront/backend/internal/service"
)

type userHandlerRepo struct {
	user    *models.User
	getErr  error
	updated *models.User
	updErr  error
}

func (m *userHandlerRepo) Create(_ context.Context, _ *models.User) error { return nil }
func (m *userHandlerRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.User, error) {
	return m.user, m.getErr
}
func (m *userHandlerRepo) GetByEmail(_ context.Context, _ uuid.UUID, _ string) (*models.User, error) {
	return nil, nil
}
func (m *userHandlerRepo) ListByTenant(_ context.Context, _ uuid.UUID) ([]models.User, error) {
	return nil, nil
}
func (m *userHandlerRepo) Update(_ context.Context, u *models.User) error {
	m.updated = u
	return m.updErr
}
func (m *userHandlerRepo) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return nil }

var _ repository.UserRepository = (*userHandlerRepo)(nil)

func TestUserGetMe_OK(t *testing.T) {
	userID := uuid.New()
	repo := &userHandlerRepo{user: &models.User{ID: userID, Email: "a@b.com", Role: models.UserRoleAdmin}}
	h := handler.NewUserHandler(service.NewUserService(repo), slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.GetMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got models.User
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if got.ID != userID {
		t.Fatalf("expected user %s, got %s", userID, got.ID)
	}
}

func TestUserGetMe_NotFound(t *testing.T) {
	repo := &userHandlerRepo{getErr: errors.New("not found")}
	h := handler.NewUserHandler(service.NewUserService(repo), slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()

	h.GetMe(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUserUpdateProfile_OK(t *testing.T) {
	userID := uuid.New()
	repo := &userHandlerRepo{user: &models.User{ID: userID, Email: "a@b.com"}}
	h := handler.NewUserHandler(service.NewUserService(repo), slog.Default())

	body, _ := json.Marshal(map[string]string{
		"first_name": "John",
		"last_name":  "Doe",
		"phone":      "+2348012345678",
	})
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if repo.updated == nil {
		t.Fatal("user not updated")
	}
	if repo.updated.FirstName == nil || *repo.updated.FirstName != "John" {
		t.Fatal("first_name not set")
	}
}

func TestUserUpdateProfile_EmptyBody(t *testing.T) {
	userID := uuid.New()
	repo := &userHandlerRepo{user: &models.User{ID: userID, Email: "a@b.com"}}
	h := handler.NewUserHandler(service.NewUserService(repo), slog.Default())

	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserUpdateProfile_UserNotFound(t *testing.T) {
	repo := &userHandlerRepo{getErr: errors.New("not found")}
	h := handler.NewUserHandler(service.NewUserService(repo), slog.Default())

	body, _ := json.Marshal(map[string]string{"first_name": "John"})
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
