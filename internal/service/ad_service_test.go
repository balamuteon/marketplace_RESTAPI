package service

import (
	"context"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdService_CreateAd_Success(t *testing.T) {
	mockAdRepo := new(postgres.MockAdRepository)
	adService := NewAdService(mockAdRepo)

	ad := &models.Ad{
		UserID:      1,
		Title:       "Test Ad",
		Description: "Test Description",
		Price:       100.0,
	}

	mockAdRepo.On("CreateAd", mock.Anything, ad).Return(int64(1), nil)

	id, err := adService.CreateAd(context.Background(), ad)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	mockAdRepo.AssertExpectations(t)
}

func TestAdService_UpdateAd_Success(t *testing.T) {
	mockAdRepo := new(postgres.MockAdRepository)
	adService := NewAdService(mockAdRepo)

	adID := int64(1)
	userID := int64(1)

	existingAd := &models.Ad{
		ID:          adID,
		UserID:      userID,
		Title:       "Old Title",
		Description: "Old Description",
		Price:       100.0,
	}

	newTitle := "New Title"
	updateReq := models.UpdateAdRequest{
		Title: &newTitle,
	}

	mockAdRepo.On("GetAdByID", mock.Anything, adID).Return(existingAd, nil)
	mockAdRepo.On("UpdateAd", mock.Anything, mock.MatchedBy(func(ad *models.Ad) bool {
		return ad.Title == newTitle && ad.ID == adID
	})).Return(nil)

	updatedAd, err := adService.UpdateAd(context.Background(), adID, userID, updateReq)

	assert.NoError(t, err)
	assert.NotNil(t, updatedAd)
	assert.Equal(t, newTitle, updatedAd.Title)
	mockAdRepo.AssertExpectations(t)
}

func TestAdService_UpdateAd_AccessDenied(t *testing.T) {
	mockAdRepo := new(postgres.MockAdRepository)
	adService := NewAdService(mockAdRepo)

	adID := int64(1)
	ownerID := int64(1)
	notOwnerID := int64(2)

	existingAd := &models.Ad{ID: adID, UserID: ownerID}
	newTitle := "New Title"
	updateReq := models.UpdateAdRequest{Title: &newTitle}

	mockAdRepo.On("GetAdByID", mock.Anything, adID).Return(existingAd, nil)

	_, err := adService.UpdateAd(context.Background(), adID, notOwnerID, updateReq)

	assert.Error(t, err)
	assert.Equal(t, postgres.ErrAdAccessDenied, err)
	mockAdRepo.AssertExpectations(t)
}

func TestAdService_DeleteAd_Success(t *testing.T) {
	mockAdRepo := new(postgres.MockAdRepository)
	adService := NewAdService(mockAdRepo)

	adID := int64(1)
	userID := int64(1)

	mockAdRepo.On("DeleteAd", mock.Anything, adID, userID).Return(nil)

	err := adService.DeleteAd(context.Background(), adID, userID)

	assert.NoError(t, err)
	mockAdRepo.AssertExpectations(t)
}
