package handler

import (
	"net/http"
	"strconv"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/repository"
)

type WalletHandler struct {
	wallets      repository.WalletRepository
	transactions repository.TransactionRepository
}

func NewWalletHandler(wallets repository.WalletRepository, txs repository.TransactionRepository) *WalletHandler {
	return &WalletHandler{wallets: wallets, transactions: txs}
}

// GET /wallet
func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	wallet, err := h.wallets.GetByTenantID(r.Context(), tenant.ID)
	if err != nil {
		respondErr(w, http.StatusNotFound, "wallet not found")
		return
	}
	respond(w, http.StatusOK, wallet)
}

// GET /wallet/transactions?limit=20&offset=0
func (h *WalletHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())

	wallet, err := h.wallets.GetByTenantID(r.Context(), tenant.ID)
	if err != nil {
		respondErr(w, http.StatusNotFound, "wallet not found")
		return
	}

	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	txs, err := h.transactions.ListByWallet(r.Context(), wallet.ID, limit, offset)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "fetch failed")
		return
	}
	respond(w, http.StatusOK, txs)
}

func queryInt(r *http.Request, key string, def int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || v < 0 {
		return def
	}
	return v
}
