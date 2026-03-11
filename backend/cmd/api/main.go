package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"storefront/backend/internal/config"
	"storefront/backend/internal/db"
	handler "storefront/backend/internal/handler"
	"storefront/backend/internal/logger"
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

	log := logger.New(cfg.Environment)

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

	_ = tierRepo // used indirectly via onboarding

	// Services
	tenantSvc := service.NewTenantService(tenantRepo, walletRepo, userRepo)
	productSvc := service.NewProductService(productRepo)
	orderSvc := service.NewOrderService(orderRepo, productRepo)

	// Handlers
	tierH := handler.NewTierHandler(tierRepo, log)
	tenantH := handler.NewTenantHandler(tenantSvc, log)
	productH := handler.NewProductHandler(productSvc, log)
	orderH := handler.NewOrderHandler(orderSvc, log)
	walletH := handler.NewWalletHandler(walletRepo, txRepo, log)

	// Monthly audit log partitions
	go scheduler.RunMonthlyPartitioner(ctx, pool)

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      router.New(log, tierH, tenantH, productH, orderH, walletH, userRepo, tenantRepo, cfg.SupabaseJWTSecret),
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
