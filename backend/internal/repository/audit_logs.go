package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/models"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
}

type auditLogRepo struct{ db *pgxpool.Pool }

func NewAuditLogRepository(db *pgxpool.Pool) AuditLogRepository {
	return &auditLogRepo{db: db}
}

func (r *auditLogRepo) Create(ctx context.Context, l *models.AuditLog) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO system_audit_logs (tenant_id, user_id, action, diff, created_at)
		VALUES ($1, $2, $3, $4, NOW())`,
		l.TenantID, l.UserID, l.Action, l.Diff)
	return err
}
