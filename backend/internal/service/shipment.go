package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"storefront/backend/internal/adapter/terminalaf"
	"storefront/backend/internal/apperr"
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
}

func NewShipmentService(
	carrier CarrierClient,
	shipments repository.ShipmentRepository,
	orders repository.OrderRepository,
	walletSvc *WalletService,
) *ShipmentService {
	return &ShipmentService{carrier: carrier, shipments: shipments, orders: orders, walletSvc: walletSvc}
}

// Dispatch books a shipment with the carrier and persists the booking to the shipments table.
// Only allowed when the order is still in processing state (initial or after a failed shipment).
func (s *ShipmentService) Dispatch(ctx context.Context, orderID, tenantID uuid.UUID, req terminalaf.BookRequest) (*models.Shipment, error) {
	order, err := s.orders.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	if order.FulfillmentStatus != models.FulfillmentStatusProcessing {
		return nil, apperr.Unprocessable("order is not in a dispatchable state")
	}

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

	// Idempotency: if already delivered, skip.
	if order.FulfillmentStatus == models.FulfillmentStatusDelivered {
		return nil
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

	// Release total order value (total + shipping) from pending to available balance.
	amount := order.TotalAmount.Add(order.ShippingFee)
	if err := s.walletSvc.ReleasePending(ctx, order.TenantID, amount, &orderID); err != nil {
		return fmt.Errorf("release pending: %w", err)
	}

	return nil
}

// HandleShipmentFailed marks the shipment as failed and resets the order to processing
// so that the user can re-dispatch.
func (s *ShipmentService) HandleShipmentFailed(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.orders.GetByIDInternal(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	shipment, err := s.shipments.GetByOrderID(ctx, order.TenantID, orderID)
	if err != nil {
		return fmt.Errorf("get shipment: %w", err)
	}

	if err := s.shipments.UpdateStatus(ctx, order.TenantID, shipment.ID, models.ShipmentStatusFailed); err != nil {
		return fmt.Errorf("update shipment status: %w", err)
	}

	if err := s.orders.UpdateFulfillmentStatus(ctx, order.TenantID, orderID, models.FulfillmentStatusProcessing); err != nil {
		return fmt.Errorf("reset fulfillment status: %w", err)
	}

	return nil
}
