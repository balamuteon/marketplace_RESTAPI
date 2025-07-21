package postgres

import (
	"context"
	"errors"
	"fmt"
	"marketplace/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) (int64, error) {
	query := fmt.Sprintf(`INSERT INTO %s (username, password_hash) VALUES ($1, $2) RETURNING id`, usersTable)
	var id int64
	err := r.db.QueryRow(ctx, query, user.Username, user.Password).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("repository.CreateUser: %w", err)
	}
	return id, nil
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := fmt.Sprintf(`SELECT id, username, password_hash, created_at, updated_at 
												FROM %s WHERE username = $1`, usersTable)
	var user models.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository.GetUserByUsername: %w", err)
	}
	return &user, nil
}
