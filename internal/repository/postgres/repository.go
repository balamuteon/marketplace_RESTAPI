package postgres

import (
	"context"
	"marketplace/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (int64, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type AdRepository interface {
	CreateAd(ctx context.Context, ad *models.Ad) (int64, error)
	GetAllAds(ctx context.Context, params GetAllAdsParams) ([]models.Ad, error)
	GetAdByID(ctx context.Context, id int64) (*models.Ad, error)
	UpdateAd(ctx context.Context, ad *models.Ad) error
	DeleteAd(ctx context.Context, id, userID int64) error
}

type Repository struct {
	User UserRepository
	Ad   AdRepository
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		User: NewUserRepository(db),
		Ad:   NewAdRepository(db),
	}
}
