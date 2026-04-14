package shipbubble

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateShipment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method: want POST, got %s", r.Method)
		}
		if r.URL.Path != "/shipping/labels" {
			t.Fatalf("path: want /shipping/labels, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization: want Bearer test-key, got %s", got)
		}
		fmt.Fprint(w, `{"status":"success","message":"ok","data":{"order_id":"SB-123","status":"pending","courier":{"name":"Kwik","email":"ops@kwik.com","phone":"0800","tracking_code":"TRK-123","tracking_message":"Tracking code: TRK-123","rider_info":null},"ship_from":{"name":"Store","phone":"0801","email":"store@example.com","address":"12 Allen","latitude":6.5,"longitude":3.3},"ship_to":{"name":"Ada","phone":"0802","email":"ada@example.com","address":"23 Broad","latitude":6.6,"longitude":3.4},"payment":{"shipping_fee":2500,"type":"wallet","status":"completed","currency":"NGN"},"items":[],"package_status":[{"status":"Pending","datetime":"2025-01-01T10:00:00Z"}],"events":[],"tracking_url":"https://track.test/SB-123","waybill_document":null,"date":"2025-01-01T10:00:00Z"}}`)
	}))
	defer server.Close()

	client := NewWithBaseURL("test-key", server.URL)
	shipment, err := client.CreateShipment(context.Background(), CreateShipmentRequest{
		RequestToken: "req-token",
		ServiceCode:  "bike",
		CourierID:    "123",
	})
	if err != nil {
		t.Fatalf("CreateShipment returned error: %v", err)
	}
	if shipment.OrderID != "SB-123" {
		t.Fatalf("order_id: want SB-123, got %s", shipment.OrderID)
	}
	if shipment.Courier.TrackingCode != "TRK-123" {
		t.Fatalf("tracking_code: want TRK-123, got %s", shipment.Courier.TrackingCode)
	}
	if len(shipment.Raw) == 0 {
		t.Fatal("expected raw shipment payload to be captured")
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	client := New("ship-secret")
	payload := []byte(`{"event":"shipment.label.created"}`)
	mac := hmac.New(sha512.New, []byte("ship-secret"))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))
	if !client.VerifyWebhookSignature(payload, signature) {
		t.Fatal("expected webhook signature to verify")
	}
}
