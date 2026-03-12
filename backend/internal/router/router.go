package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	handler "storefront/backend/internal/handler"
	mw "storefront/backend/internal/middleware"
	"storefront/backend/internal/repository"
)

func New(
	log *slog.Logger,
	tier *handler.TierHandler,
	tenant *handler.TenantHandler,
	product *handler.ProductHandler,
	order *handler.OrderHandler,
	wallet *handler.WalletHandler,
	webhook *handler.WebhookHandler,
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	jwtSecret string,
	allowedOrigins []string,
) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(mw.RequestLogger(log))
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-Slug"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check — no auth required
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Webhook endpoints — no auth, signature-verified inside each handler
	r.Post("/webhooks/paystack", webhook.Paystack)
	r.Post("/webhooks/terminalaf", webhook.TerminalAf)

	// Public endpoints
	r.Get("/tiers", tier.List)
	r.Get("/track/{slug}", order.Track)

	// Authenticated but pre-tenant routes (user has no tenant yet)
	r.Group(func(r chi.Router) {
		r.Use(mw.Authenticate(jwtSecret))
		r.Post("/tenants/onboard", tenant.Onboard)
	})

	// Authenticated + tenant-resolved routes
	r.Group(func(r chi.Router) {
		r.Use(mw.Authenticate(jwtSecret))
		r.Use(mw.ResolveTenant(userRepo, tenantRepo))

		r.Get("/tenants/me", tenant.GetMe)
		r.Put("/tenants/me/modules", tenant.SetModules)

		r.Post("/products", product.Create)
		r.Get("/products", product.List)
		r.Delete("/products/{id}", product.Delete)

		r.Post("/orders", order.Create)
		r.Get("/orders", order.List)
		r.Get("/orders/{id}", order.Get)

		r.Get("/wallet", wallet.GetBalance)
		r.Get("/wallet/transactions", wallet.ListTransactions)
	})

	return r
}
