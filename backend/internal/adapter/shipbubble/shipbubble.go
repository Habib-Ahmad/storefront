package shipbubble

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
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const defaultBaseURL = "https://api.shipbubble.com/v1"

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
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type ValidateAddressRequest struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Address   string   `json:"address"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

type ValidatedAddress struct {
	Name             string  `json:"name"`
	Email            string  `json:"email"`
	Phone            string  `json:"phone"`
	FormattedAddress string  `json:"formatted_address"`
	Country          string  `json:"country"`
	CountryCode      string  `json:"country_code"`
	City             string  `json:"city"`
	CityCode         string  `json:"city_code"`
	State            string  `json:"state"`
	StateCode        string  `json:"state_code"`
	PostalCode       string  `json:"postal_code"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	AddressCode      int64   `json:"address_code"`
}

type PackageCategory struct {
	ID   int64  `json:"category_id"`
	Name string `json:"category"`
}

type PackageBox struct {
	BoxSizeID           int64           `json:"box_size_id"`
	Name                string          `json:"name"`
	DescriptionImageURL string          `json:"description_image_url"`
	Height              decimal.Decimal `json:"height"`
	Width               decimal.Decimal `json:"width"`
	Length              decimal.Decimal `json:"length"`
	MaxWeight           decimal.Decimal `json:"max_weight"`
}

type PackageItem struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	UnitWeight  decimal.Decimal `json:"unit_weight"`
	UnitAmount  string          `json:"unit_amount"`
	Quantity    string          `json:"quantity"`
}

type PackageDimension struct {
	Length decimal.Decimal `json:"length"`
	Width  decimal.Decimal `json:"width"`
	Height decimal.Decimal `json:"height"`
}

type RateRequest struct {
	SenderAddressCode    int64            `json:"sender_address_code"`
	ReceiverAddressCode  int64            `json:"reciever_address_code"`
	PickupDate           string           `json:"pickup_date"`
	CategoryID           int64            `json:"category_id"`
	PackageItems         []PackageItem    `json:"package_items"`
	ServiceType          string           `json:"service_type,omitempty"`
	DeliveryInstructions string           `json:"delivery_instructions,omitempty"`
	PackageDimension     PackageDimension `json:"package_dimension"`
	IsInvoiceRequired    bool             `json:"is_invoice_required"`
}

type Station struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type TrackingSummary struct {
	Bars  int    `json:"bars"`
	Label string `json:"label"`
}

type RateOption struct {
	CourierID        string          `json:"courier_id"`
	CourierName      string          `json:"courier_name"`
	CourierImage     string          `json:"courier_image"`
	ServiceCode      string          `json:"service_code"`
	ServiceType      string          `json:"service_type"`
	Waybill          bool            `json:"waybill"`
	OnDemand         bool            `json:"on_demand"`
	IsCODAvailable   bool            `json:"is_cod_available"`
	TrackingLevel    int             `json:"tracking_level"`
	Ratings          float64         `json:"ratings"`
	Votes            int             `json:"votes"`
	ConnectedAccount bool            `json:"connected_account"`
	RateCardAmount   decimal.Decimal `json:"rate_card_amount"`
	RateCardCurrency string          `json:"rate_card_currency"`
	PickupETA        string          `json:"pickup_eta"`
	PickupETATime    string          `json:"pickup_eta_time"`
	DropoffStation   *Station        `json:"dropoff_station,omitempty"`
	PickupStation    *Station        `json:"pickup_station,omitempty"`
	DeliveryETA      string          `json:"delivery_eta"`
	DeliveryETATime  string          `json:"delivery_eta_time"`
	Info             []string        `json:"info,omitempty"`
	Currency         string          `json:"currency"`
	VAT              decimal.Decimal `json:"vat"`
	Total            decimal.Decimal `json:"total"`
	Tracking         TrackingSummary `json:"tracking"`
	Raw              json.RawMessage `json:"-"`
}

type RateResponse struct {
	RequestToken string          `json:"request_token"`
	Options      []RateOption    `json:"couriers"`
	Fastest      *RateOption     `json:"fastest_courier,omitempty"`
	Cheapest     *RateOption     `json:"cheapest_courier,omitempty"`
	RawResponse  json.RawMessage `json:"-"`
}

type CreateShipmentRequest struct {
	RequestToken  string `json:"request_token"`
	ServiceCode   string `json:"service_code"`
	CourierID     string `json:"courier_id"`
	InsuranceCode string `json:"insurance_code,omitempty"`
	IsCODLabel    bool   `json:"is_cod_label,omitempty"`
}

type ShipmentCourier struct {
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	Phone           string          `json:"phone"`
	TrackingCode    string          `json:"tracking_code"`
	TrackingMessage string          `json:"tracking_message"`
	RiderInfo       json.RawMessage `json:"rider_info"`
}

