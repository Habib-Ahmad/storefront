package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Port        string
	Environment string
	LogLevel    string

	DatabaseURL string
	HMACSecret  string

	SupabaseURL string

	PaystackSecretKey string

	TerminalAfricaAPIKey string
	ShipbubbleAPIKey     string

	R2AccountID  string
	R2AccessKey  string
	R2SecretKey  string
	R2BucketName string
	R2PublicURL  string

	AllowedOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		DatabaseURL: os.Getenv("DATABASE_URL"),
		HMACSecret:  os.Getenv("HMAC_SECRET"),

		SupabaseURL: os.Getenv("SUPABASE_URL"),

		PaystackSecretKey: os.Getenv("PAYSTACK_SECRET_KEY"),

		TerminalAfricaAPIKey: os.Getenv("TERMINAL_AFRICA_API_KEY"),
		ShipbubbleAPIKey:     os.Getenv("SHIPBUBBLE_API_KEY"),

		R2AccountID:  os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKey:  os.Getenv("R2_ACCESS_KEY"),
		R2SecretKey:  os.Getenv("R2_SECRET_KEY"),
		R2BucketName: os.Getenv("R2_BUCKET_NAME"),
		R2PublicURL:  os.Getenv("R2_PUBLIC_URL"),

		AllowedOrigins: parseOrigins(os.Getenv("ALLOWED_ORIGINS")),
	}

	var errs []error
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if cfg.HMACSecret == "" {
		errs = append(errs, errors.New("HMAC_SECRET is required"))
	}
	if cfg.SupabaseURL == "" {
		errs = append(errs, errors.New("SUPABASE_URL is required"))
	}

	return cfg, errors.Join(errs...)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return []string{"http://localhost:3000"}
	}
	var origins []string
	for _, o := range strings.Split(raw, ",") {
		if t := strings.TrimSpace(o); t != "" {
			origins = append(origins, t)
		}
	}
	return origins
}
