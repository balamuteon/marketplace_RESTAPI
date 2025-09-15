package service

import (
	"context"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Тестирование успешного создания объявления
func TestAdService_CreateAd_Success(t *testing.T) {
	// 1. Настройка
	mockAdRepo := new(postgres.MockAdRepository)
	adService := NewAdService(mockAdRepo)

	ad := &models.Ad{
		UserID:      1,
		Title:       "Test Ad",
		Description: "Test Description",
		Price:       100.0,
	}

	// Ожидаем вызов CreateAd и возвращаем ID 1
	mockAdRepo.On("CreateAd", mock.Anything, ad).Return(int64(1), nil)

	// 2. Действие
	id, err := adService.CreateAd(context.Background(), ad)

	// 3. Утверждение
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	mockAdRepo.AssertExpectations(t)
}

// Тестирование успешного обновления объявления владельцем
func TestAdService_UpdateAd_Success(t *testing.T) {
	// 1. Настройка
	mockAdRepo := new(postgres.MockAdRepository)
	adService := NewAdService(mockAdRepo)

	adID := int64(1)
	userID := int64(1) // Владелец

	// Объявление, которое "хранится" в базе
	existingAd := &models.Ad{
		ID:          adID,
		UserID:      userID,
		Title:       "Old Title",
		Description: "Old Description",
		Price:       100.0,
	}

	// Запрос на обновление
	newTitle := "New Title"
	updateReq := models.UpdateAdRequest{
		Title: &newTitle,
	}

	// Ожидаем, что сервис сначала запросит объявление по ID
	mockAdRepo.On("GetAdByID", mock.Anything, adID).Return(existingAd, nil)
	// Затем, ожидаем вызов UpdateAd с обновленными данными
	mockAdRepo.On("UpdateAd", mock.Anything, mock.MatchedBy(func(ad *models.Ad) bool {
		return ad.Title == newTitle && ad.ID == adID
	})).Return(nil)

	// 2. Действие
	updatedAd, err := adService.UpdateAd(context.Background(), adID, userID, updateReq)

	// 3. Утверждение
	assert.NoError(t, err)
	assert.NotNil(t, updatedAd)
	assert.Equal(t, newTitle, updatedAd.Title)
	mockAdRepo.AssertExpectations(t)
}

// Тестирование попытки обновления чужого объявления
func TestAdService_UpdateAd_AccessDenied(t *testing.T) {
	// 1. Настройка
	mockAdRepo := new(postgres.MockAdRepository)
	adService := NewAdService(mockAdRepo)

	adID := int64(1)
	ownerID := int64(1)    // Владелец
	notOwnerID := int64(2) // Посторонний пользователь

	existingAd := &models.Ad{ID: adID, UserID: ownerID}
	newTitle := "New Title"
	updateReq := models.UpdateAdRequest{Title: &newTitle}

	// Симулируем, что объявление найдено
	mockAdRepo.On("GetAdByID", mock.Anything, adID).Return(existingAd, nil)
	// Метод UpdateAd не должен быть вызван!

	// 2. Действие
	_, err := adService.UpdateAd(context.Background(), adID, notOwnerID, updateReq)

	// 3. Утверждение
	assert.Error(t, err)
	assert.Equal(t, postgres.ErrAdAccessDenied, err)
	mockAdRepo.AssertExpectations(t)
}

// Тестирование успешного удаления объявления владельцем
func TestAdService_DeleteAd_Success(t *testing.T) {
	// 1. Настройка
	mockAdRepo := new(postgres.MockAdRepository)
	adService := NewAdService(mockAdRepo)

	adID := int64(1)
	userID := int64(1)

	// Ожидаем вызов DeleteAd с правильными ID
	mockAdRepo.On("DeleteAd", mock.Anything, adID, userID).Return(nil)

	// 2. Действие
	err := adService.DeleteAd(context.Background(), adID, userID)

	// 3. Утверждение
	assert.NoError(t, err)
	mockAdRepo.AssertExpectations(t)
}
