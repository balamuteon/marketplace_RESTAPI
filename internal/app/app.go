package app

import (
	"context"
	"log/slog"
	"marketplace/docs"
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

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	log    *slog.Logger
	server *http.Server
	dbPool *pgxpool.Pool
}

// New создает новый экземпляр приложения со всеми зависимостями.
func New() *App {
	// 1. Инициализация конфига и логгера
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.Env)
	log.Info("starting application initialization")

	// 2. Настройка Swagger
	setupSwagger(cfg)

	// 3. Инициализация зависимостей (БД, менеджер токенов)
	dbPool := initDB(cfg, log)
	tokenManager := initTokenManager(cfg, log)

	// 4. Сборка слоев приложения и роутера
	router := initRouter(dbPool, tokenManager, cfg, log)

	// 5. Настройка HTTP-сервера
	server := initServer(cfg, router)

	return &App{
		log:    log,
		server: server,
		dbPool: dbPool,
	}
}

// Run запускает HTTP-сервер и обрабатывает graceful shutdown.
func (a *App) Run() {
	a.log.Info("starting server", slog.String("addr", a.server.Addr))

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("failed to start server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("server shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	a.dbPool.Close()
	a.log.Info("database connection pool closed")

	a.log.Info("server exited properly")
}

// setupSwagger настраивает статическую информацию для документации.
func setupSwagger(cfg *config.Config) {
	docs.SwaggerInfo.Title = "Marketplace API"
	docs.SwaggerInfo.Description = "API для учебного проекта торговой площадки."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080" // TODO: Вынести в конфиг
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}
}

// initDB инициализирует подключение к базе данных.
func initDB(cfg *config.Config, log *slog.Logger) *pgxpool.Pool {
	dbPool, err := postgres.NewConnection(cfg.Database, log)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	return dbPool
}

// initTokenManager инициализирует менеджер JWT.
func initTokenManager(cfg *config.Config, log *slog.Logger) *auth.TokenManager {
	tokenManager, err := auth.NewTokenManager(cfg.Auth.JWTSecret)
	if err != nil {
		log.Error("failed to init token manager", slog.String("error", err.Error()))
		os.Exit(1)
	}
	return tokenManager
}

// initRouter собирает все слои приложения и инициализирует роутер.
func initRouter(dbPool *pgxpool.Pool, tm *auth.TokenManager, cfg *config.Config, log *slog.Logger) *gin.Engine {
	repos := postgres.NewRepository(dbPool)
	services := service.NewService(repos, tm, cfg.Auth.TokenTTL)
	handlers := handler.NewHandler(services, tm, log)
	return handlers.InitRoutes()
}

// initServer настраивает HTTP-сервер.
func initServer(cfg *config.Config, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:         ":" + cfg.HTTPServer.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
}
