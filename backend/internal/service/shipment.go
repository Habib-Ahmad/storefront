package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"storefront/backend/internal/adapter/terminalaf"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

// CarrierClient abstracts the logistics provider (Terminal Africa primary; swap point for Shipbubble).
type CarrierClient interface {
	BookShipment(ctx context.Context, req terminalaf.BookRequest) (*terminalaf.BookResponse, error)
}

type ShipmentService struct {
	carrier   CarrierClient
	shipments repository.ShipmentRepository
	orders    repository.OrderRepository
	walletSvc *WalletService
	tenants   repository.TenantRepository
	tiers     repository.TierRepository
}

func NewShipmentService(
	carrier CarrierClient,
	shipments repository.ShipmentRepository,
	orders repository.OrderRepository,
	walletSvc *WalletService,
	tenants repository.TenantRepository,
	tiers repository.TierRepository,
) *ShipmentService {
	return &ShipmentService{carrier: carrier, shipments: shipments, orders: orders, walletSvc: walletSvc, tenants: tenants, tiers: tiers}
}

// Dispatch books a shipment with the carrier and persists the booking to the shipments table.
func (s *ShipmentService) Dispatch(ctx context.Context, orderID, tenantID uuid.UUID, req terminalaf.BookRequest) (*models.Shipment, error) {
	resp, err := s.carrier.BookShipment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("book shipment: %w", err)
	}

	history, _ := json.Marshal(resp)
	shipment := &models.Shipment{
		OrderID:        orderID,
		TenantID:       tenantID,
		CarrierRef:     &resp.CarrierRef,
		TrackingNumber: &resp.TrackingNumber,
		CarrierHistory: history,
	}
	if err := s.shipments.Create(ctx, shipment); err != nil {
		return nil, fmt.Errorf("save shipment: %w", err)
	}

	if err := s.orders.UpdateFulfillmentStatus(ctx, tenantID, orderID, models.FulfillmentStatusShipped); err != nil {
		return nil, fmt.Errorf("update fulfillment status: %w", err)
	}

	return shipment, nil
}

// HandleDelivered processes a Terminal Africa delivery webhook:
// updates shipment + order status, then releases the order amount from pending balance.
// orderID is extracted from the booking metadata_reference field echoed in the webhook.
func (s *ShipmentService) HandleDelivered(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.orders.GetByIDInternal(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	shipment, err := s.shipments.GetByOrderID(ctx, order.TenantID, orderID)
	if err != nil {
		return fmt.Errorf("get shipment: %w", err)
	}

	if err := s.shipments.UpdateStatus(ctx, order.TenantID, shipment.ID, models.ShipmentStatusDelivered); err != nil {
		return fmt.Errorf("update shipment status: %w", err)
	}

	if err := s.orders.UpdateFulfillmentStatus(ctx, order.TenantID, orderID, models.FulfillmentStatusDelivered); err != nil {
		return fmt.Errorf("update fulfillment status: %w", err)
	}

	// Release the net amount (gross − commission) from pending to available.
	gross := order.TotalAmount.Add(order.ShippingFee)
	tenant, err := s.tenants.GetByID(ctx, order.TenantID)
	if err != nil {
		return fmt.Errorf("get tenant: %w", err)
	}
	tier, err := s.tiers.GetByID(ctx, tenant.TierID)
	if err != nil {
		return fmt.Errorf("get tier: %w", err)
	}
	commission := gross.Mul(tier.CommissionRate)
	netAmount := gross.Sub(commission)
	if err := s.walletSvc.ReleasePending(ctx, order.TenantID, netAmount); err != nil {
		return fmt.Errorf("release pending: %w", err)
	}

	return nil
}
