package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/apperr"
	"storefront/backend/internal/db"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var (
	ErrDeliveryFieldsMissing = apperr.Unprocessable("customer_phone and shipping_address are required for delivery orders")
	ErrOrderNotFound         = apperr.NotFound("order not found")
	ErrProductUnavailable    = apperr.Unprocessable("product is not available")
	ErrOrderNotCancellable   = apperr.Conflict("only processing orders can be cancelled")
	ErrCheckoutUnavailable   = apperr.Forbidden("checkout unavailable")
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
	pool      TxBeginner
}

func NewOrderService(orders repository.OrderRepository, products repository.ProductRepository) *OrderService {
	return &OrderService{orders: orders, products: products}
}

func (s *OrderService) SetWalletService(w *WalletService)           { s.walletSvc = w }
func (s *OrderService) SetTenantRepo(t repository.TenantRepository) { s.tenants = t }
func (s *OrderService) SetTierRepo(t repository.TierRepository)     { s.tiers = t }
func (s *OrderService) SetPool(pool TxBeginner)                     { s.pool = pool }

func initialFulfillmentStatus(order *models.Order) models.FulfillmentStatus {
	if !order.IsDelivery && order.PaymentStatus == models.PaymentStatusPaid {
		return models.FulfillmentStatusCompleted
	}

	return models.FulfillmentStatusProcessing
}

func (s *OrderService) CreatePublic(ctx context.Context, slug string, order *models.Order, items []models.OrderItem) (*models.Tenant, *models.Order, bool, error) {
	if s.tenants == nil {
		return nil, nil, false, fmt.Errorf("tenant repository not configured")
	}

	tenant, err := s.tenants.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, false, ErrStorefrontNotFound
		}
		return nil, nil, false, fmt.Errorf("get storefront tenant: %w", err)
	}
	if tenant == nil || tenant.Status != models.TenantStatusActive || !tenant.StorefrontPublished {
		return nil, nil, false, ErrStorefrontNotFound
	}
	if !tenant.ActiveModules.Payments {
		return nil, nil, false, ErrCheckoutUnavailable
	}

	if order.PublicCheckoutID != nil {
		existingOrder, existingErr := s.orders.GetByPublicCheckoutID(ctx, tenant.ID, *order.PublicCheckoutID)
		switch {
		case existingErr == nil:
			return tenant, existingOrder, true, nil
		case !errors.Is(existingErr, pgx.ErrNoRows):
			return nil, nil, false, fmt.Errorf("get public checkout order: %w", existingErr)
		}
	}

	order.TenantID = tenant.ID
	order.PaymentMethod = models.PaymentMethodOnline

	out, err := s.Create(ctx, order, items)
	if err != nil {
		return nil, nil, false, err
	}

	return tenant, out, false, nil
}

// Create validates and persists a new order with its line items.
// All orders are 100% prepaid — payment_status starts as "pending" until Paystack confirms.
func (s *OrderService) Create(ctx context.Context, order *models.Order, items []models.OrderItem) (*models.Order, error) {
	if err := validateDelivery(order); err != nil {
		return nil, err
	}

	if s.pool != nil {
		return s.createTx(ctx, order, items)
	}

	return s.createWithRepos(ctx, order, items, s.orders, s.products, nil)
}

