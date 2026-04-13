package config_test

import (
	"os"
	"strings"
	"testing"

	"storefront/backend/internal/config"
)

func TestLoad_RequiresPaymentConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://storefront:test@localhost/storefront")
	t.Setenv("HMAC_SECRET", "test-secret")
	t.Setenv("SUPABASE_URL", "https://supabase.test")
	t.Setenv("PAYSTACK_SECRET_KEY", "")
	t.Setenv("PUBLIC_APP_URL", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected config validation error, got nil")
	}
	message := err.Error()
	if !strings.Contains(message, "PAYSTACK_SECRET_KEY is required") {
		t.Fatalf("expected PAYSTACK_SECRET_KEY validation error, got %s", message)
	}
	if !strings.Contains(message, "PUBLIC_APP_URL is required") {
		t.Fatalf("expected PUBLIC_APP_URL validation error, got %s", message)
	}
	_ = os.Unsetenv("PAYSTACK_SECRET_KEY")
	_ = os.Unsetenv("PUBLIC_APP_URL")
}
