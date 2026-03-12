package terminalaf_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"

	"storefront/backend/internal/adapter/terminalaf"
)

const testKey = "ta_test_key"

func newTestClient(srv *httptest.Server) *terminalaf.Client {
	return terminalaf.NewWithBaseURL(testKey, srv.URL)
}

func TestGetRates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/rates" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Rates fetched",
			"data": []map[string]any{
				{
					"carrier_id":   "gig",
					"carrier_name": "GIG Logistics",
					"service_type": "economy",
					"amount":       "1500.00",
					"eta":          "2-3 days",
				},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	rates, err := c.GetRates(context.Background(), terminalaf.RateRequest{
		SenderAddress:   terminalaf.Address{Name: "Sender", Phone: "08012345678", Line1: "1 Lagos St", City: "Lagos", State: "Lagos", Country: "NG"},
		ReceiverAddress: terminalaf.Address{Name: "Buyer", Phone: "08098765432", Line1: "2 Abuja Ave", City: "Abuja", State: "FCT", Country: "NG"},
		Parcel:          terminalaf.Parcel{WeightKg: decimal.NewFromFloat(1.5)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rates) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(rates))
	}
	if rates[0].CarrierID != "gig" {
		t.Errorf("expected gig, got %s", rates[0].CarrierID)
	}
}

func TestBookShipment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/shipments" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Shipment created",
			"data": map[string]any{
				"shipment_id":     "TA-12345",
				"tracking_number": "GIG-0099887",
				"status":          "queued",
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	resp, err := c.BookShipment(context.Background(), terminalaf.BookRequest{
		CarrierID: "gig",
		SenderAddress: terminalaf.Address{
			Name: "Sender", Phone: "08012345678", Line1: "1 Lagos St", City: "Lagos", State: "Lagos", Country: "NG",
		},
		ReceiverAddress: terminalaf.Address{
			Name: "Buyer", Phone: "08098765432", Line1: "2 Abuja Ave", City: "Abuja", State: "FCT", Country: "NG",
		},
		Parcel:    terminalaf.Parcel{WeightKg: decimal.NewFromFloat(1.5)},
		Reference: "order-uuid-here",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CarrierRef != "TA-12345" {
		t.Errorf("expected TA-12345, got %s", resp.CarrierRef)
	}
	if resp.TrackingNumber != "GIG-0099887" {
		t.Errorf("expected GIG-0099887, got %s", resp.TrackingNumber)
	}
	if resp.Status != "queued" {
		t.Errorf("expected queued, got %s", resp.Status)
	}
}

func TestBookShipment_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"status":  false,
			"message": "Invalid carrier",
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.BookShipment(context.Background(), terminalaf.BookRequest{CarrierID: "bad"})
	if err == nil {
		t.Error("expected error on API failure")
	}
}

func TestVerifyWebhookSignature_Valid(t *testing.T) {
	payload := []byte(`{"event":"shipment.delivered","data":{}}`)
	mac := hmac.New(sha256.New, []byte(testKey))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	c := terminalaf.New(testKey)
	if !c.VerifyWebhookSignature(payload, sig) {
		t.Error("expected valid signature to pass")
	}
}

func TestVerifyWebhookSignature_Tampered(t *testing.T) {
	payload := []byte(`{"event":"shipment.delivered","data":{}}`)
	c := terminalaf.New(testKey)
	if c.VerifyWebhookSignature(payload, "badsig") {
		t.Error("expected tampered signature to fail")
	}
}
