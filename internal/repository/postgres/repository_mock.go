package postgres

import (
	"context"
	"marketplace/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository является мок-реализацией UserRepository.
type MockUserRepository struct {
	mock.Mock
}

// CreateUser симулирует создание пользователя.
func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) (int64, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(int64), args.Error(1)
}

// GetUserByUsername симулирует получение пользователя по имени.
func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	// Позволяет вернуть nil, если пользователь не найден
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MockAdRepository является мок-реализацией AdRepository.
type MockAdRepository struct {
	mock.Mock
}

// CreateAd симулирует создание объявления.
func (m *MockAdRepository) CreateAd(ctx context.Context, ad *models.Ad) (int64, error) {
	args := m.Called(ctx, ad)
	return args.Get(0).(int64), args.Error(1)
}

// GetAllAds симулирует получение всех объявлений.
func (m *MockAdRepository) GetAllAds(ctx context.Context, params GetAllAdsParams) ([]models.Ad, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Ad), args.Error(1)
}

// GetAdByID симулирует получение объявления по ID.
func (m *MockAdRepository) GetAdByID(ctx context.Context, id int64) (*models.Ad, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ad), args.Error(1)
}

// UpdateAd симулирует обновление объявления.
func (m *MockAdRepository) UpdateAd(ctx context.Context, ad *models.Ad) error {
	args := m.Called(ctx, ad)
	return args.Error(0)
}

// DeleteAd симулирует удаление объявления.
func (m *MockAdRepository) DeleteAd(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}
