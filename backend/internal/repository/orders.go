package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/db"
	"storefront/backend/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, o *models.Order, items []models.OrderItem) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Order, error)
	GetByIDInternal(ctx context.Context, id uuid.UUID) (*models.Order, error)
	GetByTrackingSlug(ctx context.Context, slug string) (*models.Order, error)
	GetByPublicCheckoutID(ctx context.Context, tenantID, checkoutID uuid.UUID) (*models.Order, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]models.Order, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	UpdatePaymentStatus(ctx context.Context, tenantID, id uuid.UUID, status models.PaymentStatus) error
	UpdateFulfillmentStatus(ctx context.Context, tenantID, id uuid.UUID, status models.FulfillmentStatus) error
	ListItems(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error)
	WithTx(tx db.DBTX) OrderRepository
}

type orderRepo struct {
	db   db.DBTX
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) OrderRepository {
	return &orderRepo{db: pool, pool: pool}
}

func (r *orderRepo) WithTx(tx db.DBTX) OrderRepository {
	return &orderRepo{db: tx, pool: r.pool}
}

const orderCols = `id, tenant_id, tracking_slug, public_checkout_id, is_delivery, customer_name,
		customer_phone, customer_email, shipping_address, note,
		total_amount, shipping_fee, payment_method, payment_status, fulfillment_status,
		created_at, updated_at`

func scanOrder(row interface{ Scan(...any) error }) (*models.Order, error) {
	o := &models.Order{}
	err := row.Scan(
		&o.ID, &o.TenantID, &o.TrackingSlug, &o.PublicCheckoutID, &o.IsDelivery, &o.CustomerName,
		&o.CustomerPhone, &o.CustomerEmail, &o.ShippingAddress, &o.Note,
		&o.TotalAmount, &o.ShippingFee, &o.PaymentMethod, &o.PaymentStatus, &o.FulfillmentStatus,
		&o.CreatedAt, &o.UpdatedAt,
	)
	return o, err
}

// Create inserts the order and its items in a single transaction.
func (r *orderRepo) Create(ctx context.Context, o *models.Order, items []models.OrderItem) error {
	type txStarter interface {
		Begin(ctx context.Context) (pgx.Tx, error)
	}

	starter, ok := r.db.(txStarter)
	if !ok {
		return fmt.Errorf("repository db does not support transactions")
	}

	tx, err := starter.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO orders
		  (tenant_id, tracking_slug, public_checkout_id, is_delivery, customer_name, customer_phone,
		   customer_email, shipping_address, note, total_amount, shipping_fee, payment_method,
		   payment_status, fulfillment_status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, payment_method, payment_status, fulfillment_status, created_at, updated_at`,
		o.TenantID, o.TrackingSlug, o.PublicCheckoutID, o.IsDelivery, o.CustomerName, o.CustomerPhone,
		o.CustomerEmail, o.ShippingAddress, o.Note, o.TotalAmount, o.ShippingFee, o.PaymentMethod,
		o.PaymentStatus, o.FulfillmentStatus,
	).Scan(&o.ID, &o.PaymentMethod, &o.PaymentStatus, &o.FulfillmentStatus, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return err
	}

	for i := range items {
		items[i].OrderID = o.ID
		if err := tx.QueryRow(ctx, `
			INSERT INTO order_items (order_id, variant_id, quantity, price_at_sale, cost_price_at_sale, product_name, variant_label)
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
			items[i].OrderID, items[i].VariantID, items[i].Quantity, items[i].PriceAtSale,
			items[i].CostPriceAtSale, items[i].ProductName, items[i].VariantLabel,
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

func (r *orderRepo) GetByIDInternalForUpdate(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	return scanOrder(r.db.QueryRow(ctx,
		`SELECT `+orderCols+` FROM orders WHERE id = $1 FOR UPDATE`, id))
}

func (r *orderRepo) GetByTrackingSlug(ctx context.Context, slug string) (*models.Order, error) {
	return scanOrder(r.db.QueryRow(ctx,
		`SELECT `+orderCols+` FROM orders WHERE tracking_slug = $1`, slug))
}

func (r *orderRepo) GetByPublicCheckoutID(ctx context.Context, tenantID, checkoutID uuid.UUID) (*models.Order, error) {
	return scanOrder(r.db.QueryRow(ctx,
		`SELECT `+orderCols+` FROM orders WHERE tenant_id = $1 AND public_checkout_id = $2 AND payment_status <> 'failed'`,
		tenantID, checkoutID,
	))
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

func (r *orderRepo) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE tenant_id = $1`, tenantID).Scan(&count)
	return count, err
}

func (r *orderRepo) UpdatePaymentStatus(ctx context.Context, tenantID, id uuid.UUID, status models.PaymentStatus) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE orders SET payment_status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`, status, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *orderRepo) UpdateFulfillmentStatus(ctx context.Context, tenantID, id uuid.UUID, status models.FulfillmentStatus) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE orders SET fulfillment_status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`, status, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *orderRepo) ListItems(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, order_id, variant_id, quantity, price_at_sale, cost_price_at_sale, product_name, variant_label
		FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.VariantID, &item.Quantity, &item.PriceAtSale,
			&item.CostPriceAtSale, &item.ProductName, &item.VariantLabel); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
