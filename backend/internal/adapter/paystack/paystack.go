package paystack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
)

const defaultBaseURL = "https://api.paystack.co"

type Client struct {
	secretKey  string
	baseURL    string
	httpClient *http.Client
}

func New(secretKey string) *Client {
	return NewWithBaseURL(secretKey, defaultBaseURL)
}

func NewWithBaseURL(secretKey, baseURL string) *Client {
	return &Client{
		secretKey:  secretKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) InitializeTransaction(ctx context.Context, req InitializeRequest) (*InitializeResponse, error) {
	body := map[string]any{
		"email":     req.Email,
		"amount":    req.Amount.Mul(decimal.NewFromInt(100)).IntPart(),
		"reference": req.Reference,
		"metadata":  req.Metadata,
	}
	if req.SubaccountCode != "" {
		body["subaccount"] = req.SubaccountCode
	}

	var out struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}
	if err := c.post(ctx, "/transaction/initialize", body, &out); err != nil {
		return nil, err
	}
	if !out.Status {
		return nil, fmt.Errorf("paystack: initialize: %s", out.Message)
	}
	return &InitializeResponse{
		AuthorizationURL: out.Data.AuthorizationURL,
		Reference:        out.Data.Reference,
	}, nil
}

func (c *Client) VerifyTransaction(ctx context.Context, reference string) (*VerifyResponse, error) {
	var out struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status    string          `json:"status"`
			Amount    int64           `json:"amount"`
			Reference string          `json:"reference"`
			PaidAt    string          `json:"paid_at"`
			Customer  json.RawMessage `json:"customer"`
		} `json:"data"`
	}
	if err := c.get(ctx, "/transaction/verify/"+url.PathEscape(reference), &out); err != nil {
		return nil, err
	}
	if !out.Status {
		return nil, fmt.Errorf("paystack: verify: %s", out.Message)
	}
	amount := decimal.NewFromInt(out.Data.Amount).Div(decimal.NewFromInt(100))
	return &VerifyResponse{
		Status:    out.Data.Status,
		Amount:    amount,
		Reference: out.Data.Reference,
		PaidAt:    out.Data.PaidAt,
	}, nil
}

func (c *Client) VerifyWebhookSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha512.New, []byte(c.secretKey))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

type InitializeRequest struct {
	Email          string
	Amount         decimal.Decimal
	Reference      string
	SubaccountCode string
	Metadata       map[string]any
}

type InitializeResponse struct {
	AuthorizationURL string
	Reference        string
}

type VerifyResponse struct {
	Status    string
	Amount    decimal.Decimal
	Reference string
	PaidAt    string
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
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("paystack POST %s: status %d", path, resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.secretKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("paystack GET %s: status %d", path, resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
}
