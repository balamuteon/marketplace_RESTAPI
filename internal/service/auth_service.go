package service

import (
	"context"
	"errors"
	"fmt"
	"marketplace/internal/models"
	"marketplace/internal/repository/postgres"
	"marketplace/pkg/auth"
	"marketplace/pkg/hash"
	"time"
)

var (
	ErrUserExists         = errors.New("user with this username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

type authService struct {
	userRepo     postgres.UserRepository
	tokenManager *auth.TokenManager
	tokenTTL     time.Duration
}

func NewAuthService(userRepo postgres.UserRepository, tm *auth.TokenManager, ttl time.Duration) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenManager: tm,
		tokenTTL:     ttl,
	}
}

func (s *authService) Register(ctx context.Context, username, password string) (*models.User, error) {
	_, err := s.userRepo.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, ErrUserExists
	}
	if !errors.Is(err, postgres.ErrUserNotFound) {
		return nil, fmt.Errorf("service.Register: %w", err)
	}

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("service.Register: %w", err)
	}

	user := &models.User{
		Username: username,
		Password: hashedPassword,
	}

	id, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("service.Register: %w", err)
	}
	user.ID = id
	user.Password = "" // Очищаем пароль перед возвратом

	return user, nil
}

func (s *authService) Login(ctx context.Context, username, password string) (string, error) {
	const op = "service.Login"

	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if !hash.CheckPasswordHash(password, user.Password) {
		return "", ErrInvalidCredentials
	}

	token, err := s.tokenManager.GenerateToken(user.ID, user.Username, s.tokenTTL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}
