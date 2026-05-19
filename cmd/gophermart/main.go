package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gophermart/configs"
	"gophermart/internal/balance"
	"gophermart/internal/httpserver"
	"gophermart/internal/loyalty"
	"gophermart/internal/order"
	"gophermart/internal/user"
	"gophermart/pkg/db"
)

func main() {
	// Устанавливает логгер
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// Загружает флаги и конфигурацию
	conf, err := config.LoadFlags()
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Открывает соединение с БД
	database, err := db.Open(ctx, conf.DatabaseURI)
	if err != nil {
		slog.Error("db open", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			slog.Error("db close", slog.Any("error", err))
		}
	}()

	// Применяет миграции
	if err := db.RunMigrations(ctx, database, "migrations"); err != nil {
		slog.Error("run migrations", slog.Any("error", err))
		os.Exit(1)
	}

	// Инициализирует репозитории
	userRepo := user.NewUserRepository(database)
	orderRepo := order.NewOrderRepository(database)
	balanceRepo := balance.NewBalanceRepository(database)

	tokenTTL, err := time.ParseDuration(conf.TokenExp)
	if err != nil {
		slog.Error("parse token TTL", slog.Any("error", err))
		os.Exit(1)
	}

	// Инициализирует сервисы
	userSvc := user.New(userRepo, conf.JWTSecret, tokenTTL)
	orderSvc := order.New(orderRepo)
	balanceSvc := balance.New(balanceRepo)

	// Инициализирует worker
	worker := loyalty.NewWorker(orderRepo, loyalty.NewClient(conf.AccrualSystemAdress))

	// Инициализирует router
	handler := httpserver.New(userSvc, orderSvc, balanceSvc).Router()

	// Инициализирует listener
	ln, err := net.Listen("tcp", conf.RunAddress)
	if err != nil {
		slog.Error("listen", slog.Any("error", err))
		os.Exit(1)
	}

	// Инициализирует HTTP сервер
	httpSrv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Запускает worker
	workerDone := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(workerDone)
	}()

	// Запускает HTTP сервер
	go func() {
		if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
			slog.Error("http serve", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	slog.Info("gophermart started", slog.String("addr", ln.Addr().String()))

	// Ожидает сигнала завершения
	<-ctx.Done()

	// Создаёт контекст с таймаутом для завершения работы
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// Отменяет контекст при завершении работы
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown", slog.Any("error", err))
	}

	// Ожидает завершения работы worker
	<-workerDone
	slog.Info("gophermart stopped")
	if err := os.Stdout.Sync(); err != nil {
		slog.Error("stdout sync", slog.Any("error", err))
	}
}
