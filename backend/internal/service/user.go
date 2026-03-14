package service

import (
	"context"

	"github.com/google/uuid"

	"storefront/backend/internal/apperr"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var ErrUserNotFound = apperr.NotFound("user not found")

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

// GetByID returns the user for the given ID.
func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile updates the authenticated user's first name, last name, and phone.
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName, phone *string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}
	user.FirstName = firstName
	user.LastName = lastName
	user.Phone = phone
	return s.users.Update(ctx, user)
}
