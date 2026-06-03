// @title Effective Mobile Subscriptions API
// @version 1.0
// @description REST-сервис для агрегации данных об онлайн-подписках пользователей
// @host localhost:8080
// @BasePath /
package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"effective-mobile/internal/config"
	"effective-mobile/internal/handler"
	"effective-mobile/internal/migrate"
	"effective-mobile/internal/repository"
	"effective-mobile/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.LoadConfig()

	level := parseLogLevel(cfg.LogLevel)

	// Открываем файл для логов (создаётся в корне проекта)
	logFile, err := os.OpenFile("subscriptions.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		os.Exit(1)
	}

	// Пишем логи одновременно в консоль и в файл
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{Level: level}))

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("failed to open database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		logger.Error("database is unreachable", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("database connected successfully")

	// === ЗАПУСК МИГРАЦИЙ ===
	// Путь "./migrations" указан относительно точки запуска (корень проекта)
	// Если запускаешь из другой папки, лучше использовать абсолютный путь или config
	if err := migrate.Run(db, logger, "./migrations"); err != nil {
		logger.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
	// =======================

	repo := repository.NewSubscriptionRepository(db, logger)
	svc := service.NewSubscriptionService(repo, logger)
	h := handler.NewSubscriptionHandler(svc, logger)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/subscriptions", h.CreateSubscriptions)                   // CREATE
	r.Get("/subscriptions", h.ListSubscriptions)                      // LIST
	r.Put("/subscriptions", h.UpdateSubscriptions)                    // UPDATE
	r.Delete("/subscriptions", h.DeleteSubscriptions)                 // DELETE
	r.Get("/subscriptions/{user_subscription_id}", h.GetSubscription) // READ ONE
	r.Post("/subscriptions/total", h.AggregateSubscriptions)          // AGGREGATE

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logger.Info("server started", slog.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", slog.String("error", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Закрываем файл логов перед завершением
	_ = logFile.Close()

	logger.Info("shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("server stopped gracefully")
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
