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
	SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (int64, time.Time, error)
	DeleteRefreshToken(ctx context.Context, token string) error
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

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (domain.TokenPair, error) {
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return domain.TokenPair{}, errors.New("invalid credentials")
	}

	if !hash.CheckPasswordHash(password, user.PasswordHash) {
		return domain.TokenPair{}, errors.New("invalid credentials")
	}

	return uc.generatePair(ctx, user.ID)
}

func (uc *AuthUseCase) Verify(token string) (int64, error) {
	return uc.tokenManager.ValidateToken(token)
}

func (uc *AuthUseCase) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	userID, expiresAt, err := uc.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return domain.TokenPair{}, errors.New("invalid refresh token")
	}

	if time.Now().After(expiresAt) {
		_ = uc.repo.DeleteRefreshToken(ctx, refreshToken)
		return domain.TokenPair{}, errors.New("refresh token expired")
	}

	_ = uc.repo.DeleteRefreshToken(ctx, refreshToken)

	return uc.generatePair(ctx, userID)
}

func (uc *AuthUseCase) generatePair(ctx context.Context, userID int64) (domain.TokenPair, error) {
	accessToken, err := uc.tokenManager.GenerateAccessToken(userID, time.Minute*15)
	if err != nil {
		return domain.TokenPair{}, err
	}

	refreshToken, err := uc.tokenManager.GenerateRefreshToken()
	if err != nil {
		return domain.TokenPair{}, err
	}

	expiresAt := time.Now().Add(time.Hour * 24 * 7)
	err = uc.repo.SaveRefreshToken(ctx, userID, refreshToken, expiresAt)
	if err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
