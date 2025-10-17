package service

import (
	"context"
	"fmt"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"
)

type adService struct {
	adRepo postgres.AdRepository
}

func NewAdService(adRepo postgres.AdRepository) *adService {
	return &adService{
		adRepo: adRepo,
	}
}

func (s *adService) CreateAd(ctx context.Context, ad *models.Ad) (int64, error) {
	id, err := s.adRepo.CreateAd(ctx, ad)
	if err != nil {
		return 0, fmt.Errorf("service.CreateAd: %w", err)
	}
	return id, nil
}

func (s *adService) GetAllAds(ctx context.Context, params postgres.GetAllAdsParams) ([]models.Ad, error) {
	ads, err := s.adRepo.GetAllAds(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("service.GetAllAds: %w", err)
	}

	return ads, nil
}

func (s *adService) GetAdByID(ctx context.Context, id int64) (*models.Ad, error) {
	return s.adRepo.GetAdByID(ctx, id)
}

func (s *adService) UpdateAd(ctx context.Context, id, userID int64, req models.UpdateAdRequest) (*models.Ad, error) {
	ad, err := s.adRepo.GetAdByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if ad.UserID != userID {
		return nil, postgres.ErrAdAccessDenied
	}

	if req.Title != nil {
		ad.Title = *req.Title
	}
	if req.Description != nil {
		ad.Description = *req.Description
	}
	if req.Price != nil {
		ad.Price = *req.Price
	}

	if err := s.adRepo.UpdateAd(ctx, ad); err != nil {
		return nil, err
	}
	return ad, nil
}

func (s *adService) DeleteAd(ctx context.Context, id, userID int64) error {
	return s.adRepo.DeleteAd(ctx, id, userID)
}
