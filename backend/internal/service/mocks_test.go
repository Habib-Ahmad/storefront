package service_test

import (
	"context"

	"github.com/google/uuid"

	"storefront/backend/internal/db"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

// ── Tenant repo mock ─────────────────────────────────────────

type mockTenantRepo struct {
	tenant  *models.Tenant
	err     error
	updated *models.Tenant
}

func (m *mockTenantRepo) Create(_ context.Context, t *models.Tenant) error {
	t.ID = uuid.New()
	m.tenant = t
	return m.err
}
func (m *mockTenantRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockTenantRepo) GetBySlug(_ context.Context, _ string) (*models.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockTenantRepo) Update(_ context.Context, t *models.Tenant) error {
	m.updated = t
	return m.err
}
func (m *mockTenantRepo) SoftDelete(_ context.Context, _ uuid.UUID) error { return m.err }

// ── Tier repo mock ─────────────────────────────────────────

type mockTierRepo struct {
	tier *models.Tier
	err  error
}

func (m *mockTierRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Tier, error) {
	return m.tier, m.err
}
func (m *mockTierRepo) GetByName(_ context.Context, _ string) (*models.Tier, error) {
	return m.tier, m.err
}
func (m *mockTierRepo) List(_ context.Context) ([]models.Tier, error) { return nil, m.err }

// ── Wallet repo mock ──────────────────────────────────────────

type mockWalletRepo struct {
	wallet  *models.Wallet
	err     error
	updated *models.Wallet
}

func (m *mockWalletRepo) Create(_ context.Context, w *models.Wallet) error {
	w.ID = uuid.New()
	m.wallet = w
	return m.err
}
func (m *mockWalletRepo) GetByTenantID(_ context.Context, _ uuid.UUID) (*models.Wallet, error) {
	return m.wallet, m.err
}
func (m *mockWalletRepo) GetByTenantIDForUpdate(_ context.Context, _ uuid.UUID) (*models.Wallet, error) {
	return m.wallet, m.err
}
func (m *mockWalletRepo) UpdateBalances(_ context.Context, w *models.Wallet) error {
	m.updated = w
	return m.err
}
func (m *mockWalletRepo) WithTx(_ db.DBTX) repository.WalletRepository { return m }

// ── User repo mock ────────────────────────────────────────────

type mockUserRepo struct {
	user *models.User
	err  error
}

func (m *mockUserRepo) Create(_ context.Context, u *models.User) error {
	m.user = u
	return m.err
}
func (m *mockUserRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) GetByEmail(_ context.Context, _ uuid.UUID, _ string) (*models.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) ListByTenant(_ context.Context, _ uuid.UUID) ([]models.User, error) {
	return nil, m.err
}
func (m *mockUserRepo) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return m.err }
func (m *mockUserRepo) Update(_ context.Context, u *models.User) error {
	m.user = u
	return m.err
}

// ── Product repo mock ─────────────────────────────────────────

type mockProductRepo struct {
	product        *models.Product
	variant        *models.ProductVariant
	variantCreated *models.ProductVariant
	err            error
	restocked      map[uuid.UUID]int
}

func (m *mockProductRepo) Create(_ context.Context, p *models.Product) error {
	p.ID = uuid.New()
	m.product = p
	return m.err
}
func (m *mockProductRepo) GetByID(_ context.Context, _ uuid.UUID, id uuid.UUID) (*models.Product, error) {
	if m.product != nil {
		return m.product, m.err
	}
	return &models.Product{ID: id, IsAvailable: true}, m.err
}
func (m *mockProductRepo) ListByTenant(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Product, error) {
	return nil, m.err
}
func (m *mockProductRepo) CountByTenant(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, m.err
}
func (m *mockProductRepo) Update(_ context.Context, _ *models.Product) error  { return m.err }
func (m *mockProductRepo) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return m.err }
func (m *mockProductRepo) CreateVariant(_ context.Context, v *models.ProductVariant) error {
	v.ID = uuid.New()
	m.variantCreated = v
	return m.err
}
func (m *mockProductRepo) GetVariantByID(_ context.Context, _ uuid.UUID) (*models.ProductVariant, error) {
	return m.variant, m.err
}
func (m *mockProductRepo) ListVariants(_ context.Context, _ uuid.UUID) ([]models.ProductVariant, error) {
	return nil, m.err
}
func (m *mockProductRepo) UpdateVariant(_ context.Context, v *models.ProductVariant) error {
	m.variant = v
	return m.err
}
func (m *mockProductRepo) DecrementStock(_ context.Context, _ uuid.UUID, _ int) error {
	return m.err
}
func (m *mockProductRepo) RestoreStock(_ context.Context, id uuid.UUID, qty int) error {
	if m.restocked == nil {
		m.restocked = make(map[uuid.UUID]int)
	}
	m.restocked[id] += qty
	return m.err
}
func (m *mockProductRepo) SoftDeleteVariant(_ context.Context, _ uuid.UUID) error { return m.err }
func (m *mockProductRepo) AddImage(_ context.Context, img *models.ProductImage) error {
	img.ID = uuid.New()
	return m.err
}
func (m *mockProductRepo) ListImagesByProduct(_ context.Context, _ uuid.UUID) ([]models.ProductImage, error) {
	return nil, m.err
}
func (m *mockProductRepo) DeleteImage(_ context.Context, _ uuid.UUID) error { return m.err }
func (m *mockProductRepo) UpdateImage(_ context.Context, _ *models.ProductImage) error {
	return m.err
}