type ShipmentParty struct {
	Name      string  `json:"name"`
	Phone     string  `json:"phone"`
	Email     string  `json:"email"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ShipmentPayment struct {
	ShippingFee decimal.Decimal `json:"shipping_fee"`
	Type        string          `json:"type"`
	Status      string          `json:"status"`
	Currency    string          `json:"currency"`
}

type ShipmentItem struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Weight      decimal.Decimal `json:"weight"`
	Amount      string          `json:"amount"`
	Quantity    string          `json:"quantity"`
	Total       decimal.Decimal `json:"total"`
}

type ShipmentPackageStatus struct {
	Status   string `json:"status"`
	DateTime string `json:"datetime"`
}

type ShipmentEvent struct {
	Location string `json:"location"`
	Message  string `json:"message"`
	Captured string `json:"captured"`
}

type ShipmentRecord struct {
	OrderID         string                  `json:"order_id"`
	Status          string                  `json:"status"`
	Courier         ShipmentCourier         `json:"courier"`
	ShipFrom        ShipmentParty           `json:"ship_from"`
	ShipTo          ShipmentParty           `json:"ship_to"`
	Payment         ShipmentPayment         `json:"payment"`
	Items           []ShipmentItem          `json:"items"`
	PackageStatus   []ShipmentPackageStatus `json:"package_status"`
	Events          []ShipmentEvent         `json:"events"`
	TrackingURL     string                  `json:"tracking_url"`
	WaybillDocument *string                 `json:"waybill_document"`
	Date            string                  `json:"date"`
	Raw             json.RawMessage         `json:"-"`
}

func (c *Client) ValidateAddress(ctx context.Context, req ValidateAddressRequest) (*ValidatedAddress, error) {
	var out struct {
		Status  string           `json:"status"`
		Message string           `json:"message"`
		Data    ValidatedAddress `json:"data"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/shipping/address/validate", req, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, fmt.Errorf("shipbubble: validate address: %s", out.Message)
	}
	return &out.Data, nil
}

func (c *Client) GetPackageCategories(ctx context.Context) ([]PackageCategory, error) {
	var out struct {
		Status  string            `json:"status"`
		Message string            `json:"message"`
		Data    []PackageCategory `json:"data"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/shipping/labels/categories", nil, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, fmt.Errorf("shipbubble: get package categories: %s", out.Message)
	}
	return out.Data, nil
}

func (c *Client) GetPackageBoxes(ctx context.Context) ([]PackageBox, error) {
	var out struct {
		Status  string       `json:"status"`
		Message string       `json:"message"`
		Data    []PackageBox `json:"data"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/shipping/labels/boxes", nil, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, fmt.Errorf("shipbubble: get package boxes: %s", out.Message)
	}
	return out.Data, nil
}

func (c *Client) FetchRates(ctx context.Context, req RateRequest) (*RateResponse, error) {
	type rawRateOption struct {
		CourierID        any             `json:"courier_id"`
		CourierName      string          `json:"courier_name"`
		CourierImage     string          `json:"courier_image"`
		ServiceCode      string          `json:"service_code"`
		ServiceType      string          `json:"service_type"`
		Waybill          bool            `json:"waybill"`
		OnDemand         bool            `json:"on_demand"`
		IsCODAvailable   bool            `json:"is_cod_available"`
		TrackingLevel    int             `json:"tracking_level"`
		Ratings          float64         `json:"ratings"`
		Votes            int             `json:"votes"`
		ConnectedAccount bool            `json:"connected_account"`
		RateCardAmount   decimal.Decimal `json:"rate_card_amount"`
		RateCardCurrency string          `json:"rate_card_currency"`
		PickupETA        string          `json:"pickup_eta"`
		PickupETATime    string          `json:"pickup_eta_time"`
		DropoffStation   *Station        `json:"dropoff_station"`
		PickupStation    *Station        `json:"pickup_station"`
		DeliveryETA      string          `json:"delivery_eta"`
		DeliveryETATime  string          `json:"delivery_eta_time"`
		Info             []string        `json:"info"`
		Currency         string          `json:"currency"`
		VAT              decimal.Decimal `json:"vat"`
		Total            decimal.Decimal `json:"total"`
		Tracking         TrackingSummary `json:"tracking"`
	}

	var out struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			RequestToken string          `json:"request_token"`
			Couriers     []rawRateOption `json:"couriers"`
			Fastest      *rawRateOption  `json:"fastest_courier"`
			Cheapest     *rawRateOption  `json:"cheapest_courier"`
		} `json:"data"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/shipping/fetch_rates", req, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, fmt.Errorf("shipbubble: fetch rates: %s", out.Message)
	}

	rawResponse, err := json.Marshal(out.Data)
	if err != nil {
		return nil, fmt.Errorf("shipbubble: marshal rates response: %w", err)
	}

	options := make([]RateOption, 0, len(out.Data.Couriers))
	for _, courier := range out.Data.Couriers {
		option, err := normalizeRateOption(courier)
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}

	response := &RateResponse{
		RequestToken: out.Data.RequestToken,
		Options:      options,
		RawResponse:  rawResponse,
	}
	if out.Data.Fastest != nil {
		option, err := normalizeRateOption(*out.Data.Fastest)
		if err != nil {
			return nil, err
		}
		response.Fastest = &option
	}
	if out.Data.Cheapest != nil {
		option, err := normalizeRateOption(*out.Data.Cheapest)
		if err != nil {
			return nil, err
		}
		response.Cheapest = &option
	}
	return response, nil
}

func (c *Client) CreateShipment(ctx context.Context, req CreateShipmentRequest) (*ShipmentRecord, error) {
	var out struct {
		Status  string         `json:"status"`
		Message string         `json:"message"`
		Data    ShipmentRecord `json:"data"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/shipping/labels", req, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, fmt.Errorf("shipbubble: create shipment: %s", out.Message)
	}

	rawResponse, err := json.Marshal(out.Data)
	if err != nil {
		return nil, fmt.Errorf("shipbubble: marshal shipment response: %w", err)
	}
	out.Data.Raw = rawResponse
	return &out.Data, nil
}

