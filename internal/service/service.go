package service

import (
	"context"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"
	"marketplace/pkg/auth"
	"time"
)

type AdService interface {
	CreateAd(ctx context.Context, ad *models.Ad) (int64, error)
	GetAllAds(ctx context.Context, params postgres.GetAllAdsParams) ([]models.Ad, error)
	GetAdByID(ctx context.Context, id int64) (*models.Ad, error)
	UpdateAd(ctx context.Context, id, userID int64, req models.UpdateAdRequest) (*models.Ad, error)
	DeleteAd(ctx context.Context, id, userID int64) error
}

type AuthService interface {
	Register(ctx context.Context, username, password string) (*models.User, error)
	Login(ctx context.Context, username, password string) (string, error)
}

type Service struct {
	Auth AuthService
	Ad   AdService
}

type Deps struct {
	UserRepo     postgres.UserRepository
	AdRepo       postgres.AdRepository
	TokenManager *auth.TokenManager
	TokenTTL     time.Duration
}

func NewService(repos *postgres.Repository, tm *auth.TokenManager, ttl time.Duration) *Service {
	return &Service{
		Auth: NewAuthService(repos.User, tm, ttl),
		Ad:   NewAdService(repos.Ad),
	}
}
