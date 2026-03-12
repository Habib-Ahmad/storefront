package terminalaf

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

const defaultBaseURL = "https://api.terminal.africa"

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func New(apiKey string) *Client {
	return NewWithBaseURL(apiKey, defaultBaseURL)
}

func NewWithBaseURL(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) GetRates(ctx context.Context, req RateRequest) ([]Rate, error) {
	var out struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    []struct {
			CarrierID   string          `json:"carrier_id"`
			CarrierName string          `json:"carrier_name"`
			ServiceType string          `json:"service_type"`
			Amount      decimal.Decimal `json:"amount"`
			ETA         string          `json:"eta"`
		} `json:"data"`
	}
	if err := c.post(ctx, "/v1/rates", req, &out); err != nil {
		return nil, err
	}
	if !out.Status {
		return nil, fmt.Errorf("terminal africa: get rates: %s", out.Message)
	}
	rates := make([]Rate, len(out.Data))
	for i, r := range out.Data {
		rates[i] = Rate{
			CarrierID:   r.CarrierID,
			CarrierName: r.CarrierName,
			ServiceType: r.ServiceType,
			Amount:      r.Amount,
			ETA:         r.ETA,
		}
	}
	return rates, nil
}

func (c *Client) BookShipment(ctx context.Context, req BookRequest) (*BookResponse, error) {
	var out struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			ShipmentID     string `json:"shipment_id"`
			TrackingNumber string `json:"tracking_number"`
			Status         string `json:"status"`
		} `json:"data"`
	}
	if err := c.post(ctx, "/v1/shipments", req, &out); err != nil {
		return nil, err
	}
	if !out.Status {
		return nil, fmt.Errorf("terminal africa: book shipment: %s", out.Message)
	}
	return &BookResponse{
		CarrierRef:     out.Data.ShipmentID,
		TrackingNumber: out.Data.TrackingNumber,
		Status:         out.Data.Status,
	}, nil
}

func (c *Client) VerifyWebhookSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(c.apiKey))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

type Address struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email,omitempty"`
	Line1   string `json:"line1"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

type Parcel struct {
	WeightKg decimal.Decimal `json:"weight"`
	Length   decimal.Decimal `json:"length"`
	Width    decimal.Decimal `json:"width"`
	Height   decimal.Decimal `json:"height"`
}

type RateRequest struct {
	SenderAddress   Address `json:"sender_address"`
	ReceiverAddress Address `json:"receiver_address"`
	Parcel          Parcel  `json:"parcel"`
}

type Rate struct {
	CarrierID   string
	CarrierName string
	ServiceType string
	Amount      decimal.Decimal
	ETA         string
}

type BookRequest struct {
	CarrierID       string  `json:"carrier_id"`
	SenderAddress   Address `json:"sender_address"`
	ReceiverAddress Address `json:"receiver_address"`
	Parcel          Parcel  `json:"parcel"`
	Reference       string  `json:"metadata_reference"`
}

type BookResponse struct {
	CarrierRef     string
	TrackingNumber string
	Status         string
}

type WebhookEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
