package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"marketplace/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	usersTable = "users"
	adsTable   = "ads"
)

func NewConnection(cfg config.Database, log *slog.Logger) (*pgxpool.Pool, error) {
	const op = "repository.postgres.NewConnection" // op - operation

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	log.Info("connecting to database", slog.String("op", op))

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully connected to database", slog.String("op", op))

	return pool, nil
}
