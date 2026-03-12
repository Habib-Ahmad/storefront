package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"storefront/backend/internal/service"
)

// chargeSuccessHandler is satisfied by *service.PaymentService.
type chargeSuccessHandler interface {
	HandleChargeSuccess(ctx context.Context, reference string) error
}

// deliveredHandler is satisfied by *service.ShipmentService.
type deliveredHandler interface {
	HandleDelivered(ctx context.Context, orderID uuid.UUID) error
}

type webhookVerifier interface {
	VerifyWebhookSignature(payload []byte, signature string) bool
}

type WebhookHandler struct {
	paystackClient webhookVerifier
	terminalClient webhookVerifier
	paymentSvc     chargeSuccessHandler
	shipmentSvc    deliveredHandler
	log            *slog.Logger
}

func NewWebhookHandler(
	paystackClient webhookVerifier,
	terminalClient webhookVerifier,
	paymentSvc chargeSuccessHandler,
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
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
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
		if err := json.Unmarshal(event.Data, &data); err == nil && data.Reference != "" {
			if err := h.paymentSvc.HandleChargeSuccess(r.Context(), data.Reference); err != nil {
				if !errors.Is(err, service.ErrAlreadyPaid) {
					h.log.Error("paystack webhook: charge.success", "error", err)
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// POST /webhooks/terminalaf
func (h *WebhookHandler) TerminalAf(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
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

	if event.Event == "shipment.delivered" {
		var data struct {
			Reference string `json:"metadata_reference"`
		}
		if err := json.Unmarshal(event.Data, &data); err == nil && data.Reference != "" {
			if orderID, err := uuid.Parse(data.Reference); err == nil {
				if err := h.shipmentSvc.HandleDelivered(r.Context(), orderID); err != nil {
					h.log.Error("terminalaf webhook: shipment.delivered", "error", err)
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
