package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"storefront/backend/internal/config"
	"storefront/backend/internal/db"
	handler "storefront/backend/internal/handler"
	"storefront/backend/internal/repository"
	"storefront/backend/internal/router"
	"storefront/backend/internal/scheduler"
	"storefront/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
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
	tenantH := handler.NewTenantHandler(tenantSvc)
	productH := handler.NewProductHandler(productSvc)
	orderH := handler.NewOrderHandler(orderSvc)
	walletH := handler.NewWalletHandler(walletRepo, txRepo)

	// Monthly audit log partitions
	go scheduler.RunMonthlyPartitioner(ctx, pool)

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      router.New(tenantH, productH, orderH, walletH, userRepo, tenantRepo, cfg.SupabaseJWTSecret),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	os.Exit(0)
}
