package service

import (
	"context"
	"marketplace/internal/config"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"
	"marketplace/pkg/auth"
	"marketplace/pkg/hash"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Register_Success(t *testing.T) {
	cfg := config.Auth{
		JWTSecret: "secret",
		TokenTTL:  time.Hour,
	}
	mockUserRepo := new(postgres.MockUserRepository)
	tm, _ := auth.NewTokenManager(cfg)
	authService := NewAuthService(mockUserRepo, tm)

	username := "testuser"
	password := "password123"

	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(nil, postgres.ErrUserNotFound)

	mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(int64(1), nil)

	user, err := authService.Register(context.Background(), username, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, username, user.Username)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_UserExists(t *testing.T) {
	cfg := config.Auth{
		JWTSecret: "secret",
		TokenTTL:  time.Hour,
	}
	mockUserRepo := new(postgres.MockUserRepository)
	tm, _ := auth.NewTokenManager(cfg)
	authService := NewAuthService(mockUserRepo, tm)

	username := "existinguser"
	password := "password123"

	existingUser := &models.User{ID: 1, Username: username}
	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(existingUser, nil)

	user, err := authService.Register(context.Background(), username, password)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserExists, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	cfg := config.Auth{
		JWTSecret: "secret",
		TokenTTL:  time.Hour,
	}
	mockUserRepo := new(postgres.MockUserRepository)
	tm, _ := auth.NewTokenManager(cfg)
	authService := NewAuthService(mockUserRepo, tm)

	username := "testuser"
	password := "password123"
	hashedPassword, _ := hash.HashPassword(password)

	userFromDB := &models.User{
		ID:       1,
		Username: username,
		Password: hashedPassword,
	}

	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(userFromDB, nil)

	token, err := authService.Login(context.Background(), username, password)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	cfg := config.Auth{
		JWTSecret: "secret",
		TokenTTL:  time.Hour,
	}
	mockUserRepo := new(postgres.MockUserRepository)
	tm, _ := auth.NewTokenManager(cfg)
	authService := NewAuthService(mockUserRepo, tm)

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

	token, err := authService.Login(context.Background(), username, wrongPassword)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}
