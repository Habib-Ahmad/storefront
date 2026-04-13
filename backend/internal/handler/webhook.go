package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// chargeHandler is satisfied by *service.PaymentService.
type chargeHandler interface {
	HandleChargeSuccess(ctx context.Context, reference string) error
	HandleChargeFailed(ctx context.Context, reference string) error
}

// deliveredHandler is satisfied by *service.ShipmentService.
type deliveredHandler interface {
	HandleDelivered(ctx context.Context, orderID uuid.UUID) error
	HandleShipmentFailed(ctx context.Context, orderID uuid.UUID) error
}

type webhookVerifier interface {
	VerifyWebhookSignature(payload []byte, signature string) bool
}

type WebhookHandler struct {
	paystackClient webhookVerifier
	terminalClient webhookVerifier
	paymentSvc     chargeHandler
	shipmentSvc    deliveredHandler
	log            *slog.Logger
}

func NewWebhookHandler(
	paystackClient webhookVerifier,
	terminalClient webhookVerifier,
	paymentSvc chargeHandler,
	shipmentSvc deliveredHandler,
	log *slog.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		paystackClient: paystackClient,
		terminalClient: terminalClient,
		paymentSvc:     paymentSvc,
		shipmentSvc:    shipmentSvc,
		log:            log,
	}
}

type incomingWebhookEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// POST /webhooks/paystack
func (h *WebhookHandler) Paystack(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10) // 64KB
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "cannot read body")
		return
	}

	if !h.paystackClient.VerifyWebhookSignature(body, r.Header.Get("X-Paystack-Signature")) {
		respondErr(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	var event incomingWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid event payload")
		return
	}

	if event.Event == "charge.success" {
		var data struct {
			Reference string `json:"reference"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil || data.Reference == "" {
			respondErr(w, http.StatusBadRequest, "invalid charge.success payload")
			return
		}
		if err := h.paymentSvc.HandleChargeSuccess(r.Context(), data.Reference); err != nil {
			h.log.Error("paystack webhook: charge.success", "reference", data.Reference, "error", err)
			respondErr(w, http.StatusInternalServerError, "webhook processing failed")
			return
		}
	}

	if event.Event == "charge.failed" {
		var data struct {
			Reference string `json:"reference"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil || data.Reference == "" {
			respondErr(w, http.StatusBadRequest, "invalid charge.failed payload")
			return
		}
		if err := h.paymentSvc.HandleChargeFailed(r.Context(), data.Reference); err != nil {
			h.log.Error("paystack webhook: charge.failed", "reference", data.Reference, "error", err)
			respondErr(w, http.StatusInternalServerError, "webhook processing failed")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// POST /webhooks/terminalaf
func (h *WebhookHandler) TerminalAf(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10) // 64KB
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "cannot read body")
		return
	}

	if !h.terminalClient.VerifyWebhookSignature(body, r.Header.Get("X-Terminal-Africa-Signature")) {
		respondErr(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	var event incomingWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid event payload")
		return
	}

	if event.Event == "shipment.delivered" || event.Event == "shipment.failed" {
		var data struct {
			Reference string `json:"metadata_reference"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil || data.Reference == "" {
			respondErr(w, http.StatusBadRequest, "invalid shipment event payload")
			return
		}
		orderID, err := uuid.Parse(data.Reference)
		if err != nil {
			respondErr(w, http.StatusBadRequest, "invalid shipment reference")
			return
		}
		switch event.Event {
		case "shipment.delivered":
			if err := h.shipmentSvc.HandleDelivered(r.Context(), orderID); err != nil {
				h.log.Error("terminalaf webhook: shipment.delivered", "order_id", orderID, "error", err)
				respondErr(w, http.StatusInternalServerError, "webhook processing failed")
				return
			}
		case "shipment.failed":
			if err := h.shipmentSvc.HandleShipmentFailed(r.Context(), orderID); err != nil {
				h.log.Error("terminalaf webhook: shipment.failed", "order_id", orderID, "error", err)
				respondErr(w, http.StatusInternalServerError, "webhook processing failed")
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
