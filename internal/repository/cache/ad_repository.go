package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"
	"marketplace/pkg/cache"
)

// AdRepository является декоратором над postgres.AdRepository для добавления кеширования.
type AdRepository struct {
	postgresRepo postgres.AdRepository // Основной репозиторий, который ходит в БД
	cache        *cache.CacheClient
}

// NewAdRepository создает новый экземпляр кеширующего репозитория.
func NewAdRepository(postgresRepo postgres.AdRepository, cache *cache.CacheClient) *AdRepository {
	return &AdRepository{
		postgresRepo: postgresRepo,
		cache:        cache,
	}
}

// adListCacheKey генерирует уникальный ключ для кеша списка объявлений.
func adListCacheKey(params postgres.GetAllAdsParams) string {
	limit := params.Limit
	page := 1
	if limit > 0 {
		page = params.Offset/limit + 1
	}
	return fmt.Sprintf("ads:page=%d&limit=%d&sort_by=%s&sort_order=%s",
		page,
		limit,
		params.SortBy,
		params.SortOrder,
	)
}

// GetAllAds сначала проверяет кеш, и только в случае промаха обращается к репозиторию БД.
func (r *AdRepository) GetAllAds(ctx context.Context, params postgres.GetAllAdsParams) ([]models.Ad, error) {
	key := adListCacheKey(params)

	if r.cache != nil && r.cache.Client != nil {
		cachedJSON, err := r.cache.Client.Get(ctx, key).Result()
		if err == nil {
			var ads []models.Ad
			if json.Unmarshal([]byte(cachedJSON), &ads) == nil {
				return ads, nil
			}
		}
	}

	ads, err := r.postgresRepo.GetAllAds(ctx, params)
	if err != nil {
		return nil, err
	}

	if r.cache != nil && r.cache.Client != nil {
		adsJSON, err := json.Marshal(ads)
		if err == nil {
			r.cache.Client.Set(ctx, key, adsJSON, r.cache.TTL()).Err()
		}
	}

	return ads, nil
}

// --- Методы, которые изменяют данные и инвалидируют кеш ---

// CreateAd создает объявление в БД. В текущей стратегии с TTL мы не инвалидируем кеш принудительно.
func (r *AdRepository) CreateAd(ctx context.Context, ad *models.Ad) (int64, error) {
	return r.postgresRepo.CreateAd(ctx, ad)
}

// UpdateAd обновляет объявление в БД.
func (r *AdRepository) UpdateAd(ctx context.Context, ad *models.Ad) error {
	// В будущем здесь можно добавить логику инвалидации кеша для конкретного объявления (ad:ID).
	return r.postgresRepo.UpdateAd(ctx, ad)
}

// DeleteAd удаляет объявление из БД.
func (r *AdRepository) DeleteAd(ctx context.Context, id, userID int64) error {
	return r.postgresRepo.DeleteAd(ctx, id, userID)
}

// GetAdByID просто проксирует вызов к основному репозиторию.
func (r *AdRepository) GetAdByID(ctx context.Context, id int64) (*models.Ad, error) {
	return r.postgresRepo.GetAdByID(ctx, id)
}