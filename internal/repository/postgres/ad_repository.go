package postgres

import (
	"context"
	"errors"
	"fmt"
	"marketplace/internal/models"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAdNotFound     = errors.New("ad not found")
	ErrAdAccessDenied = errors.New("access denied")

	allowedSortBy = map[string]struct{}{
		"created_at": {},
		"price":      {},
	}
)

type adRepository struct {
	db *pgxpool.Pool
}

func NewAdRepository(db *pgxpool.Pool) AdRepository {
	return &adRepository{db: db}
}

func (r *adRepository) CreateAd(ctx context.Context, ad *models.Ad) (int64, error) {
	query := fmt.Sprintf(`INSERT INTO %s (user_id, title, description, price, image_url) 
	          						VALUES ($1, $2, $3, $4, $5) RETURNING id`, adsTable)
	var id int64
	err := r.db.QueryRow(ctx, query, ad.UserID, ad.Title, ad.Description, ad.Price, ad.ImageURL).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("repository.CreateAd: %w", err)
	}
	return id, nil
}

type GetAllAdsParams struct {
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

func (r adRepository) GetAllAds(ctx context.Context, params GetAllAdsParams) ([]models.Ad, error) {
	baseQuery := fmt.Sprintf(`SELECT id, user_id, title, description, price, image_url, created_at 
														FROM %s`, adsTable)

	var queryBuilder strings.Builder
	queryBuilder.WriteString(baseQuery)

	if _, ok := allowedSortBy[params.SortBy]; ok {
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s", params.SortBy))

		if strings.ToUpper(params.SortOrder) == "ASC" {
			queryBuilder.WriteString(" ASC")
		} else {
			queryBuilder.WriteString(" DESC")
		}
	} else {
		queryBuilder.WriteString(" ORDER BY created_at DESC")
	}

	queryBuilder.WriteString(" LIMIT $1 OFFSET $2")

	finalQuery := queryBuilder.String()

	rows, err := r.db.Query(ctx, finalQuery, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("repository.GetAllAds: query error: %w", err)
	}
	defer rows.Close()

	var ads []models.Ad
	for rows.Next() {
		var ad models.Ad
		if err := rows.Scan(&ad.ID, &ad.UserID, &ad.Title, &ad.Description, &ad.Price, &ad.ImageURL, &ad.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository.GetAllAds: row scan error: %w", err)
		}
		ads = append(ads, ad)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.GetAllAds: %w", err)
	}

	return ads, nil
}

func (r *adRepository) GetAdByID(ctx context.Context, id int64) (*models.Ad, error) {
	query := fmt.Sprintf(`SELECT id, user_id, title, description, price, image_url, created_at, updated_at 
												FROM %s WHERE id = $1`, adsTable)
	var ad models.Ad
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ad.ID, &ad.UserID, &ad.Title, &ad.Description, &ad.Price, &ad.ImageURL, &ad.CreatedAt, &ad.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAdNotFound
		}
		return nil, fmt.Errorf("repository.GetAdByID: %w", err)
	}
	return &ad, nil
}

func (r *adRepository) UpdateAd(ctx context.Context, ad *models.Ad) error {
	query := fmt.Sprintf(`UPDATE %s SET title = $1, description = $2, price = $3, updated_at = NOW()
												WHERE id = $4 AND user_id = $5`, adsTable)

	res, err := r.db.Exec(ctx, query, ad.Title, ad.Description, ad.Price, ad.ID, ad.UserID)
	if err != nil {
		return fmt.Errorf("repository.UpdateAd: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrAdAccessDenied
	}
	return nil
}

func (r *adRepository) DeleteAd(ctx context.Context, id, userID int64) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1 AND user_id = $2`, adsTable)
	res, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("repository.DeleteAd: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrAdAccessDenied
	}
	return nil
}
