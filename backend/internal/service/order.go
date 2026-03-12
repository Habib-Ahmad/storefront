package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var (
	ErrDeliveryFieldsMissing = errors.New("customer_phone and shipping_address are required for delivery orders")
	ErrOrderNotFound         = errors.New("order not found")
	ErrProductUnavailable    = errors.New("product is not available")
)

// generateTrackingSlug returns a 12-character lowercase hex string (48 bits of entropy).
func generateTrackingSlug() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate slug: %w", err)
	}
	return hex.EncodeToString(b), nil
}

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

	slug, err := generateTrackingSlug()
	if err != nil {
		return nil, err
	}
	order.TrackingSlug = slug

	// Validate stock and collect variant prices before persisting.
	type variantDecrement struct {
		variant  *models.ProductVariant
		quantity int
	}
	var decrements []variantDecrement

	for i := range items {
		v, err := s.products.GetVariantByID(ctx, items[i].VariantID)
		if err != nil {
			return nil, fmt.Errorf("variant %s: %w", items[i].VariantID, ErrVariantNotFound)
		}

		// Check parent product availability.
		product, err := s.products.GetByID(ctx, order.TenantID, v.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product for variant %s: %w", v.ID, ErrVariantNotFound)
		}
		if !product.IsAvailable {
			return nil, fmt.Errorf("variant %s: %w", v.ID, ErrProductUnavailable)
		}

		// nil = infinite; 0 = sold out; < qty = insufficient stock
		if v.StockQty != nil && (*v.StockQty == 0 || *v.StockQty < items[i].Quantity) {
			return nil, fmt.Errorf("variant %s: %w", v.ID, ErrSoldOut)
		}

		// Snapshot the price at sale time (spec: price_at_sale is immutable).
		items[i].PriceAtSale = v.Price

		if v.StockQty != nil {
			decrements = append(decrements, variantDecrement{variant: v, quantity: items[i].Quantity})
		}
	}

	// Decrement stock for all finite-stock variants.
	for _, d := range decrements {
		newQty := *d.variant.StockQty - d.quantity
		d.variant.StockQty = &newQty
		if err := s.products.UpdateVariant(ctx, d.variant); err != nil {
			return nil, fmt.Errorf("decrement stock for variant %s: %w", d.variant.ID, err)
		}
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

func (s *OrderService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Order, error) {
	return s.orders.GetByID(ctx, tenantID, id)
}

func (s *OrderService) GetByTrackingSlug(ctx context.Context, slug string) (*models.Order, error) {
	return s.orders.GetByTrackingSlug(ctx, slug)
}

func (s *OrderService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]models.Order, error) {
	return s.orders.ListByTenant(ctx, tenantID, limit, offset)
}
