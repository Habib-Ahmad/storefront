package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/models"
)

type TierRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tier, error)
	List(ctx context.Context) ([]models.Tier, error)
}

type tierRepo struct{ db *pgxpool.Pool }

func NewTierRepository(db *pgxpool.Pool) TierRepository {
	return &tierRepo{db: db}
}

func (r *tierRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Tier, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, debt_ceiling, commission_rate, created_at, updated_at
		FROM tiers WHERE id = $1`, id)

	t := &models.Tier{}
	err := row.Scan(&t.ID, &t.Name, &t.DebtCeiling, &t.CommissionRate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *tierRepo) List(ctx context.Context) ([]models.Tier, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, debt_ceiling, commission_rate, created_at, updated_at
		FROM tiers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []models.Tier
	for rows.Next() {
		var t models.Tier
		if err := rows.Scan(&t.ID, &t.Name, &t.DebtCeiling, &t.CommissionRate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tiers = append(tiers, t)
	}
	return tiers, rows.Err()
}
