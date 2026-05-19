package httpserver

import (
	"log/slog"
	"net/http"
	"time"

	"gophermart/internal/balance"
	"gophermart/internal/order"
	"gophermart/internal/user"
	"gophermart/pkg/middleware"
)

type Server struct {
	users   *user.Service
	orders  *order.Service
	balance *balance.Service
}

func New(users *user.Service, orders *order.Service, balance *balance.Service) *Server {
	return &Server{users: users, orders: orders, balance: balance}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

// Записывает статус код в response writer
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Возвращает статус код
func (rw *responseWriter) statusCode() int {
	if rw.status == 0 {
		return http.StatusOK
	}
	return rw.status
}

// Создаёт router
func (s *Server) Router() http.Handler {
	router := http.NewServeMux()
	auth := middleware.Auth(s.users)

	user.RegisterHTTP(router, user.NewHandler(user.HandlerDeps{Service: s.users}))
	order.RegisterHTTP(router, auth, order.NewHandler(order.HandlerDeps{Service: s.orders}))
	balance.RegisterHTTP(router, auth, balance.NewHandler(balance.HandlerDeps{Service: s.balance}))

	return s.logging(router)
}

// Логирует запрос
func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(wrapped, r)
		slog.InfoContext(r.Context(), "http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.statusCode()),
			slog.Duration("duration", time.Since(start)),
		)
	})
}
