package user

import (
	"encoding/json"
	"errors"
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

// Устанавливает cookie с токеном авторизации
func setAuthCookie(w http.ResponseWriter, svc *Service, token string) {
	ttl := int(svc.TokenTTL().Seconds())
	if ttl < 1 {
		ttl = 86400
	}
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   ttl,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func RegisterHTTP(router *http.ServeMux, h *Handler) {
	router.HandleFunc("POST /api/user/register", h.Register)
	router.HandleFunc("POST /api/user/login", h.Login)
}

// Регистрация пользователя
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token, err := h.svc.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		slog.ErrorContext(r.Context(), "register", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	setAuthCookie(w, h.svc, token)
	w.WriteHeader(http.StatusOK)
}

// Вход пользователя
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token, err := h.svc.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		slog.ErrorContext(r.Context(), "login", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	setAuthCookie(w, h.svc, token)
	w.WriteHeader(http.StatusOK)
}
