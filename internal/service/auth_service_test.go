package service

import (
	"context"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"
	"marketplace/pkg/auth"
	"testing"
	"time"
	"marketplace/pkg/hash"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Тестирование успешной регистрации
func TestAuthService_Register_Success(t *testing.T) {
	// 1. Настройка (Arrange)
	mockUserRepo := new(postgres.MockUserRepository)
	tm, _ := auth.NewTokenManager("secret")
	authService := NewAuthService(mockUserRepo, tm, time.Hour)

	username := "testuser"
	password := "password123"

	// Ожидаем, что GetUserByUsername будет вызван с "testuser"
	// и вернет ошибку, что пользователя нет (это хорошо, мы можем его создать)
	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(nil, postgres.ErrUserNotFound)

	// Ожидаем, что CreateUser будет вызван и вернет ID=1 без ошибки.
	mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(int64(1), nil)

	// 2. Действие (Act)
	user, err := authService.Register(context.Background(), username, password)

	// 3. Утверждение (Assert)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, username, user.Username)
	mockUserRepo.AssertExpectations(t) // Проверяем, что все ожидаемые вызовы были сделаны
}

// Тестирование регистрации, когда пользователь уже существует
func TestAuthService_Register_UserExists(t *testing.T) {
	// 1. Настройка
	mockUserRepo := new(postgres.MockUserRepository)
	tm, _ := auth.NewTokenManager("secret")
	authService := NewAuthService(mockUserRepo, tm, time.Hour)

	username := "existinguser"
	password := "password123"

	// Симулируем, что пользователь уже существует в базе
	existingUser := &models.User{ID: 1, Username: username}
	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(existingUser, nil)

	// 2. Действие
	user, err := authService.Register(context.Background(), username, password)

	// 3. Утверждение
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserExists, err)
	mockUserRepo.AssertExpectations(t)
}

// Тестирование успешного входа
func TestAuthService_Login_Success(t *testing.T) {
	// 1. Настройка
	mockUserRepo := new(postgres.MockUserRepository)
	tm, _ := auth.NewTokenManager("secret")
	authService := NewAuthService(mockUserRepo, tm, time.Hour)

	username := "testuser"
	password := "password123"
	// Пароль, который хранится в базе (хэшированный)
	hashedPassword, _ := hash.HashPassword(password)

	userFromDB := &models.User{
		ID:       1,
		Username: username,
		Password: hashedPassword,
	}

	// Симулируем, что пользователь найден
	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(userFromDB, nil)

	// 2. Действие
	token, err := authService.Login(context.Background(), username, password)

	// 3. Утверждение
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockUserRepo.AssertExpectations(t)
}

// Тестирование входа с неверным паролем
func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	// 1. Настройка
	mockUserRepo := new(postgres.MockUserRepository)
	tm, _ := auth.NewTokenManager("secret")
	authService := NewAuthService(mockUserRepo, tm, time.Hour)

	username := "testuser"
	correctPassword := "password123"
	wrongPassword := "wrongpassword"

	hashedPassword, _ := hash.HashPassword(correctPassword)
	userFromDB := &models.User{
		ID:       1,
		Username: username,
		Password: hashedPassword,
	}

	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(userFromDB, nil)

	// 2. Действие
	token, err := authService.Login(context.Background(), username, wrongPassword)

	// 3. Утверждение
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}