package service

import (
	"context"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"

	"github.com/stretchr/testify/mock"
)

// MockAuthService является мок-реализацией AuthService.
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, username, password string) (*models.User, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

// MockAdService является мок-реализацией AdService.
type MockAdService struct {
	mock.Mock
}

func (m *MockAdService) CreateAd(ctx context.Context, ad *models.Ad) (int64, error) {
	args := m.Called(ctx, ad)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAdService) GetAllAds(ctx context.Context, params postgres.GetAllAdsParams) ([]models.Ad, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Ad), args.Error(1)
}

func (m *MockAdService) GetAdByID(ctx context.Context, id int64) (*models.Ad, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ad), args.Error(1)
}

func (m *MockAdService) UpdateAd(ctx context.Context, id, userID int64, req models.UpdateAdRequest) (*models.Ad, error) {
	args := m.Called(ctx, id, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ad), args.Error(1)
}

func (m *MockAdService) DeleteAd(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}
