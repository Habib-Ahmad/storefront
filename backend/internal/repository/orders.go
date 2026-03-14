package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, o *models.Order, items []models.OrderItem) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Order, error)
	GetByIDInternal(ctx context.Context, id uuid.UUID) (*models.Order, error)
	GetByTrackingSlug(ctx context.Context, slug string) (*models.Order, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]models.Order, error)
	UpdatePaymentStatus(ctx context.Context, tenantID, id uuid.UUID, status models.PaymentStatus) error
	UpdateFulfillmentStatus(ctx context.Context, tenantID, id uuid.UUID, status models.FulfillmentStatus) error
	ListItems(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error)
}

type orderRepo struct{ db *pgxpool.Pool }

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return &orderRepo{db: db}
}

const orderCols = `id, tenant_id, tracking_slug, is_delivery, customer_name,
		customer_phone, customer_email, shipping_address,
		total_amount, shipping_fee, payment_method, payment_status, fulfillment_status,
		created_at, updated_at`

func scanOrder(row interface{ Scan(...any) error }) (*models.Order, error) {
	o := &models.Order{}
	err := row.Scan(
		&o.ID, &o.TenantID, &o.TrackingSlug, &o.IsDelivery, &o.CustomerName,
		&o.CustomerPhone, &o.CustomerEmail, &o.ShippingAddress,
		&o.TotalAmount, &o.ShippingFee, &o.PaymentMethod, &o.PaymentStatus, &o.FulfillmentStatus,
		&o.CreatedAt, &o.UpdatedAt,
	)
	return o, err
}

// Create inserts the order and its items in a single transaction.
func (r *orderRepo) Create(ctx context.Context, o *models.Order, items []models.OrderItem) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO orders
		  (tenant_id, tracking_slug, is_delivery, customer_name, customer_phone,
		   customer_email, shipping_address, total_amount, shipping_fee, payment_method)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, payment_method, payment_status, fulfillment_status, created_at, updated_at`,
		o.TenantID, o.TrackingSlug, o.IsDelivery, o.CustomerName, o.CustomerPhone,
		o.CustomerEmail, o.ShippingAddress, o.TotalAmount, o.ShippingFee, o.PaymentMethod,
	).Scan(&o.ID, &o.PaymentMethod, &o.PaymentStatus, &o.FulfillmentStatus, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return err
	}

	for i := range items {
		items[i].OrderID = o.ID
		if err := tx.QueryRow(ctx, `
			INSERT INTO order_items (order_id, variant_id, quantity, price_at_sale)
			VALUES ($1, $2, $3, $4) RETURNING id`,
			items[i].OrderID, items[i].VariantID, items[i].Quantity, items[i].PriceAtSale,
		).Scan(&items[i].ID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *orderRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Order, error) {
	return scanOrder(r.db.QueryRow(ctx,
		`SELECT `+orderCols+` FROM orders WHERE id = $1 AND tenant_id = $2`, id, tenantID))
}

// GetByIDInternal fetches an order by ID without tenant filtering.
// Only for trusted internal paths (verified webhooks) where tenant_id is unknown.
func (r *orderRepo) GetByIDInternal(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	return scanOrder(r.db.QueryRow(ctx,
		`SELECT `+orderCols+` FROM orders WHERE id = $1`, id))
}

func (r *orderRepo) GetByTrackingSlug(ctx context.Context, slug string) (*models.Order, error) {
	return scanOrder(r.db.QueryRow(ctx,
		`SELECT `+orderCols+` FROM orders WHERE tracking_slug = $1`, slug))
}

func (r *orderRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]models.Order, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+orderCols+` FROM orders WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *o)
	}
	return orders, rows.Err()
}

func (r *orderRepo) UpdatePaymentStatus(ctx context.Context, tenantID, id uuid.UUID, status models.PaymentStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders SET payment_status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`, status, id, tenantID)
	return err
}

func (r *orderRepo) UpdateFulfillmentStatus(ctx context.Context, tenantID, id uuid.UUID, status models.FulfillmentStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders SET fulfillment_status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`, status, id, tenantID)
	return err
}

func (r *orderRepo) ListItems(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, order_id, variant_id, quantity, price_at_sale
		FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.VariantID, &item.Quantity, &item.PriceAtSale); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
