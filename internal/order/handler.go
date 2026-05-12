package order

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

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
	router.Handle("POST /api/user/orders", auth(http.HandlerFunc(h.UploadOrder)))
	router.Handle("GET /api/user/orders", auth(http.HandlerFunc(h.ListOrders)))
}

// Получает user id из контекста
func userIDFromContext(ctx context.Context) (int64, bool) {
	uid, ok := middleware.UserIDFromContext(ctx)
	return uid, ok
}

// Загружает заказ
func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.WarnContext(r.Context(), "upload order read body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	number := strings.TrimSpace(string(body))
	if number == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	statusCode, err := h.svc.SubmitOrder(r.Context(), uid, number)
	if err != nil {
		slog.ErrorContext(r.Context(), "submit order", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
}

// Получает список заказов
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	orders, err := h.svc.ListOrders(r.Context(), uid)
	if err != nil {
		slog.ErrorContext(r.Context(), "list orders", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ListOrdersResponse(orders)); err != nil {
		slog.ErrorContext(r.Context(), "encode list orders", slog.Any("error", err))
	}
}
