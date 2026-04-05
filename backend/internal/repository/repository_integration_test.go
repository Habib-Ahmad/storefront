package repository_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/migrations"
)

const storefrontTestDatabaseURLEnv = "STOREFRONT_TEST_DATABASE_URL"

func setupRepositoryTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	baseURL := repositoryTestDatabaseURL(t)
	adminPool := newTestPool(t, baseURL, "")
	schemaName := "repo_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	migrationSearchPath := schemaName
	runtimeSearchPath := schemaName + ",public"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := adminPool.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA %s`, schemaName)); err != nil {
		adminPool.Close()
		t.Fatalf("create test schema: %v", err)
	}

	applyMigrations(t, baseURL, migrationSearchPath)

	pool := newTestPool(t, baseURL, runtimeSearchPath)
	t.Cleanup(func() {
		pool.Close()

		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()

		if _, err := adminPool.Exec(cleanupCtx, fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, schemaName)); err != nil {
			t.Fatalf("drop test schema: %v", err)
		}

		adminPool.Close()
	})

	return pool
}

func repositoryTestDatabaseURL(t *testing.T) string {
	t.Helper()

	if dsn := os.Getenv(storefrontTestDatabaseURLEnv); dsn != "" {
		return dsn
	}
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	t.Skipf("set %s or DATABASE_URL to run repository integration tests", storefrontTestDatabaseURLEnv)
	return ""
}

func newTestPool(t *testing.T, databaseURL, searchPath string) *pgxpool.Pool {
	t.Helper()

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse database config: %v", err)
	}
	if searchPath != "" {
		if cfg.ConnConfig.RuntimeParams == nil {
			cfg.ConnConfig.RuntimeParams = make(map[string]string)
		}
		cfg.ConnConfig.RuntimeParams["search_path"] = searchPath
	}
	cfg.MaxConns = 4
	cfg.MinConns = 0

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping database: %v", err)
	}

	return pool
}

func applyMigrations(t *testing.T, databaseURL, searchPath string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	connConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse migration database config: %v", err)
	}
	if connConfig.RuntimeParams == nil {
		connConfig.RuntimeParams = make(map[string]string)
	}
	connConfig.RuntimeParams["search_path"] = searchPath

	db := stdlib.OpenDB(*connConfig)
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping migration database: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	goose.SetSequential(true)
	goose.SetBaseFS(migrations.FS)

	if err := goose.UpContext(ctx, db, "."); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
}

func createTenantFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool, status models.TenantStatus) uuid.UUID {
	t.Helper()

	var tenantID uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO tenants (tier_id, name, slug, status)
		VALUES ((SELECT id FROM tiers ORDER BY created_at LIMIT 1), $1, $2, $3)
		RETURNING id`,
		"Tenant "+uuid.NewString(),
		"tenant-"+uuid.NewString(),
		status,
	).Scan(&tenantID)
	if err != nil {
		t.Fatalf("create tenant fixture: %v", err)
	}

	return tenantID
}

func createProductVariantFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID, price decimal.Decimal, costPrice *decimal.Decimal) uuid.UUID {
	t.Helper()

	var productID uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO products (tenant_id, name)
		VALUES ($1, $2)
		RETURNING id`,
		tenantID,
		"Product "+uuid.NewString(),
	).Scan(&productID)
	if err != nil {
		t.Fatalf("create product fixture: %v", err)
	}

	var variantID uuid.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO product_variants (product_id, sku, price, cost_price, is_default)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		productID,
		"SKU-"+strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")),
		price,
		costPrice,
		true,
	).Scan(&variantID)
	if err != nil {
		t.Fatalf("create product variant fixture: %v", err)
	}

	return variantID
}
