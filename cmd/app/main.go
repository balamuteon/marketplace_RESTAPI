// cmd/app/main.go
package main

import (
	"context"
	"log/slog"
	"marketplace/internal/config"
	"marketplace/internal/handler"
	"marketplace/internal/repository/postgres"
	"marketplace/internal/service"
	"marketplace/pkg/auth"
	"marketplace/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title Marketplace API
// @version 1.0
// @description API для учебного проекта торговой площадки.
// @contact.name Lev Vinogradov
// @contact.url http://github.com/your-profile
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Для доступа к защищенным эндпоинтам, укажите токен в формате "Bearer <ваш_токен>"
func main() {
	cfg := config.LoadConfig()

	log := logger.NewLogger(cfg.Env)
	log.Info("starting marketplace application", slog.String("env", cfg.Env))

	dbPool, err := postgres.NewConnection(cfg.Database, log)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	tokenManager, err := auth.NewTokenManager(cfg.Auth.JWTSecret)
	if err != nil {
		log.Error("failed to init token manager", slog.String("error", err.Error()))
		os.Exit(1)
	}

	repos := postgres.NewRepository(dbPool)
	services := service.NewService(repos, tokenManager, cfg.Auth.TokenTTL)
	handlers := handler.NewHandler(services, tokenManager, log)
	router := handlers.InitRoutes()

	log.Info("starting server", slog.String("port", cfg.HTTPServer.Port))

	server := &http.Server{
		Addr:         ":" + cfg.HTTPServer.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("server exited properly")
}
