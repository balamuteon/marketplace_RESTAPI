package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"
	"marketplace/internal/service"
	"marketplace/pkg/auth"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Тестируем обработчик регистрации пользователя
func TestHandler_signUp(t *testing.T) {
	// --- Подготовка ---
	// Создаем "пустой" логгер, который не будет выводить логи во время тестов
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	tm, _ := auth.NewTokenManager("secret")

	// --- Тестовые случаи ---
	testCases := []struct {
		name                string
		requestBody         string
		mockServiceResponse *models.User
		mockServiceError    error
		expectedStatusCode  int
		expectedBody        string
	}{
		{
			name:                "Успешная регистрация",
			requestBody:         `{"username": "testuser", "password": "password123"}`,
			mockServiceResponse: &models.User{ID: 1, Username: "testuser"},
			mockServiceError:    nil,
			expectedStatusCode:  http.StatusCreated,
			expectedBody:        `{"id":1,"username":"testuser","created_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:                "Пользователь уже существует",
			requestBody:         `{"username": "existinguser", "password": "password123"}`,
			mockServiceResponse: nil,
			mockServiceError:    service.ErrUserExists, // Симулируем ошибку от сервиса
			expectedStatusCode:  http.StatusConflict,
			expectedBody:        `{"message":"user already exists"}`,
		},
		{
			name:                "Некорректное тело запроса (нет пароля)",
			requestBody:         `{"username": "nouser"}`,
			mockServiceResponse: nil,
			mockServiceError:    nil, // Сервис не будет вызван, ошибка на уровне Gin
			expectedStatusCode:  http.StatusBadRequest,
			expectedBody:        `{"message":"invalid request body"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- Настройка мока для каждого случая ---
			mockAuthService := new(service.MockAuthService)
			// Программируем мок, только если ожидается вызов сервиса
			if tc.mockServiceError != nil || tc.mockServiceResponse != nil {
				var req models.RegisterRequest
				json.Unmarshal([]byte(tc.requestBody), &req)
				mockAuthService.On("Register", mock.Anything, req.Username, req.Password).
					Return(tc.mockServiceResponse, tc.mockServiceError)
			}

			// --- Инициализация хендлера и роутера ---
			services := &service.Service{Auth: mockAuthService}
			handler := NewHandler(services, tm, logger)
			router := handler.InitRoutes()

			// --- Создание фейкового HTTP запроса ---
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")

			// --- Запись ответа ---
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req) // Выполняем запрос через роутер

			// --- Проверка результатов ---
			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			assert.JSONEq(t, tc.expectedBody, rec.Body.String())
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestHandler_signIn(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	tm, _ := auth.NewTokenManager("secret")

	testCases := []struct {
		name                string
		requestBody         string
		mockServiceResponse string
		mockServiceError    error
		expectedStatusCode  int
		expectedBodyPart    string // Проверяем только часть тела, т.к. токен всегда разный
	}{
		{
			name:                "Успешный вход",
			requestBody:         `{"username": "testuser", "password": "password123"}`,
			mockServiceResponse: "some.jwt.token",
			mockServiceError:    nil,
			expectedStatusCode:  http.StatusOK,
			expectedBodyPart:    `"token":"some.jwt.token"`,
		},
		{
			name:                "Неверные учетные данные",
			requestBody:         `{"username": "testuser", "password": "wrongpassword"}`,
			mockServiceResponse: "",
			mockServiceError:    service.ErrInvalidCredentials,
			expectedStatusCode:  http.StatusUnauthorized,
			expectedBodyPart:    `"message":"invalid credentials"`,
		},
		{
			name:                "Некорректное тело запроса",
			requestBody:         `{"username": "testuser"}`,
			mockServiceResponse: "",
			mockServiceError:    nil, // Сервис не будет вызван
			expectedStatusCode:  http.StatusBadRequest,
			expectedBodyPart:    `"message":"invalid request body"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAuthService := new(service.MockAuthService)
			if tc.mockServiceError != nil || tc.mockServiceResponse != "" {
				var req models.LoginRequest
				json.Unmarshal([]byte(tc.requestBody), &req)
				mockAuthService.On("Login", mock.Anything, req.Username, req.Password).
					Return(tc.mockServiceResponse, tc.mockServiceError)
			}

			services := &service.Service{Auth: mockAuthService}
			handler := NewHandler(services, tm, logger)
			router := handler.InitRoutes()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedBodyPart)
			mockAuthService.AssertExpectations(t)
		})
	}
}

// Тестируем обработчик создания объявления
func TestHandler_CreateAd(t *testing.T) {
	// --- Подготовка ---
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	tm, _ := auth.NewTokenManager("secret")

	// --- Настройка мока ---
	mockAdService := new(service.MockAdService)
	adID := int64(123)
	// Ожидаем, что сервис будет вызван с данными объявления и вернет ID
	mockAdService.On("CreateAd", mock.Anything, mock.AnythingOfType("*models.Ad")).Return(adID, nil)

	// --- Инициализация ---
	services := &service.Service{Ad: mockAdService}
	handler := NewHandler(services, tm, logger)
	router := handler.InitRoutes()

	// --- Создание запроса ---
	requestBody := `{"title": "Test Ad", "description": "A great ad", "price": 99.99}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ads", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// --- Симуляция авторизации через Middleware ---
	// В реальном приложении токен генерируется при логине
	// В тесте мы его просто создаем для авторизованного пользователя с ID=1
	testUserID := int64(1)
	token, _ := tm.GenerateToken(testUserID, "testuser", time.Hour)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// --- Запись ответа ---
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// --- Проверка ---
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.JSONEq(t, fmt.Sprintf(`{"id":%d}`, adID), rec.Body.String())
	mockAdService.AssertExpectations(t)
}

// НОВЫЙ ТЕСТ: Тестируем обновление объявления с проверкой прав
func TestHandler_UpdateAd(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	tm, _ := auth.NewTokenManager("secret")

	adID := int64(1)
	ownerID := int64(10)
	notOwnerID := int64(20)

	testCases := []struct {
		name               string
		actorID            int64 // ID пользователя, который выполняет действие
		mockServiceError   error
		expectedStatusCode int
		expectedBodyPart   string
	}{
		{
			name:               "Успешное обновление владельцем",
			actorID:            ownerID,
			mockServiceError:   nil,
			expectedStatusCode: http.StatusOK,
			expectedBodyPart:   `"id":1`,
		},
		{
			name:               "Попытка обновления НЕ владельцем",
			actorID:            notOwnerID,
			mockServiceError:   postgres.ErrAdAccessDenied, // Симулируем ошибку доступа от сервиса
			expectedStatusCode: http.StatusForbidden,
			expectedBodyPart:   `"message":"access denied"`,
		},
		{
			name:               "Попытка обновления несуществующего объявления",
			actorID:            ownerID,
			mockServiceError:   postgres.ErrAdNotFound, // Симулируем ошибку "не найдено"
			expectedStatusCode: http.StatusNotFound,
			expectedBodyPart:   `"message":"ad not found"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAdService := new(service.MockAdService)
			// Программируем мок, ожидая вызов UpdateAd
			// Здесь мы не будем проверять тело запроса для простоты,
			// но в реальном проекте это стоило бы сделать.
			mockAdService.On("UpdateAd", mock.Anything, adID, tc.actorID, mock.AnythingOfType("models.UpdateAdRequest")).
				Return(&models.Ad{ID: adID}, tc.mockServiceError)

			services := &service.Service{Ad: mockAdService}
			handler := NewHandler(services, tm, logger)
			router := handler.InitRoutes()

			requestBody := `{"title": "New Title"}`
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/ads/%d", adID), bytes.NewBufferString(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Генерируем токен для "актера"
			token, _ := tm.GenerateToken(tc.actorID, "actor", time.Hour)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedBodyPart)
		})
	}
}

// НОВЫЙ ТЕСТ: Тестируем удаление объявления с проверкой прав
func TestHandler_DeleteAd(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	tm, _ := auth.NewTokenManager("secret")

	adID := int64(1)
	ownerID := int64(10)
	notOwnerID := int64(20)

	// Сценарий 1: Успешное удаление
	t.Run("Успешное удаление владельцем", func(t *testing.T) {
		mockAdService := new(service.MockAdService)
		mockAdService.On("DeleteAd", mock.Anything, adID, ownerID).Return(nil)

		services := &service.Service{Ad: mockAdService}
		handler := NewHandler(services, tm, logger)
		router := handler.InitRoutes()

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/ads/%d", adID), nil)
		token, _ := tm.GenerateToken(ownerID, "owner", time.Hour)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockAdService.AssertExpectations(t)
	})

	// Сценарий 2: Попытка удаления НЕ владельцем
	t.Run("Попытка удаления НЕ владельцем", func(t *testing.T) {
		mockAdService := new(service.MockAdService)
		mockAdService.On("DeleteAd", mock.Anything, adID, notOwnerID).Return(postgres.ErrAdAccessDenied)

		services := &service.Service{Ad: mockAdService}
		handler := NewHandler(services, tm, logger)
		router := handler.InitRoutes()

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/ads/%d", adID), nil)
		token, _ := tm.GenerateToken(notOwnerID, "not-owner", time.Hour)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		mockAdService.AssertExpectations(t)
	})
}
