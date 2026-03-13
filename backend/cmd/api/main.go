package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

	// Services
	tenantSvc := service.NewTenantService(tenantRepo, tierRepo, walletRepo, userRepo)
	productSvc := service.NewProductService(productRepo)
	orderSvc := service.NewOrderService(orderRepo, productRepo)
	walletSvc := service.NewWalletService(walletRepo, txRepo, tenantRepo, cfg.HMACSecret)
	walletSvc.SetTierRepo(tierRepo)
	walletSvc.SetAuditLogRepo(auditLogRepo)
	paymentSvc := service.NewPaymentService(paystackClient, orderRepo, walletSvc, tierRepo, tenantRepo)
	shipmentSvc := service.NewShipmentService(terminalClient, shipmentRepo, orderRepo, walletSvc, tenantRepo, tierRepo)

	// Handlers
	tierH := handler.NewTierHandler(tierRepo, log)
	tenantH := handler.NewTenantHandler(tenantSvc, log)
	productH := handler.NewProductHandler(productSvc, log)
	orderH := handler.NewOrderHandler(orderSvc, paymentSvc, log)
	walletH := handler.NewWalletHandler(walletRepo, txRepo, log)
	webhookH := handler.NewWebhookHandler(paystackClient, terminalClient, paymentSvc, shipmentSvc, log)

	// Ensure audit log partitions exist on startup (fresh deploy safety).
	if err := scheduler.EnsureAuditLogPartitions(ctx, pool); err != nil {
		log.Warn("startup partition check", "error", err)
	}

	// Monthly audit log partitions
	go scheduler.RunMonthlyPartitioner(ctx, pool)
	// Daily HMAC chain verification across all active tenants — run once on startup too.
	go scheduler.RunDailyChainVerifier(ctx, pool, walletSvc)

	// Fetch Supabase JWKS (ES256 public key) for JWT verification
	ecKey, err := config.FetchJWKS(cfg.SupabaseURL)
	if err != nil {
		log.Warn("jwks fetch failed, falling back to HS256 only", "error", err)
	}
	jwtKeyFunc := mw.NewKeyFunc(ecKey, cfg.SupabaseJWTSecret)

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      router.New(log, tierH, tenantH, productH, orderH, walletH, webhookH, userRepo, tenantRepo, jwtKeyFunc, cfg.AllowedOrigins),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("shutdown", "error", err)
	}
	os.Exit(0)
}
