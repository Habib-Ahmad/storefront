package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"storefront/backend/internal/config"
	"storefront/backend/internal/db"
	handler "storefront/backend/internal/handler"
	"storefront/backend/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Environment, cfg.LogLevel)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	media := handler.NewMediaHandler(cfg.R2BucketName, cfg.R2S3API, cfg.R2AccessKey, cfg.R2SecretKey, log)
	if cfg.R2BucketName == "" || cfg.R2S3API == "" || cfg.R2AccessKey == "" || cfg.R2SecretKey == "" {
		log.Error("product media cleanup requires R2 configuration")
		os.Exit(1)
	}

	activeKeys, err := loadActiveProductImageKeys(ctx, pool)
	if err != nil {
		log.Error("load active product image keys", "error", err)
		os.Exit(1)
	}

	objectKeys, err := media.ListObjectKeys(ctx, "tenants/")
	if err != nil {
		log.Error("list product media objects", "error", err)
		os.Exit(1)
	}

	scanned := 0
	retained := 0
	deleted := 0
	failures := 0
	for _, key := range objectKeys {
		if !strings.Contains(key, "/products/") {
			continue
		}
		scanned++
		if _, ok := activeKeys[key]; ok {
			retained++
			continue
		}
		if err := media.DeleteObjectByKey(ctx, key); err != nil {
			failures++
			log.Error("delete stale product media", "key", key, "error", err)
			continue
		}
		deleted++
		log.Info("deleted stale product media", "key", key)
	}

	log.Info("product media cleanup complete",
		"scanned", scanned,
		"retained", retained,
		"deleted", deleted,
		"failures", failures,
	)

	if failures > 0 {
		os.Exit(1)
	}
}

func loadActiveProductImageKeys(ctx context.Context, pool db.DBTX) (map[string]struct{}, error) {
	rows, err := pool.Query(ctx, `
		SELECT pi.url
		FROM product_images pi
		JOIN products p ON p.id = pi.product_id
		WHERE p.deleted_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make(map[string]struct{})
	for rows.Next() {
		var objectURL string
		if err := rows.Scan(&objectURL); err != nil {
			return nil, err
		}
		key, err := objectKeyFromURL(objectURL)
		if err != nil {
			return nil, fmt.Errorf("parse product image url %q: %w", objectURL, err)
		}
		keys[key] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

func objectKeyFromURL(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	key := strings.TrimSpace(parsed.Query().Get("key"))
	if key == "" {
		return "", fmt.Errorf("missing key query param")
	}
	return key, nil
}
