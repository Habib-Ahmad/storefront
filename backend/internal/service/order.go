package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/apperr"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var (
	ErrDeliveryFieldsMissing = apperr.Unprocessable("customer_phone and shipping_address are required for delivery orders")
	ErrOrderNotFound         = apperr.NotFound("order not found")
	ErrProductUnavailable    = apperr.Unprocessable("product is not available")
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
	orders    repository.OrderRepository
	products  repository.ProductRepository
	walletSvc *WalletService
	tenants   repository.TenantRepository
	tiers     repository.TierRepository
}

func NewOrderService(orders repository.OrderRepository, products repository.ProductRepository) *OrderService {
	return &OrderService{orders: orders, products: products}
}

func (s *OrderService) SetWalletService(w *WalletService)           { s.walletSvc = w }
func (s *OrderService) SetTenantRepo(t repository.TenantRepository) { s.tenants = t }
func (s *OrderService) SetTierRepo(t repository.TierRepository)     { s.tiers = t }

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

	type variantDecrement struct {
		variantID uuid.UUID
		quantity  int
	}
	var decrements []variantDecrement

	for i := range items {
		v, err := s.products.GetVariantByID(ctx, items[i].VariantID)
		if err != nil {
			return nil, fmt.Errorf("variant %s: %w", items[i].VariantID, ErrVariantNotFound)
		}

		product, err := s.products.GetByID(ctx, order.TenantID, v.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product for variant %s: %w", v.ID, ErrVariantNotFound)
		}
		if !product.IsAvailable {
			return nil, fmt.Errorf("variant %s: %w", v.ID, ErrProductUnavailable)
		}

		if v.StockQty != nil && (*v.StockQty == 0 || *v.StockQty < items[i].Quantity) {
			return nil, fmt.Errorf("variant %s: %w", v.ID, ErrSoldOut)
		}

		items[i].PriceAtSale = v.Price
		items[i].ProductName = &product.Name
		items[i].VariantLabel = &v.SKU

		if v.StockQty != nil {
			decrements = append(decrements, variantDecrement{variantID: v.ID, quantity: items[i].Quantity})
		}
	}

	for _, d := range decrements {
		if err := s.products.DecrementStock(ctx, d.variantID, d.quantity); err != nil {
			return nil, fmt.Errorf("decrement stock for variant %s: %w", d.variantID, err)
		}
	}

	if len(items) > 0 {
		var total decimal.Decimal
		for _, item := range items {
			total = total.Add(item.PriceAtSale.Mul(decimal.NewFromInt(int64(item.Quantity))))
		}
		order.TotalAmount = total
	}
	// When items is empty, order.TotalAmount is already set by the handler (quick sale).

	if order.PaymentMethod == models.PaymentMethodOnline || order.PaymentMethod == "" {
		order.PaymentStatus = models.PaymentStatusPending
	} else {
		order.PaymentStatus = models.PaymentStatusPaid
	}
	order.FulfillmentStatus = models.FulfillmentStatusProcessing

	if err := s.orders.Create(ctx, order, items); err != nil {
		return nil, fmt.Errorf("persist order: %w", err)
	}

	if order.PaymentMethod != models.PaymentMethodOnline && order.PaymentMethod != "" {
		if err := s.settleOffline(ctx, order); err != nil {
			return nil, fmt.Errorf("settle offline: %w", err)
		}
	}

	return order, nil
}

// settleOffline credits the tenant wallet for a cash/transfer sale.
// Funds go to available_balance (not pending) since payment is already received.
// Commission is deducted if tiers are configured.
func (s *OrderService) settleOffline(ctx context.Context, order *models.Order) error {
	if s.walletSvc == nil {
		return nil
	}

	amount := order.TotalAmount
	orderID := order.ID

	var commission decimal.Decimal
	if s.tenants != nil && s.tiers != nil {
		tenant, err := s.tenants.GetByID(ctx, order.TenantID)
		if err == nil {
			tier, err := s.tiers.GetByID(ctx, tenant.TierID)
			if err == nil && tier.CommissionRate.IsPositive() {
				commission = amount.Mul(tier.CommissionRate)
			}
		}
	}

	netAmount := amount.Sub(commission)
	if _, err := s.walletSvc.CreditAvailable(ctx, order.TenantID, netAmount, &orderID); err != nil {
		return fmt.Errorf("credit wallet: %w", err)
	}

	if commission.IsPositive() {
		if _, err := s.walletSvc.RecordCommission(ctx, order.TenantID, commission, &orderID); err != nil {
			return fmt.Errorf("record commission: %w", err)
		}
	}

	return nil
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
