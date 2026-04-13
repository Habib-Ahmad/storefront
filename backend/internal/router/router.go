package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"

	handler "storefront/backend/internal/handler"
	mw "storefront/backend/internal/middleware"
	"storefront/backend/internal/repository"
)

func New(
	log *slog.Logger,
	auth *handler.AuthHandler,
	tier *handler.TierHandler,
	storefront *handler.StorefrontHandler,
	tenant *handler.TenantHandler,
	user *handler.UserHandler,
	product *handler.ProductHandler,
	order *handler.OrderHandler,
	wallet *handler.WalletHandler,
	analytics *handler.AnalyticsHandler,
	webhook *handler.WebhookHandler,
	media *handler.MediaHandler,
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	jwtKeyFunc jwt.Keyfunc,
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
	r.Get("/storefronts/{slug}", storefront.GetPublic)
	r.Get("/storefronts/{slug}/products/{id}", storefront.GetPublicProduct)
	r.Post("/storefronts/{slug}/orders", order.CreatePublic)
	r.Get("/track/{slug}", order.Track)
	r.Post("/track/{slug}/resume-payment", order.ResumePaymentPublic)

	// Authenticated but pre-tenant routes (user has no tenant yet)
	r.Group(func(r chi.Router) {
		r.Use(mw.Authenticate(jwtKeyFunc))
		r.Get("/auth/me", auth.Me)
		r.Post("/tenants/onboard", tenant.Onboard)
	})

	// Authenticated + tenant-resolved routes
	r.Group(func(r chi.Router) {
		r.Use(mw.Authenticate(jwtKeyFunc))
		r.Use(mw.ResolveTenant(userRepo, tenantRepo))

		r.Put("/tenants/me", tenant.UpdateProfile)
		r.Put("/tenants/me/storefront", tenant.UpdateStorefront)
		r.Put("/tenants/me/modules", tenant.SetModules)

		r.Get("/users/me", user.GetMe)
		r.Put("/users/me", user.UpdateProfile)

		r.Post("/products", product.Create)
		r.Get("/products", product.List)
		r.Get("/products/{id}", product.Get)
		r.Put("/products/{id}", product.Update)
		r.Delete("/products/{id}", product.Delete)
		r.Post("/products/{id}/variants", product.CreateVariant)
		r.Get("/products/{id}/variants", product.ListVariants)
		r.Put("/products/{id}/variants/{variantId}", product.UpdateVariant)
		r.Delete("/products/{id}/variants/{variantId}", product.DeleteVariant)
		r.Post("/products/{id}/images", product.AddImage)
		r.Get("/products/{id}/images", product.ListImages)
		r.Put("/products/{id}/images/{imageId}", product.UpdateImage)
		r.Delete("/products/{id}/images/{imageId}", product.DeleteImage)

		r.Post("/orders", order.Create)
		r.Get("/orders", order.List)
		r.Get("/orders/{id}", order.Get)
		r.Get("/orders/{id}/items", order.ListItems)
		r.Post("/orders/{id}/cancel", order.Cancel)
		r.Post("/orders/{id}/resume-payment", order.ResumePayment)
		r.Post("/orders/{id}/dispatch", order.Dispatch)

		r.Get("/wallet", wallet.GetBalance)
		r.Get("/wallet/transactions", wallet.ListTransactions)

		r.Get("/analytics/summary", analytics.Summary)

		r.Post("/media/upload-url", media.GetUploadURL)
	})

	return r
}
