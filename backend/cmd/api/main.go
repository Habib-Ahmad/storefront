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
	"storefront/backend/internal/adapter/terminalaf"
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
	terminalClient := terminalaf.New(cfg.TerminalAfricaAPIKey)

	// Warn if adapter API keys are missing
	if cfg.PaystackSecretKey == "" {
		log.Warn("PAYSTACK_SECRET_KEY is empty — payment webhooks will fail")
	}
	if cfg.PublicAppURL == "" {
		log.Warn("PUBLIC_APP_URL is empty — public Paystack callbacks will not return to order confirmation")
	}
	if cfg.TerminalAfricaAPIKey == "" {
		log.Warn("TERMINAL_AFRICA_API_KEY is empty — shipping features will fail")
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
	shipmentSvc := service.NewShipmentService(terminalClient, shipmentRepo, orderRepo, walletSvc)

	// Handlers
	authH := handler.NewAuthHandler(userRepo, tenantRepo, log)
	tierH := handler.NewTierHandler(tierRepo, log)
	storefrontH := handler.NewStorefrontHandler(storefrontSvc, log)
	tenantH := handler.NewTenantHandler(tenantSvc, log)
	userSvc := service.NewUserService(userRepo)
	userH := handler.NewUserHandler(userSvc, log)
	productH := handler.NewProductHandler(productSvc, log)
	orderH := handler.NewOrderHandler(orderSvc, paymentSvc, shipmentSvc, cfg.PublicAppURL, log)
	walletH := handler.NewWalletHandler(walletRepo, txRepo, log)
	analyticsRepo := repository.NewAnalyticsRepository(pool)
	analyticsH := handler.NewAnalyticsHandler(analyticsRepo, log)
	webhookH := handler.NewWebhookHandler(paystackClient, terminalClient, paymentSvc, shipmentSvc, log)
	mediaH := handler.NewMediaHandler(cfg.CloudflareAccountID, cfg.CloudflareAPIToken, log)

	// Ensure audit log partitions exist on startup (fresh deploy safety).
	if err := scheduler.EnsureAuditLogPartitions(ctx, pool); err != nil {
		log.Warn("startup partition check", "error", err)
	}

	// Monthly audit log partitions
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		scheduler.RunMonthlyPartitioner(ctx, pool, log)
	}()
	// Daily HMAC chain verification across all active tenants — run once on startup too.
	go func() {
		defer wg.Done()
		scheduler.RunDailyChainVerifier(ctx, walletRepo, walletSvc, log)
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
