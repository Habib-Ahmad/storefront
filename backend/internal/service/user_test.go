package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

func TestUserGetByID_Found(t *testing.T) {
	user := &models.User{ID: uuid.New(), Email: "a@b.com", Role: models.UserRoleAdmin}
	svc := service.NewUserService(&mockUserRepo{user: user})

	got, err := svc.GetByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != user.ID {
		t.Fatalf("expected user %s, got %s", user.ID, got.ID)
	}
}

func TestUserGetByID_NotFound(t *testing.T) {
	svc := service.NewUserService(&mockUserRepo{err: errors.New("no rows")})

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, service.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserUpdateProfile_HappyPath(t *testing.T) {
	user := &models.User{ID: uuid.New(), Email: "a@b.com", Role: models.UserRoleAdmin}
	repo := &mockUserRepo{user: user}
	svc := service.NewUserService(repo)

	first := "John"
	last := "Doe"
	phone := "+2348012345678"
	err := svc.UpdateProfile(context.Background(), user.ID, &first, &last, &phone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.user.FirstName == nil || *repo.user.FirstName != "John" {
		t.Fatal("first_name not updated")
	}
	if repo.user.LastName == nil || *repo.user.LastName != "Doe" {
		t.Fatal("last_name not updated")
	}
	if repo.user.Phone == nil || *repo.user.Phone != "+2348012345678" {
		t.Fatal("phone not updated")
	}
}

func TestUserUpdateProfile_NotFound(t *testing.T) {
	svc := service.NewUserService(&mockUserRepo{err: errors.New("no rows")})

	first := "John"
	err := svc.UpdateProfile(context.Background(), uuid.New(), &first, nil, nil)
	if !errors.Is(err, service.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
