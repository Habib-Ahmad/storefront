package paystack_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"

	"storefront/backend/internal/adapter/paystack"
)

const testSecret = "sk_test_secret"

func newTestClient(srv *httptest.Server) *paystack.Client {
	return paystack.NewWithBaseURL(testSecret, srv.URL)
}

func TestInitializeTransaction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transaction/initialize" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Authorization URL created",
			"data": map[string]any{
				"authorization_url": "https://checkout.paystack.com/abc123",
				"reference":         "ref_xyz",
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	resp, err := c.InitializeTransaction(context.Background(), paystack.InitializeRequest{
		Email:     "buyer@example.com",
		Amount:    decimal.NewFromInt(5000),
		Reference: "ref_xyz",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AuthorizationURL != "https://checkout.paystack.com/abc123" {
		t.Errorf("wrong authorization_url: %s", resp.AuthorizationURL)
	}
	if resp.Reference != "ref_xyz" {
		t.Errorf("wrong reference: %s", resp.Reference)
	}
}

func TestVerifyTransaction_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transaction/verify/ref_xyz" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Verification successful",
			"data": map[string]any{
				"status":    "success",
				"amount":    500000,
				"reference": "ref_xyz",
				"paid_at":   "2026-03-12T10:00:00.000Z",
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	resp, err := c.VerifyTransaction(context.Background(), "ref_xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "success" {
		t.Errorf("expected success, got %s", resp.Status)
	}
	if !resp.Amount.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("expected 5000 naira, got %s", resp.Amount)
	}
}

func TestVerifyTransaction_Failed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Verification successful",
			"data": map[string]any{
				"status":    "failed",
				"amount":    500000,
				"reference": "ref_fail",
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	resp, err := c.VerifyTransaction(context.Background(), "ref_fail")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "failed" {
		t.Errorf("expected failed, got %s", resp.Status)
	}
}

func TestVerifyWebhookSignature_Valid(t *testing.T) {
	payload := []byte(`{"event":"charge.success","data":{}}`)
	mac := hmac.New(sha512.New, []byte(testSecret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	c := paystack.New(testSecret)
	if !c.VerifyWebhookSignature(payload, sig) {
		t.Error("expected valid signature to pass")
	}
}

func TestVerifyWebhookSignature_Tampered(t *testing.T) {
	payload := []byte(`{"event":"charge.success","data":{}}`)
	c := paystack.New(testSecret)
	if c.VerifyWebhookSignature(payload, "badsignature") {
		t.Error("expected tampered signature to fail")
	}
}

func TestInitializeTransaction_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"status":  false,
			"message": "Invalid key",
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.InitializeTransaction(context.Background(), paystack.InitializeRequest{
		Email:     "x@x.com",
		Amount:    decimal.NewFromInt(100),
		Reference: "ref_bad",
	})
	if err == nil {
		t.Error("expected error on API failure")
	}
}
