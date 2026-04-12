package models

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PublicStorefront struct {
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	LogoURL      *string `json:"logo_url,omitempty"`
	ContactEmail *string `json:"contact_email,omitempty"`
	ContactPhone *string `json:"contact_phone,omitempty"`
	Address      *string `json:"address,omitempty"`
}

type PublicStorefrontProduct struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Category    *string         `json:"category,omitempty"`
	ImageURL    *string         `json:"image_url,omitempty"`
	Price       decimal.Decimal `json:"price"`
	InStock     bool            `json:"in_stock"`
}

type PublicStorefrontCatalog struct {
	Storefront PublicStorefront          `json:"storefront"`
	Products   []PublicStorefrontProduct `json:"products"`
}

type PublicStorefrontProductVariant struct {
	ID         uuid.UUID       `json:"id"`
	Attributes json.RawMessage `json:"attributes"`
	Price      decimal.Decimal `json:"price"`
	InStock    bool            `json:"in_stock"`
	IsDefault  bool            `json:"is_default"`
}

type PublicStorefrontProductImage struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	SortOrder int       `json:"sort_order"`
	IsPrimary bool      `json:"is_primary"`
}

type PublicStorefrontProductDetail struct {
	Storefront PublicStorefront                 `json:"storefront"`
	Product    PublicStorefrontProduct          `json:"product"`
	Variants   []PublicStorefrontProductVariant `json:"variants"`
	Images     []PublicStorefrontProductImage   `json:"images"`
}

type PublicStorefrontCheckoutOrder struct {
	TrackingSlug      string            `json:"tracking_slug"`
	IsDelivery        bool              `json:"is_delivery"`
	CustomerName      *string           `json:"customer_name,omitempty"`
	CustomerPhone     *string           `json:"customer_phone,omitempty"`
	CustomerEmail     *string           `json:"customer_email,omitempty"`
	ShippingAddress   *string           `json:"shipping_address,omitempty"`
	Note              *string           `json:"note,omitempty"`
	TotalAmount       decimal.Decimal   `json:"total_amount"`
	ShippingFee       decimal.Decimal   `json:"shipping_fee"`
	PaymentMethod     PaymentMethod     `json:"payment_method"`
	PaymentStatus     PaymentStatus     `json:"payment_status"`
	FulfillmentStatus FulfillmentStatus `json:"fulfillment_status"`
}

type PublicStorefrontCheckoutResponse struct {
	Storefront       PublicStorefront              `json:"storefront"`
	Order            PublicStorefrontCheckoutOrder `json:"order"`
	AuthorizationURL string                        `json:"authorization_url,omitempty"`
}
