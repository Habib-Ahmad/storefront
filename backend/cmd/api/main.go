package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"storefront/backend/internal/adapter/paystack"
	"storefront/backend/internal/adapter/shipbubble"
	"storefront/backend/internal/config"
	"storefront/backend/internal/db"
	handler "storefront/backend/internal/handler"
	"storefront/backend/internal/logger"
	mw "storefront/backend/internal/middleware"
	"storefront/backend/internal/repository"
	"storefront/backend/internal/router"
	"storefront/backend/internal/scheduler"
	"storefront/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Environment, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Repositories
	tierRepo := repository.NewTierRepository(pool)
	tenantRepo := repository.NewTenantRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	walletRepo := repository.NewWalletRepository(pool)
	txRepo := repository.NewTransactionRepository(pool)
	productRepo := repository.NewProductRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)
	shipmentRepo := repository.NewShipmentRepository(pool)
	auditLogRepo := repository.NewAuditLogRepository(pool)

	// External adapter clients
	paystackClient := paystack.New(cfg.PaystackSecretKey)
	shipbubbleClient := shipbubble.New(cfg.ShipbubbleAPIKey)

	// Warn if non-core adapter API keys are missing
	if cfg.ShipbubbleAPIKey == "" {
		log.Warn("SHIPBUBBLE_API_KEY is empty — delivery quote features will fail")
	}
	if cfg.PendingOrderTTL <= 0 {
		log.Warn("PENDING_ORDER_TTL disables stale pending-order cleanup")
	}

	// Services
	tenantSvc := service.NewTenantService(tenantRepo, tierRepo, walletRepo, userRepo)
	tenantSvc.SetPool(pool)
	productSvc := service.NewProductService(productRepo)
	storefrontSvc := service.NewStorefrontService(tenantRepo, productRepo)
	orderSvc := service.NewOrderService(orderRepo, productRepo)
	orderSvc.SetPool(pool)
	walletSvc := service.NewWalletService(walletRepo, txRepo, tenantRepo, cfg.HMACSecret)
	walletSvc.SetTierRepo(tierRepo)
	walletSvc.SetAuditLogRepo(auditLogRepo)
	walletSvc.SetPool(pool)
	orderSvc.SetWalletService(walletSvc)
	orderSvc.SetTenantRepo(tenantRepo)
	orderSvc.SetTierRepo(tierRepo)
	paymentSvc := service.NewPaymentService(paystackClient, orderRepo, productRepo, walletSvc)
	paymentSvc.SetPool(pool)
	deliveryQuoteSvc := service.NewDeliveryQuoteService(storefrontSvc, productRepo, shipbubbleClient)
	shipmentSvc := service.NewShipmentService(shipbubbleClient, shipmentRepo, orderRepo, productRepo, tenantRepo, walletSvc)

	// Handlers
	authH := handler.NewAuthHandler(userRepo, tenantRepo, log)
	tierH := handler.NewTierHandler(tierRepo, log)
	storefrontH := handler.NewStorefrontHandler(storefrontSvc, log)
	tenantH := handler.NewTenantHandler(tenantSvc, log)
	userSvc := service.NewUserService(userRepo)
	userH := handler.NewUserHandler(userSvc, log)
	orderH := handler.NewOrderHandler(orderSvc, paymentSvc, cfg.PublicAppURL, log)
	orderH.SetDeliveryQuoteService(deliveryQuoteSvc)
	orderH.SetShipmentService(shipmentSvc)
	walletH := handler.NewWalletHandler(walletRepo, txRepo, log)
	analyticsRepo := repository.NewAnalyticsRepository(pool)
	analyticsH := handler.NewAnalyticsHandler(analyticsRepo, log)
	webhookH := handler.NewWebhookHandler(paystackClient, paymentSvc, log)
	webhookH.SetShipmentService(shipbubbleClient, shipmentSvc)
	mediaH := handler.NewMediaHandler(
		cfg.R2BucketName,
		cfg.R2S3API,
		cfg.R2AccessKey,
		cfg.R2SecretKey,
		log,
	)
	productH := handler.NewProductHandler(productSvc, mediaH, log)

	// Ensure audit log partitions exist on startup (fresh deploy safety).
	if err := scheduler.EnsureAuditLogPartitions(ctx, pool); err != nil {
		log.Warn("startup partition check", "error", err)
	}

	// Monthly audit log partitions
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		scheduler.RunMonthlyPartitioner(ctx, pool, log)
	}()
	// Daily HMAC chain verification across all active tenants — run once on startup too.
	go func() {
		defer wg.Done()
		scheduler.RunDailyChainVerifier(ctx, walletRepo, walletSvc, log)
	}()
	go func() {
		defer wg.Done()
		scheduler.RunPendingOrderExpiry(ctx, pool, paymentSvc, cfg.PendingOrderTTL, log)
	}()

	// Fetch Supabase JWKS (ES256 public key) for JWT verification
	ecKey, err := config.FetchJWKS(cfg.SupabaseURL)
	if err != nil {
		log.Error("failed to fetch Supabase JWKS", "error", err)
		os.Exit(1)
	}
	if ecKey == nil {
		log.Error("no EC key found in Supabase JWKS endpoint")
		os.Exit(1)
	}
	jwtKeyFunc := mw.NewKeyFunc(ecKey)

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      router.New(log, authH, tierH, storefrontH, tenantH, userH, productH, orderH, walletH, analyticsH, webhookH, mediaH, userRepo, tenantRepo, jwtKeyFunc, cfg.AllowedOrigins),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		log.Error("server", "error", err)
	}
	log.Info("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("shutdown", "error", err)
	}
	wg.Wait()
}
