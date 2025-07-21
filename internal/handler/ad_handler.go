package handler

import (
	"errors"
	"fmt"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary Создание нового объявления
// @Security ApiKeyAuth
// @Tags ads
// @Description Создает новое объявление от имени авторизованного пользователя
// @Accept  json
// @Produce  json
// @Param   input body models.CreateAdRequest true "Данные для создания объявления"
// @Success 201 {object} models.CreateAdResponse "ID созданного объявления" // <--- ИЗМЕНЕНО
// @Failure 400 {object} ErrorResponse "Неверный формат запроса"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /ads [post]
func (h *Handler) CreateAd(c *gin.Context) {
	var req models.CreateAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	userID, ok := GetUserIDFromCtx(c)
	if !ok {
		h.newErrorResponse(c, http.StatusUnauthorized, "user context not found", fmt.Errorf("user context not found"))
		return
	}

	ad := &models.Ad{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
	}

	adID, err := h.service.Ad.CreateAd(c.Request.Context(), ad)
	if err != nil {
		h.newErrorResponse(c, http.StatusInternalServerError, "failed to create ad", err)
		return
	}

	c.JSON(http.StatusCreated, models.CreateAdResponse{ID: adID})
}

// @Summary Получение списка объявлений
// @Tags ads
// @Description Возвращает список объявлений с возможностью пагинации и сортировки
// @Produce  json
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Param sort_by query string false "Поле для сортировки" Enums(created_at, price) default(created_at)
// @Param sort_order query string false "Порядок сортировки" Enums(asc, desc) default(desc)
// @Success 200 {array} models.AdResponse "Список объявлений"
// @Failure 400 {object} ErrorResponse "Неверные параметры запроса"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /ads [get]
func (h *Handler) GetAllAds(c *gin.Context) {
	var query models.AdsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "invalid query parameters", err)
		return
	}

	offset := (query.Page - 1) * query.Limit

	params := postgres.GetAllAdsParams{
		Limit:     query.Limit,
		Offset:    offset,
		SortBy:    query.SortBy,
		SortOrder: query.SortOrder,
	}

	ads, err := h.service.Ad.GetAllAds(c.Request.Context(), params)
	if err != nil {
		h.newErrorResponse(c, http.StatusInternalServerError, "failed to get ads", err)
		return
	}

	var responses []models.AdResponse
	for _, ad := range ads {
		responses = append(responses, models.AdResponse{
			ID:          ad.ID,
			Title:       ad.Title,
			Description: ad.Description,
			Price:       ad.Price,
			ImageURL:    ad.ImageURL,
			AuthorID:    ad.UserID,
			CreatedAt:   ad.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, responses)
}

// @Summary Получение объявления по ID
// @Tags ads
// @Description Возвращает одно объявление по его уникальному идентификатору
// @Produce  json
// @Param id path int true "ID объявления"
// @Success 200 {object} models.Ad "Полные данные объявления"
// @Failure 400 {object} ErrorResponse "Неверный ID объявления"
// @Failure 404 {object} ErrorResponse "Объявление не найдено"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /ads/{id} [get]
func (h *Handler) GetAdByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "invalid ad ID", err)
		return
	}

	ad, err := h.service.Ad.GetAdByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, postgres.ErrAdNotFound) {
			h.newErrorResponse(c, http.StatusNotFound, "ad not found", err)
			return
		}
		h.newErrorResponse(c, http.StatusInternalServerError, "internal server error", err)
		return
	}

	c.JSON(http.StatusOK, toAdResponse(ad))
}

// @Summary Обновление объявления
// @Security ApiKeyAuth
// @Tags ads
// @Description Обновляет данные объявления (только владелец)
// @Accept  json
// @Produce  json
// @Param id path int true "ID объявления"
// @Param input body models.UpdateAdRequest true "Поля для обновления"
// @Success 200 {object} models.Ad "Обновленные данные объявления"
// @Failure 400 {object} ErrorResponse "Неверный формат запроса или ID"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} ErrorResponse "Доступ запрещен (не владелец)"
// @Failure 404 {object} ErrorResponse "Объявление не найдено"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /ads/{id} [patch]
func (h *Handler) UpdateAd(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "invalid ad ID", err)
		return
	}

	var req models.UpdateAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	userID, ok := GetUserIDFromCtx(c)
	if !ok {
		h.newErrorResponse(c, http.StatusUnauthorized, "invalid user context", fmt.Errorf("user context not found"))
		return
	}

	updatedAd, err := h.service.Ad.UpdateAd(c.Request.Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, postgres.ErrAdNotFound) {
			h.newErrorResponse(c, http.StatusNotFound, "ad not found", err)
		} else if errors.Is(err, postgres.ErrAdAccessDenied) {
			h.newErrorResponse(c, http.StatusForbidden, "access denied", err)
		} else {
			h.newErrorResponse(c, http.StatusInternalServerError, "internal server error", err)
		}
		return
	}

	c.JSON(http.StatusOK, toAdResponse(updatedAd))
}

// @Summary Удаление объявления
// @Security ApiKeyAuth
// @Tags ads
// @Description Удаляет объявление (только владелец)
// @Param id path int true "ID объявления для удаления"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse "Неверный ID объявления"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} ErrorResponse "Доступ запрещен (не владелец)"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /ads/{id} [delete]
func (h *Handler) DeleteAd(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "invalid ad ID", err)
		return
	}

	userID, ok := GetUserIDFromCtx(c)
	if !ok {
		h.newErrorResponse(c, http.StatusUnauthorized, "invalid user context", fmt.Errorf("user context not found"))
		return
	}

	err = h.service.Ad.DeleteAd(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrAdAccessDenied) {
			h.newErrorResponse(c, http.StatusForbidden, "access denied", err)
		} else {
			h.newErrorResponse(c, http.StatusInternalServerError, "internal server error", err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func toAdResponse(ad *models.Ad) models.AdResponse {
	return models.AdResponse{
		ID:          ad.ID,
		Title:       ad.Title,
		Description: ad.Description,
		Price:       ad.Price,
		ImageURL:    ad.ImageURL,
		AuthorID:    ad.UserID,
		CreatedAt:   ad.CreatedAt,
	}
}
