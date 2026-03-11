package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var (
	ErrDeliveryFieldsMissing = errors.New("customer_phone and shipping_address are required for delivery orders")
	ErrOrderNotFound         = errors.New("order not found")
)

type OrderService struct {
	orders   repository.OrderRepository
	products repository.ProductRepository
}

func NewOrderService(orders repository.OrderRepository, products repository.ProductRepository) *OrderService {
	return &OrderService{orders: orders, products: products}
}

// Create validates and persists a new order with its line items.
// All orders are 100% prepaid — payment_status starts as "pending" until Paystack confirms.
func (s *OrderService) Create(ctx context.Context, order *models.Order, items []models.OrderItem) (*models.Order, error) {
	if err := validateDelivery(order); err != nil {
		return nil, err
	}

	// Validate stock and collect variant prices before persisting.
	for i := range items {
		v, err := s.products.GetVariantByID(ctx, items[i].VariantID)
		if err != nil {
			return nil, fmt.Errorf("variant %s: %w", items[i].VariantID, ErrVariantNotFound)
		}

		// nil = infinite; 0 = sold out; < qty = insufficient stock
		if v.StockQty != nil && (*v.StockQty == 0 || *v.StockQty < items[i].Quantity) {
			return nil, fmt.Errorf("variant %s: %w", v.ID, ErrSoldOut)
		}

		// Snapshot the price at sale time (spec: price_at_sale is immutable).
		items[i].PriceAtSale = v.Price
	}

	// Compute order total from snapshotted prices.
	var total decimal.Decimal
	for _, item := range items {
		total = total.Add(item.PriceAtSale.Mul(decimal.NewFromInt(int64(item.Quantity))))
	}
	order.TotalAmount = total

	order.PaymentStatus = models.PaymentStatusPending
	order.FulfillmentStatus = models.FulfillmentStatusProcessing

	if err := s.orders.Create(ctx, order, items); err != nil {
		return nil, fmt.Errorf("persist order: %w", err)
	}
	return order, nil
}

func validateDelivery(o *models.Order) error {
	if !o.IsDelivery {
		return nil
	}
	if o.CustomerPhone == nil || *o.CustomerPhone == "" {
		return ErrDeliveryFieldsMissing
	}
	if o.ShippingAddress == nil || *o.ShippingAddress == "" {
		return ErrDeliveryFieldsMissing
	}
	return nil
}