func normalizeRateOption(raw any) (RateOption, error) {
	bytesValue, err := json.Marshal(raw)
	if err != nil {
		return RateOption{}, fmt.Errorf("shipbubble: marshal rate option: %w", err)
	}

	var payload struct {
		CourierID        any             `json:"courier_id"`
		CourierName      string          `json:"courier_name"`
		CourierImage     string          `json:"courier_image"`
		ServiceCode      string          `json:"service_code"`
		ServiceType      string          `json:"service_type"`
		Waybill          bool            `json:"waybill"`
		OnDemand         bool            `json:"on_demand"`
		IsCODAvailable   bool            `json:"is_cod_available"`
		TrackingLevel    int             `json:"tracking_level"`
		Ratings          float64         `json:"ratings"`
		Votes            int             `json:"votes"`
		ConnectedAccount bool            `json:"connected_account"`
		RateCardAmount   decimal.Decimal `json:"rate_card_amount"`
		RateCardCurrency string          `json:"rate_card_currency"`
		PickupETA        string          `json:"pickup_eta"`
		PickupETATime    string          `json:"pickup_eta_time"`
		DropoffStation   *Station        `json:"dropoff_station"`
		PickupStation    *Station        `json:"pickup_station"`
		DeliveryETA      string          `json:"delivery_eta"`
		DeliveryETATime  string          `json:"delivery_eta_time"`
		Info             []string        `json:"info"`
		Currency         string          `json:"currency"`
		VAT              decimal.Decimal `json:"vat"`
		Total            decimal.Decimal `json:"total"`
		Tracking         TrackingSummary `json:"tracking"`
	}
	if err := json.Unmarshal(bytesValue, &payload); err != nil {
		return RateOption{}, fmt.Errorf("shipbubble: decode rate option: %w", err)
	}

	return RateOption{
		CourierID:        fmt.Sprint(payload.CourierID),
		CourierName:      payload.CourierName,
		CourierImage:     payload.CourierImage,
		ServiceCode:      payload.ServiceCode,
		ServiceType:      payload.ServiceType,
		Waybill:          payload.Waybill,
		OnDemand:         payload.OnDemand,
		IsCODAvailable:   payload.IsCODAvailable,
		TrackingLevel:    payload.TrackingLevel,
		Ratings:          payload.Ratings,
		Votes:            payload.Votes,
		ConnectedAccount: payload.ConnectedAccount,
		RateCardAmount:   payload.RateCardAmount,
		RateCardCurrency: payload.RateCardCurrency,
		PickupETA:        payload.PickupETA,
		PickupETATime:    payload.PickupETATime,
		DropoffStation:   payload.DropoffStation,
		PickupStation:    payload.PickupStation,
		DeliveryETA:      payload.DeliveryETA,
		DeliveryETATime:  payload.DeliveryETATime,
		Info:             payload.Info,
		Currency:         payload.Currency,
		VAT:              payload.VAT,
		Total:            payload.Total,
		Tracking:         payload.Tracking,
		Raw:              bytesValue,
	}, nil
}

func (c *Client) VerifyWebhookSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha512.New, []byte(c.apiKey))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature)))
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
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
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("shipbubble %s %s: status %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
}
