package models

import (
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
