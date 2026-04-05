package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

func TestOrderRepositoryCreatePersistsStatusesAndSnapshots(t *testing.T) {
	ctx := context.Background()
	pool := setupRepositoryTestDB(t)
	tenantID := createTenantFixture(t, ctx, pool, models.TenantStatusActive)
	price := decimal.NewFromInt(3200)
	costPrice := decimal.NewFromInt(1400)
	variantID := createProductVariantFixture(t, ctx, pool, tenantID, price, &costPrice)

	repo := repository.NewOrderRepository(pool)
	customerName := "Ada"
	note := "Customer paid in store"
	productName := "Braided Sandals"
	variantLabel := "Size 39"

	order := &models.Order{
		TenantID:          tenantID,
		TrackingSlug:      "ord-" + uuid.NewString(),
		IsDelivery:        false,
		CustomerName:      &customerName,
		Note:              &note,
		TotalAmount:       price,
		ShippingFee:       decimal.Zero,
		PaymentMethod:     models.PaymentMethodTransfer,
		PaymentStatus:     models.PaymentStatusPaid,
		FulfillmentStatus: models.FulfillmentStatusCompleted,
	}
	items := []models.OrderItem{{
		VariantID:       variantID,
		Quantity:        1,
		PriceAtSale:     price,
		CostPriceAtSale: &costPrice,
		ProductName:     &productName,
		VariantLabel:    &variantLabel,
	}}

	if err := repo.Create(ctx, order, items); err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.ID == uuid.Nil {
		t.Fatal("expected order id to be set")
	}
	if items[0].ID == uuid.Nil {
		t.Fatal("expected item id to be set")
	}

	persisted, err := repo.GetByID(ctx, tenantID, order.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if persisted.PaymentMethod != models.PaymentMethodTransfer {
		t.Fatalf("payment_method: want %s, got %s", models.PaymentMethodTransfer, persisted.PaymentMethod)
	}
	if persisted.PaymentStatus != models.PaymentStatusPaid {
		t.Fatalf("payment_status: want %s, got %s", models.PaymentStatusPaid, persisted.PaymentStatus)
	}
	if persisted.FulfillmentStatus != models.FulfillmentStatusCompleted {
		t.Fatalf("fulfillment_status: want %s, got %s", models.FulfillmentStatusCompleted, persisted.FulfillmentStatus)
	}
	if persisted.Note == nil || *persisted.Note != note {
		t.Fatalf("note: want %q, got %v", note, persisted.Note)
	}

	persistedItems, err := repo.ListItems(ctx, order.ID)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if len(persistedItems) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(persistedItems))
	}
	if persistedItems[0].CostPriceAtSale == nil || !persistedItems[0].CostPriceAtSale.Equal(costPrice) {
		t.Fatalf("cost_price_at_sale: want %s, got %v", costPrice, persistedItems[0].CostPriceAtSale)
	}
	if persistedItems[0].ProductName == nil || *persistedItems[0].ProductName != productName {
		t.Fatalf("product_name: want %q, got %v", productName, persistedItems[0].ProductName)
	}
	if persistedItems[0].VariantLabel == nil || *persistedItems[0].VariantLabel != variantLabel {
		t.Fatalf("variant_label: want %q, got %v", variantLabel, persistedItems[0].VariantLabel)
	}
}

func TestOrderRepositoryStatusUpdatesRespectTenantScope(t *testing.T) {
	ctx := context.Background()
	pool := setupRepositoryTestDB(t)
	tenantID := createTenantFixture(t, ctx, pool, models.TenantStatusActive)
	otherTenantID := createTenantFixture(t, ctx, pool, models.TenantStatusActive)
	variantID := createProductVariantFixture(t, ctx, pool, tenantID, decimal.NewFromInt(1800), nil)

	repo := repository.NewOrderRepository(pool)
	order := &models.Order{
		TenantID:          tenantID,
		TrackingSlug:      "ord-" + uuid.NewString(),
		CustomerName:      strPtr("Bola"),
		TotalAmount:       decimal.NewFromInt(1800),
		ShippingFee:       decimal.Zero,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPending,
		FulfillmentStatus: models.FulfillmentStatusProcessing,
	}

	if err := repo.Create(ctx, order, []models.OrderItem{{
		VariantID:   variantID,
		Quantity:    1,
		PriceAtSale: decimal.NewFromInt(1800),
	}}); err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := repo.UpdatePaymentStatus(ctx, tenantID, order.ID, models.PaymentStatusFailed); err != nil {
		t.Fatalf("update payment status: %v", err)
	}
	if err := repo.UpdateFulfillmentStatus(ctx, tenantID, order.ID, models.FulfillmentStatusCancelled); err != nil {
		t.Fatalf("update fulfillment status: %v", err)
	}

	persisted, err := repo.GetByID(ctx, tenantID, order.ID)
	if err != nil {
		t.Fatalf("get updated order: %v", err)
	}
	if persisted.PaymentStatus != models.PaymentStatusFailed {
		t.Fatalf("payment_status: want %s, got %s", models.PaymentStatusFailed, persisted.PaymentStatus)
	}
	if persisted.FulfillmentStatus != models.FulfillmentStatusCancelled {
		t.Fatalf("fulfillment_status: want %s, got %s", models.FulfillmentStatusCancelled, persisted.FulfillmentStatus)
	}

	err = repo.UpdatePaymentStatus(ctx, otherTenantID, order.ID, models.PaymentStatusPaid)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows for wrong-tenant payment update, got %v", err)
	}

	err = repo.UpdateFulfillmentStatus(ctx, otherTenantID, order.ID, models.FulfillmentStatusCompleted)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows for wrong-tenant fulfillment update, got %v", err)
	}

	count, err := repo.CountByTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("count orders by tenant: %v", err)
	}
	if count != 1 {
		t.Fatalf("count: want 1, got %d", count)
	}
}

func strPtr(value string) *string {
	return &value
}
