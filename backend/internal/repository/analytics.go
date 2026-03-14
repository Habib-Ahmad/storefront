package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type AnalyticsSummary struct {
	TotalRevenue    decimal.Decimal      `json:"total_revenue"`
	TotalCost       decimal.Decimal      `json:"total_cost"`
	TotalProfit     decimal.Decimal      `json:"total_profit"`
	OrderCount      int                  `json:"order_count"`
	AvgOrderValue   decimal.Decimal      `json:"avg_order_value"`
	ByPaymentMethod []PaymentMethodStats `json:"by_payment_method"`
	TopProducts     []TopProductStats    `json:"top_products"`
	Period          AnalyticsPeriod      `json:"period"`
}

type AnalyticsPeriod struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type PaymentMethodStats struct {
	Method  string          `json:"method"`
	Revenue decimal.Decimal `json:"revenue"`
	Count   int             `json:"count"`
}

type TopProductStats struct {
	ProductName  string          `json:"product_name"`
	QuantitySold int             `json:"quantity_sold"`
	Revenue      decimal.Decimal `json:"revenue"`
}

type AnalyticsRepository interface {
	Summary(ctx context.Context, tenantID uuid.UUID, from, to time.Time) (*AnalyticsSummary, error)
}

type analyticsRepo struct{ db *pgxpool.Pool }

func NewAnalyticsRepository(db *pgxpool.Pool) AnalyticsRepository {
	return &analyticsRepo{db: db}
}

func (r *analyticsRepo) Summary(ctx context.Context, tenantID uuid.UUID, from, to time.Time) (*AnalyticsSummary, error) {
	s := &AnalyticsSummary{Period: AnalyticsPeriod{From: from, To: to}}

	// Revenue, cost, profit, count, avg — only paid orders
	err := r.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(o.total_amount), 0),
			COALESCE(SUM(oi.cost_price_at_sale * oi.quantity), 0),
			COUNT(DISTINCT o.id),
			COALESCE(AVG(o.total_amount), 0)
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		WHERE o.tenant_id = $1
		  AND o.payment_status = 'paid'
		  AND o.created_at >= $2
		  AND o.created_at < $3`,
		tenantID, from, to,
	).Scan(&s.TotalRevenue, &s.TotalCost, &s.OrderCount, &s.AvgOrderValue)
	if err != nil {
		return nil, err
	}
	s.TotalProfit = s.TotalRevenue.Sub(s.TotalCost)

	// Breakdown by payment method
	rows, err := r.db.Query(ctx, `
		SELECT payment_method, COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM orders
		WHERE tenant_id = $1
		  AND payment_status = 'paid'
		  AND created_at >= $2
		  AND created_at < $3
		GROUP BY payment_method
		ORDER BY SUM(total_amount) DESC`,
		tenantID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pm PaymentMethodStats
		if err := rows.Scan(&pm.Method, &pm.Revenue, &pm.Count); err != nil {
			return nil, err
		}
		s.ByPaymentMethod = append(s.ByPaymentMethod, pm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Top 10 products by quantity sold
	rows2, err := r.db.Query(ctx, `
		SELECT oi.product_name, SUM(oi.quantity)::int, COALESCE(SUM(oi.price_at_sale * oi.quantity), 0)
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.tenant_id = $1
		  AND o.payment_status = 'paid'
		  AND o.created_at >= $2
		  AND o.created_at < $3
		GROUP BY oi.product_name
		ORDER BY SUM(oi.quantity) DESC
		LIMIT 10`,
		tenantID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var tp TopProductStats
		if err := rows2.Scan(&tp.ProductName, &tp.QuantitySold, &tp.Revenue); err != nil {
			return nil, err
		}
		s.TopProducts = append(s.TopProducts, tp)
	}
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	// Ensure non-nil slices in JSON
	if s.ByPaymentMethod == nil {
		s.ByPaymentMethod = []PaymentMethodStats{}
	}
	if s.TopProducts == nil {
		s.TopProducts = []TopProductStats{}
	}

	return s, nil
}
