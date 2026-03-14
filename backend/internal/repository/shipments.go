package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/models"
)

type ShipmentRepository interface {
	Create(ctx context.Context, s *models.Shipment) error
	GetByOrderID(ctx context.Context, tenantID, orderID uuid.UUID) (*models.Shipment, error)
	UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status models.ShipmentStatus) error
	AppendCarrierEvent(ctx context.Context, tenantID, id uuid.UUID, event []byte) error
}

type shipmentRepo struct{ db *pgxpool.Pool }

func NewShipmentRepository(db *pgxpool.Pool) ShipmentRepository {
	return &shipmentRepo{db: db}
}

func (r *shipmentRepo) Create(ctx context.Context, s *models.Shipment) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO shipments (order_id, tenant_id, carrier_ref, tracking_number, carrier_history)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, created_at, updated_at`,
		s.OrderID, s.TenantID, s.CarrierRef, s.TrackingNumber, s.CarrierHistory,
	).Scan(&s.ID, &s.Status, &s.CreatedAt, &s.UpdatedAt)
}

func (r *shipmentRepo) GetByOrderID(ctx context.Context, tenantID, orderID uuid.UUID) (*models.Shipment, error) {
	s := &models.Shipment{}
	err := r.db.QueryRow(ctx, `
		SELECT id, order_id, tenant_id, status, carrier_ref, tracking_number, carrier_history, created_at, updated_at
		FROM shipments WHERE order_id = $1 AND tenant_id = $2`, orderID, tenantID,
	).Scan(&s.ID, &s.OrderID, &s.TenantID, &s.Status, &s.CarrierRef,
		&s.TrackingNumber, &s.CarrierHistory, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *shipmentRepo) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status models.ShipmentStatus) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE shipments SET status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`, status, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// AppendCarrierEvent appends a raw JSON event to the carrier_history array.
func (r *shipmentRepo) AppendCarrierEvent(ctx context.Context, tenantID, id uuid.UUID, event []byte) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE shipments
		SET carrier_history = carrier_history || $1::jsonb, updated_at = NOW()
		WHERE id = $2 AND tenant_id = $3`,
		event, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