func (s *OrderService) createTx(ctx context.Context, order *models.Order, items []models.OrderItem) (*models.Order, error) {
	dbTx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer dbTx.Rollback(ctx) //nolint:errcheck

	out, err := s.createWithRepos(ctx, order, items, s.orders.WithTx(dbTx), s.products.WithTx(dbTx), dbTx)
	if err != nil {
		return nil, err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return out, nil
}

func (s *OrderService) createWithRepos(ctx context.Context, order *models.Order, items []models.OrderItem, orders repository.OrderRepository, products repository.ProductRepository, dbTx db.DBTX) (*models.Order, error) {

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
		v, err := products.GetVariantByID(ctx, items[i].VariantID)
		if err != nil {
			return nil, fmt.Errorf("variant %s: %w", items[i].VariantID, ErrVariantNotFound)
		}

		product, err := products.GetByID(ctx, order.TenantID, v.ProductID)
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
		items[i].CostPriceAtSale = v.CostPrice
		items[i].ProductName = &product.Name
		items[i].VariantLabel = &v.SKU

		if v.StockQty != nil {
			decrements = append(decrements, variantDecrement{variantID: v.ID, quantity: items[i].Quantity})
		}
	}

	for _, d := range decrements {
		if err := products.DecrementStock(ctx, d.variantID, d.quantity); err != nil {
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
	order.FulfillmentStatus = initialFulfillmentStatus(order)

	// Retry with a new slug on unique-constraint collision (unlikely but possible).
	const maxSlugRetries = 3
	for attempt := 0; attempt < maxSlugRetries; attempt++ {
		err = orders.Create(ctx, order, items)
		if err == nil {
			break
		}
		if !apperr.IsUniqueViolation(err) {
			return nil, fmt.Errorf("persist order: %w", err)
		}
		slug, slugErr := generateTrackingSlug()
		if slugErr != nil {
			return nil, slugErr
		}
		order.TrackingSlug = slug
	}
	if err != nil {
		return nil, fmt.Errorf("persist order after retries: %w", err)
	}

	if order.PaymentMethod != models.PaymentMethodOnline && order.PaymentMethod != "" {
		if err := s.settleOffline(ctx, dbTx, order); err != nil {
			return nil, fmt.Errorf("settle offline: %w", err)
		}
	}

	return order, nil
}

// settleOffline credits the tenant wallet for a cash/transfer sale.
// Funds go to available_balance (not pending) since payment is already received.
// Commission is deducted if tiers are configured.
func (s *OrderService) settleOffline(ctx context.Context, dbTx db.DBTX, order *models.Order) error {
	if s.walletSvc == nil {
		return nil
	}

	amount := order.TotalAmount
	orderID := order.ID

	var commission decimal.Decimal
	if s.tenants != nil && s.tiers != nil {
		tenantRepo := s.tenants
		if dbTx != nil {
			tenantRepo = tenantRepo.WithTx(dbTx)
		}

		tenant, err := tenantRepo.GetByID(ctx, order.TenantID)
		if err == nil {
			tier, err := s.tiers.GetByID(ctx, tenant.TierID)
			if err == nil && tier.CommissionRate.IsPositive() {
				commission = amount.Mul(tier.CommissionRate)
			}
		}
	}

	netAmount := amount.Sub(commission)
	if dbTx != nil {
		if _, err := s.walletSvc.CreditAvailableWithTx(ctx, dbTx, order.TenantID, netAmount, &orderID); err != nil {
			return fmt.Errorf("credit wallet: %w", err)
		}
	} else {
		if _, err := s.walletSvc.CreditAvailable(ctx, order.TenantID, netAmount, &orderID); err != nil {
			return fmt.Errorf("credit wallet: %w", err)
		}
	}

	if commission.IsPositive() {
		if dbTx != nil {
			if _, err := s.walletSvc.RecordCommissionWithTx(ctx, dbTx, order.TenantID, commission, &orderID); err != nil {
				return fmt.Errorf("record commission: %w", err)
			}
		} else {
			if _, err := s.walletSvc.RecordCommission(ctx, order.TenantID, commission, &orderID); err != nil {
				return fmt.Errorf("record commission: %w", err)
			}
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

func (s *OrderService) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return s.orders.CountByTenant(ctx, tenantID)
}

func (s *OrderService) ListItems(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error) {
	return s.orders.ListItems(ctx, orderID)
}

// Cancel cancels a processing order: marks it cancelled/refunded, restocks items, and refunds wallet.
func (s *OrderService) Cancel(ctx context.Context, tenantID, orderID uuid.UUID) error {
	order, err := s.orders.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return ErrOrderNotFound
	}
	if order.FulfillmentStatus != models.FulfillmentStatusProcessing {
		return ErrOrderNotCancellable
	}

	if err := s.orders.UpdateFulfillmentStatus(ctx, tenantID, orderID, models.FulfillmentStatusCancelled); err != nil {
		return fmt.Errorf("cancel fulfillment: %w", err)
	}

	items, err := s.orders.ListItems(ctx, orderID)
	if err != nil {
		return fmt.Errorf("list items for restock: %w", err)
	}
	for _, item := range items {
		_ = s.products.RestoreStock(ctx, item.VariantID, item.Quantity)
	}

	if order.PaymentStatus == models.PaymentStatusPaid {
		if err := s.orders.UpdatePaymentStatus(ctx, tenantID, orderID, models.PaymentStatusRefunded); err != nil {
			return fmt.Errorf("refund payment status: %w", err)
		}
		if s.walletSvc != nil {
			amount := order.TotalAmount.Add(order.ShippingFee)
			if _, err := s.walletSvc.Refund(ctx, tenantID, amount, &orderID); err != nil {
				return fmt.Errorf("refund wallet: %w", err)
			}
		}
	}

	return nil
}
