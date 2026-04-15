package service

import (
	"strings"

	"storefront/backend/internal/models"
)

func PublicStorefrontFromTenant(tenant *models.Tenant) models.PublicStorefront {
	return models.PublicStorefront{
		Name:         tenant.Name,
		Slug:         tenant.Slug,
		LogoURL:      tenant.LogoURL,
		ContactEmail: tenant.ContactEmail,
		ContactPhone: tenant.ContactPhone,
		Address:      tenant.Address,
		Delivery:     StorefrontDeliveryStatus(tenant),
	}
}

func StorefrontDeliveryStatus(tenant *models.Tenant) models.PublicStorefrontDeliveryStatus {
	status := models.PublicStorefrontDeliveryStatus{}
	if tenant == nil {
		reason := "Delivery is temporarily unavailable while the store completes its pickup profile."
		status.UnavailableReason = &reason
		return status
	}

	status.Enabled = tenant.ActiveModules.Logistics
	if !status.Enabled {
		reason := "This store has not enabled delivery yet."
		status.UnavailableReason = &reason
		return status
	}

	if !StorefrontDeliveryReady(tenant) {
		reason := "Delivery is temporarily unavailable while the store completes its pickup profile."
		status.UnavailableReason = &reason
		return status
	}

	status.Ready = true
	return status
}

func StorefrontDeliveryReady(tenant *models.Tenant) bool {
	if tenant == nil {
		return false
	}

	return trimmedTenantValue(tenant.ContactEmail) != "" &&
		trimmedTenantValue(tenant.ContactPhone) != "" &&
		hasCompleteLogisticsAddress(tenant.Address)
}

func hasCompleteLogisticsAddress(address *string) bool {
	trimmed := trimmedTenantValue(address)
	if trimmed == "" {
		return false
	}

	segments := make([]string, 0, 4)
	for _, segment := range strings.Split(trimmed, ",") {
		segment = strings.TrimSpace(segment)
		if segment != "" {
			segments = append(segments, segment)
		}
	}

	return len(segments) >= 4
}

func trimmedTenantValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
