package handler

import (
	"errors"
	"fmt"
	"marketplace/internal/models"
	"marketplace/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Регистрация нового пользователя
// @Tags auth
// @Description Создает нового пользователя в системе
// @Accept  json
// @Produce  json
// @Param   input body models.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} models.UserResponse "Пользователь успешно создан"
// @Failure 400 {object} ErrorResponse "Неверный формат запроса"
// @Failure 409 {object} ErrorResponse "Пользователь уже существует"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
func (h *Handler) signUp(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	user, err := h.service.Auth.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			h.newErrorResponse(c, http.StatusConflict, "user already exists", err)
			return
		}
		h.newErrorResponse(c, http.StatusInternalServerError, "internal server error", fmt.Errorf("failed to register user: %w", err))
		return
	}

	c.JSON(http.StatusCreated, models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	})
}

// @Summary Авторизация пользователя
// @Tags auth
// @Description Авторизует пользователя и возвращает JWT токен
// @Accept  json
// @Produce  json
// @Param   input body models.LoginRequest true "Данные для входа"
// @Success 200 {object} models.LoginResponse "Успешная авторизация"
// @Failure 400 {object} ErrorResponse "Неверный формат запроса"
// @Failure 401 {object} ErrorResponse "Неверные учетные данные"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *Handler) signIn(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	token, err := h.service.Auth.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.newErrorResponse(c, http.StatusUnauthorized, "invalid credentials", err)
			return
		}
		h.newErrorResponse(c, http.StatusInternalServerError, "internal server error", err)
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{Token: token})
}
