package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"marketplace/docs"
	"marketplace/internal/config"
	"marketplace/internal/handler"
	cache "marketplace/internal/repository/cache"
	"marketplace/internal/repository/postgres"
	"marketplace/internal/service"
	"marketplace/pkg/auth"
	redis "marketplace/pkg/cache"
	"marketplace/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	log         *slog.Logger
	server      *http.Server
	dbPool      *pgxpool.Pool
	redisClient *redis.CacheClient
}

// New создает новый экземпляр приложения со всеми зависимостями.
func New() (*App, error) {
	// 1. Инициализация конфига и логгера
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.Env)
	log.Info("starting application initialization")

	// 2. Настройка Swagger
	setupSwagger(cfg)

	// 3. Инициализация зависимостей (БД, менеджер токенов)
	dbPool, err := initDB(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}

	// 4. Подключение к Redis
	redisClient, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("failed to init cache: %w", err)
	}

	// 5. Применение миграций
	if err := runMigrations(cfg, log); err != nil {
		dbPool.Close()
		redisClient.Client.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// 6. Инициализация менеджера токенов
	tokenManager, err := auth.NewTokenManager(cfg.Auth)
	if err != nil {
		dbPool.Close()
		redisClient.Client.Close()
		return nil, fmt.Errorf("failed to init token manager: %w", err)
	}

	// 7. Инициализация роутера
	router := initRouter(dbPool, redisClient, tokenManager, cfg, log)

	// 8. Настройка HTTP-сервера
	server := initServer(cfg, router)

	return &App{
		log:         log,
		server:      server,
		dbPool:      dbPool,
		redisClient: redisClient,
	}, nil
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

	defer a.dbPool.Close()
	defer a.redisClient.Client.Close()

	a.log.Info("database connection pool closed")

	a.log.Info("server exited properly")
}

// runMigrations применяет миграции базы данных при старте приложения.
func runMigrations(cfg *config.Config, log *slog.Logger) error {
	sslMode := "disable"
	if cfg.Env != "local" {
		sslMode = "require"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		sslMode,
	)

	log.Info("applying database migrations...")

	m, err := migrate.New("file:///app/migrations", dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Info("database migrations applied successfully")
	return nil
}

// setupSwagger настраивает статическую информацию для документации.
func setupSwagger(cfg *config.Config) {
	docs.SwaggerInfo.Title = "Marketplace API"
	docs.SwaggerInfo.Description = "API для учебного проекта торговой площадки."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = cfg.Swagger.Host
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	if cfg.Env == "local" {
		docs.SwaggerInfo.Schemes = []string{"http"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"https"}
	}
}

// initDB инициализирует подключение к базе данных.
func initDB(cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	dbPool, err := postgres.NewConnection(cfg.Database, log)
	if err != nil {
		return nil, err // Просто возвращаем ошибку
	}
	// Ping остается для проверки
	if err := dbPool.Ping(context.Background()); err != nil {
		return nil, err
	}
	return dbPool, err
}

// initRouter собирает все слои приложения и инициализирует роутер.
func initRouter(dbPool *pgxpool.Pool, redis *redis.CacheClient, tm *auth.TokenManager, cfg *config.Config, log *slog.Logger) *gin.Engine {
	// 1. Создаем основной репозиторий, который работает с PostgreSQL.
	postgresRepos := postgres.NewRepository(dbPool)

	// 2. "Оборачиваем" репозиторий объявлений кеширующим декоратором.
	cachedAdRepo := cache.NewAdRepository(postgresRepos.Ad, redis)

	// 3. Создаем "обертку" для репозиториев, где Ad заменен на кеширующий.
	finalRepos := &postgres.Repository{
		User: postgresRepos.User,
		Ad:   cachedAdRepo,
	}

	// 4. Передаем итоговый набор репозиториев в сервис.
	// AdService теперь будет работать с кеширующей версией, даже не зная об этом.
	services := service.NewService(finalRepos, tm)
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
