package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Kovalyovv/auth-service/internal/domain"
	"github.com/Kovalyovv/auth-service/internal/pkg/hash"
	"github.com/Kovalyovv/auth-service/internal/pkg/jwt"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type AuthUseCase struct {
	repo         UserRepository
	tokenManager *jwt.TokenManager
}

func NewAuthUseCase(repo UserRepository, tm *jwt.TokenManager) *AuthUseCase {
	return &AuthUseCase{repo: repo, tokenManager: tm}
}

func (uc *AuthUseCase) Register(ctx context.Context, username, email, password string) error {
	h, err := hash.HashPassword(password)
	if err != nil {
		return err
	}

	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: h,
	}
	return uc.repo.Create(ctx, user)
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !hash.CheckPasswordHash(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	return uc.tokenManager.GenerateToken(user.ID, time.Hour*24)
}

func (uc *AuthUseCase) Verify(token string) (int64, error) {
	return uc.tokenManager.ValidateToken(token)
}
