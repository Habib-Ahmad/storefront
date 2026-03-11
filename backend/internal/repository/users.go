package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, u *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.User, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type userRepo struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func scanUser(row interface{ Scan(...any) error }) (*models.User, error) {
	u := &models.User{}
	err := row.Scan(&u.ID, &u.TenantID, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	return u, err
}

func (r *userRepo) Create(ctx context.Context, u *models.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, role)
		VALUES ($1, $2, $3, $4)`,
		u.ID, u.TenantID, u.Email, u.Role)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return scanUser(r.db.QueryRow(ctx, `
		SELECT id, tenant_id, email, role, created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`, id))
}

func (r *userRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	return scanUser(r.db.QueryRow(ctx, `
		SELECT id, tenant_id, email, role, created_at, updated_at, deleted_at
		FROM users WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL`,
		tenantID, email))
}

func (r *userRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, email, role, created_at, updated_at, deleted_at
		FROM users WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		u := models.User{}
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET deleted_at = NOW() WHERE id = $1`, id)
	return err
}
