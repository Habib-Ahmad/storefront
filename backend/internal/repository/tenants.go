package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/models"
)

type TenantRepository interface {
	Create(ctx context.Context, t *models.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	Update(ctx context.Context, t *models.Tenant) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type tenantRepo struct{ db *pgxpool.Pool }

func NewTenantRepository(db *pgxpool.Pool) TenantRepository {
	return &tenantRepo{db: db}
}

const tenantCols = `id, tier_id, name, slug, paystack_subaccount_id,
		active_modules, status, created_at, updated_at, deleted_at`

func scanTenant(row interface{ Scan(...any) error }) (*models.Tenant, error) {
	t := &models.Tenant{}
	var modulesJSON []byte
	err := row.Scan(
		&t.ID, &t.TierID, &t.Name, &t.Slug, &t.PaystackSubaccountID,
		&modulesJSON, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(modulesJSON, &t.ActiveModules); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *tenantRepo) Create(ctx context.Context, t *models.Tenant) error {
	modulesJSON, err := json.Marshal(t.ActiveModules)
	if err != nil {
		return err
	}
	return r.db.QueryRow(ctx, `
		INSERT INTO tenants (tier_id, name, slug, paystack_subaccount_id, active_modules, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`,
		t.TierID, t.Name, t.Slug, t.PaystackSubaccountID, modulesJSON, t.Status,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *tenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return scanTenant(r.db.QueryRow(ctx,
		`SELECT `+tenantCols+` FROM tenants WHERE id = $1 AND deleted_at IS NULL`, id))
}

func (r *tenantRepo) GetBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	return scanTenant(r.db.QueryRow(ctx,
		`SELECT `+tenantCols+` FROM tenants WHERE slug = $1 AND deleted_at IS NULL`, slug))
}

func (r *tenantRepo) Update(ctx context.Context, t *models.Tenant) error {
	modulesJSON, err := json.Marshal(t.ActiveModules)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		UPDATE tenants
		SET tier_id = $1, name = $2, paystack_subaccount_id = $3,
		    active_modules = $4, status = $5, updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL`,
		t.TierID, t.Name, t.PaystackSubaccountID, modulesJSON, t.Status, t.ID)
	return err
}

func (r *tenantRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenants SET deleted_at = NOW() WHERE id = $1`, id)
	return err
}
