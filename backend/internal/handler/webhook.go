package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// chargeHandler is satisfied by *service.PaymentService.
type chargeHandler interface {
	HandleChargeSuccess(ctx context.Context, reference string) error
	HandleChargeFailed(ctx context.Context, reference string) error
}

type webhookVerifier interface {
	VerifyWebhookSignature(payload []byte, signature string) bool
}

type WebhookHandler struct {
	paystackClient webhookVerifier
	paymentSvc     chargeHandler
	log            *slog.Logger
}

func NewWebhookHandler(
	paystackClient webhookVerifier,
	paymentSvc chargeHandler,
	log *slog.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		paystackClient: paystackClient,
		paymentSvc:     paymentSvc,
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
