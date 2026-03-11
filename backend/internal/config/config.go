package config

import (
	"errors"
	"os"
)

type Config struct {
	Port        string
	Environment string

	DatabaseURL string
	HMACSecret  string

	SupabaseURL       string
	SupabaseJWTSecret string

	PaystackSecretKey string

	TerminalAfricaAPIKey string
	ShipbubbleAPIKey     string

	R2AccountID  string
	R2AccessKey  string
	R2SecretKey  string
	R2BucketName string
	R2PublicURL  string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		DatabaseURL: os.Getenv("DATABASE_URL"),
		HMACSecret:  os.Getenv("HMAC_SECRET"),

		SupabaseURL:       os.Getenv("SUPABASE_URL"),
		SupabaseJWTSecret: os.Getenv("SUPABASE_JWT_SECRET"),

		PaystackSecretKey: os.Getenv("PAYSTACK_SECRET_KEY"),

		TerminalAfricaAPIKey: os.Getenv("TERMINAL_AFRICA_API_KEY"),
		ShipbubbleAPIKey:     os.Getenv("SHIPBUBBLE_API_KEY"),

		R2AccountID:  os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKey:  os.Getenv("R2_ACCESS_KEY"),
		R2SecretKey:  os.Getenv("R2_SECRET_KEY"),
		R2BucketName: os.Getenv("R2_BUCKET_NAME"),
		R2PublicURL:  os.Getenv("R2_PUBLIC_URL"),
	}

	var errs []error
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if cfg.HMACSecret == "" {
		errs = append(errs, errors.New("HMAC_SECRET is required"))
	}
	if cfg.SupabaseJWTSecret == "" {
		errs = append(errs, errors.New("SUPABASE_JWT_SECRET is required"))
	}

	return cfg, errors.Join(errs...)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
