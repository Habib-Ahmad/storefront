package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type MediaHandler struct {
	accountID string
	apiToken  string
	log       *slog.Logger
}

func NewMediaHandler(accountID, apiToken string, log *slog.Logger) *MediaHandler {
	return &MediaHandler{accountID: accountID, apiToken: apiToken, log: log}
}

// POST /media/upload-url
// Returns a Cloudflare Images one-time direct-upload URL.
// The client POSTs the image file directly to that URL (no credentials needed).
func (h *MediaHandler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	if h.accountID == "" || h.apiToken == "" {
		respondErr(w, http.StatusServiceUnavailable, "image upload not configured")
		return
	}

	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/images/v2/direct_upload",
		h.accountID,
	)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, url, http.NoBody)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+h.apiToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	defer resp.Body.Close()

	var cfResp struct {
		Result struct {
			ID        string `json:"id"`
			UploadURL string `json:"uploadURL"`
		} `json:"result"`
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	if !cfResp.Success {
		respondErr(w, http.StatusBadGateway, "failed to get upload URL from Cloudflare")
		return
	}

	respond(w, http.StatusOK, map[string]string{
		"id":         cfResp.Result.ID,
		"upload_url": cfResp.Result.UploadURL,
	})
}