// ── Order repo mock ───────────────────────────────────────────

type mockOrderRepo struct {
	order             *models.Order
	items             []models.OrderItem
	err               error
	paymentStatus     models.PaymentStatus
	fulfillmentStatus models.FulfillmentStatus
}

func (m *mockOrderRepo) Create(_ context.Context, o *models.Order, items []models.OrderItem) error {
	o.ID = uuid.New()
	m.order = o
	m.items = items
	return m.err
}
func (m *mockOrderRepo) GetByID(_ context.Context, _, _ uuid.UUID) (*models.Order, error) {
	return m.order, m.err
}
func (m *mockOrderRepo) GetByIDInternal(_ context.Context, _ uuid.UUID) (*models.Order, error) {
	return m.order, m.err
}
func (m *mockOrderRepo) GetByTrackingSlug(_ context.Context, _ string) (*models.Order, error) {
	return m.order, m.err
}
func (m *mockOrderRepo) ListByTenant(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Order, error) {
	return nil, m.err
}
func (m *mockOrderRepo) CountByTenant(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, m.err
}
func (m *mockOrderRepo) UpdatePaymentStatus(_ context.Context, _ uuid.UUID, _ uuid.UUID, s models.PaymentStatus) error {
	m.paymentStatus = s
	if m.order != nil {
		m.order.PaymentStatus = s
	}
	return m.err
}
func (m *mockOrderRepo) UpdateFulfillmentStatus(_ context.Context, _ uuid.UUID, _ uuid.UUID, s models.FulfillmentStatus) error {
	m.fulfillmentStatus = s
	if m.order != nil {
		m.order.FulfillmentStatus = s
	}
	return m.err
}
func (m *mockOrderRepo) ListItems(_ context.Context, _ uuid.UUID) ([]models.OrderItem, error) {
	return m.items, m.err
}

// ── Transaction repo mock ─────────────────────────────────────

type mockTxRepo struct {
	txs        []models.Transaction
	latest     *models.Transaction
	err        error
	created    *models.Transaction
	allCreated []*models.Transaction
}

func (m *mockTxRepo) Create(_ context.Context, tx *models.Transaction) error {
	tx.ID = uuid.New()
	m.created = tx
	m.latest = tx
	m.allCreated = append(m.allCreated, tx)
	return m.err
}
func (m *mockTxRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Transaction, error) {
	return m.latest, m.err
}
func (m *mockTxRepo) ListByWallet(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Transaction, error) {
	return m.txs, m.err
}
func (m *mockTxRepo) CountByWallet(_ context.Context, _ uuid.UUID) (int, error) {
	return len(m.txs), m.err
}
func (m *mockTxRepo) ListByWalletAsc(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Transaction, error) {
	return m.txs, m.err
}
func (m *mockTxRepo) GetLatestByWallet(_ context.Context, _ uuid.UUID) (*models.Transaction, error) {
	return m.latest, m.err
}
func (m *mockTxRepo) WithTx(_ db.DBTX) repository.TransactionRepository { return m }

// ── Audit log repo mock ──────────────────────────────────────

type mockAuditLogRepo struct {
	created *models.AuditLog
	err     error
}

func (m *mockAuditLogRepo) Create(_ context.Context, l *models.AuditLog) error {
	m.created = l
	return m.err
}
