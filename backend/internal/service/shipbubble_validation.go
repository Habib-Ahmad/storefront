package service

import (
	"fmt"
	"strings"
	"unicode"

	"storefront/backend/internal/apperr"
)

func shipbubbleSenderName(name, slug string) string {
	cleanedName := sanitizeShipbubbleName(name)
	if len(strings.Fields(cleanedName)) >= 2 {
		return cleanedName
	}

	cleanedSlug := sanitizeShipbubbleName(strings.ReplaceAll(slug, "-", " "))
	if len(strings.Fields(cleanedSlug)) >= 2 {
		return cleanedSlug
	}

	if cleanedName != "" {
		return cleanedName + " Store"
	}
	if cleanedSlug != "" {
		return cleanedSlug + " Store"
	}
	return "Storefront Delivery"
}

func shipbubbleReceiverName(name string) string {
	cleanedName := sanitizeShipbubbleName(name)
	if len(strings.Fields(cleanedName)) >= 2 {
		return cleanedName
	}
	if cleanedName != "" {
		return cleanedName + " Customer"
	}
	return "Delivery Customer"
}

func sanitizeShipbubbleName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var builder strings.Builder
	lastSpace := true
	for _, r := range value {
		switch {
		case unicode.IsLetter(r):
			builder.WriteRune(r)
			lastSpace = false
		case unicode.IsSpace(r) || r == '-' || r == '_':
			if !lastSpace && builder.Len() > 0 {
				builder.WriteByte(' ')
				lastSpace = true
			}
		case r == '\'' || r == '’' || r == '.':
			continue
		default:
			if !lastSpace && builder.Len() > 0 {
				builder.WriteByte(' ')
				lastSpace = true
			}
		}
	}

	return strings.TrimSpace(builder.String())
}

func mapShipbubbleValidationError(role, senderAddressMessage, receiverAddressMessage string, err error) error {
	message := strings.ToLower(strings.TrimSpace(err.Error()))

	switch {
	case strings.Contains(message, "insufficient wallet balance"):
		return apperr.Conflict("delivery is temporarily unavailable because the shipping provider wallet needs funding")
	case strings.Contains(message, "full name") || strings.Contains(message, "remove all numbers and symbols"):
		if role == "sender" {
			return apperr.Unprocessable("the store pickup contact name is not accepted by the delivery provider")
		}
		return apperr.Unprocessable("the customer name is not accepted by the delivery provider")
	case strings.Contains(message, "address/validate") || strings.Contains(message, "validate address") || strings.Contains(message, "couldn't validate the provided address"):
		if role == "sender" {
			return apperr.Unprocessable(senderAddressMessage)
		}
		return apperr.Unprocessable(receiverAddressMessage)
	default:
		return fmt.Errorf("validate %s address: %w", role, err)
	}
}
