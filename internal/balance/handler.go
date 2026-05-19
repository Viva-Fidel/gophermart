package balance

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"gophermart/pkg/middleware"
)

type Handler struct {
	svc *Service
}

type HandlerDeps struct {
	Service *Service
}

func NewHandler(deps HandlerDeps) *Handler {
	return &Handler{svc: deps.Service}
}

func RegisterHTTP(router *http.ServeMux, auth func(http.Handler) http.Handler, h *Handler) {
	router.Handle("GET /api/user/balance", auth(http.HandlerFunc(h.GetBalance)))
	router.Handle("POST /api/user/balance/withdraw", auth(http.HandlerFunc(h.Withdraw)))
	router.Handle("GET /api/user/withdrawals", auth(http.HandlerFunc(h.ListWithdrawals)))
}

// Получает баланс пользователя
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	bal, err := h.svc.GetBalance(r.Context(), uid)
	if err != nil {
		slog.ErrorContext(r.Context(), "get balance", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(bal); err != nil {
		slog.ErrorContext(r.Context(), "encode balance", slog.Any("error", err))
	}
}

// Списывает средства с баланса пользователя
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Order == "" || req.Sum <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	statusCode, err := h.svc.Withdraw(r.Context(), uid, req.Order, req.Sum)
	if err != nil {
		slog.ErrorContext(r.Context(), "withdraw", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
}

// Получает список выводов средств
func (h *Handler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	list, err := h.svc.ListWithdrawals(r.Context(), uid)
	if err != nil {
		slog.ErrorContext(r.Context(), "list withdrawals", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ListWithdrawalsResponse(list)); err != nil {
		slog.ErrorContext(r.Context(), "encode withdrawals", slog.Any("error", err))
	}
}
