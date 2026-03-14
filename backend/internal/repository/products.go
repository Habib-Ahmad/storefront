package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/models"
)

type ProductRepository interface {
	Create(ctx context.Context, p *models.Product) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.Product, error)
	Update(ctx context.Context, p *models.Product) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error

	CreateVariant(ctx context.Context, v *models.ProductVariant) error
	GetVariantByID(ctx context.Context, id uuid.UUID) (*models.ProductVariant, error)
	ListVariants(ctx context.Context, productID uuid.UUID) ([]models.ProductVariant, error)
	UpdateVariant(ctx context.Context, v *models.ProductVariant) error
	DecrementStock(ctx context.Context, variantID uuid.UUID, qty int) error
	SoftDeleteVariant(ctx context.Context, id uuid.UUID) error

	AddImage(ctx context.Context, img *models.ProductImage) error
	ListImagesByProduct(ctx context.Context, productID uuid.UUID) ([]models.ProductImage, error)
	UpdateImage(ctx context.Context, img *models.ProductImage) error
	DeleteImage(ctx context.Context, id uuid.UUID) error
}

type productRepo struct{ db *pgxpool.Pool }

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, p *models.Product) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO products (tenant_id, name, description, category, is_available)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`,
		p.TenantID, p.Name, p.Description, p.Category, p.IsAvailable,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *productRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error) {
	p := &models.Product{}
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, name, description, category, is_available, created_at, updated_at, deleted_at
		FROM products WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`, id, tenantID,
	).Scan(&p.ID, &p.TenantID, &p.Name, &p.Description, &p.Category,
		&p.IsAvailable, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *productRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.Product, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, name, description, category, is_available, created_at, updated_at, deleted_at
		FROM products WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Description, &p.Category,
			&p.IsAvailable, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *productRepo) Update(ctx context.Context, p *models.Product) error {
	_, err := r.db.Exec(ctx, `
		UPDATE products
		SET name = $1, description = $2, category = $3, is_available = $4, updated_at = NOW()
		WHERE id = $5 AND tenant_id = $6 AND deleted_at IS NULL`,
		p.Name, p.Description, p.Category, p.IsAvailable, p.ID, p.TenantID)
	return err
}

func (r *productRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE products SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *productRepo) CreateVariant(ctx context.Context, v *models.ProductVariant) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO product_variants (product_id, sku, attributes, price, cost_price, stock_qty, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		v.ProductID, v.SKU, v.Attributes, v.Price, v.CostPrice, v.StockQty, v.IsDefault,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
}

func (r *productRepo) GetVariantByID(ctx context.Context, id uuid.UUID) (*models.ProductVariant, error) {
	v := &models.ProductVariant{}
	err := r.db.QueryRow(ctx, `
		SELECT id, product_id, sku, attributes, price, cost_price, stock_qty, is_default, created_at, updated_at, deleted_at
		FROM product_variants WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&v.ID, &v.ProductID, &v.SKU, &v.Attributes, &v.Price, &v.CostPrice,
		&v.StockQty, &v.IsDefault, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *productRepo) ListVariants(ctx context.Context, productID uuid.UUID) ([]models.ProductVariant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, product_id, sku, attributes, price, cost_price, stock_qty, is_default, created_at, updated_at, deleted_at
		FROM product_variants WHERE product_id = $1 AND deleted_at IS NULL ORDER BY is_default DESC, created_at`,
		productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []models.ProductVariant
	for rows.Next() {
		var v models.ProductVariant
		if err := rows.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Attributes, &v.Price, &v.CostPrice,
			&v.StockQty, &v.IsDefault, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt); err != nil {
			return nil, err
		}
		variants = append(variants, v)
	}
	return variants, rows.Err()
}

func (r *productRepo) UpdateVariant(ctx context.Context, v *models.ProductVariant) error {
	_, err := r.db.Exec(ctx, `
		UPDATE product_variants
		SET sku = $1, attributes = $2, price = $3, cost_price = $4, stock_qty = $5, updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL`,
		v.SKU, v.Attributes, v.Price, v.CostPrice, v.StockQty, v.ID)
	return err
}

func (r *productRepo) DecrementStock(ctx context.Context, variantID uuid.UUID, qty int) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE product_variants
		SET stock_qty = stock_qty - $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND stock_qty >= $1`, qty, variantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient stock for variant %s", variantID)
	}
	return nil
}

func (r *productRepo) SoftDeleteVariant(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE product_variants SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *productRepo) AddImage(ctx context.Context, img *models.ProductImage) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO product_images (product_id, url, sort_order, is_primary)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		img.ProductID, img.URL, img.SortOrder, img.IsPrimary,
	).Scan(&img.ID, &img.CreatedAt)
}

func (r *productRepo) ListImagesByProduct(ctx context.Context, productID uuid.UUID) ([]models.ProductImage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, product_id, url, sort_order, is_primary, created_at
		FROM product_images WHERE product_id = $1
		ORDER BY sort_order, created_at`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.ProductImage
	for rows.Next() {
		var img models.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.URL, &img.SortOrder,
			&img.IsPrimary, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, rows.Err()
}

func (r *productRepo) DeleteImage(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM product_images WHERE id = $1`, id)
	return err
}

func (r *productRepo) UpdateImage(ctx context.Context, img *models.ProductImage) error {
	_, err := r.db.Exec(ctx, `
		UPDATE product_images
		SET url = $1, sort_order = $2, is_primary = $3
		WHERE id = $4`,
		img.URL, img.SortOrder, img.IsPrimary, img.ID)
	return err
}
