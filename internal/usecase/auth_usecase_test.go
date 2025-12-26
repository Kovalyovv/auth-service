package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/Kovalyovv/auth-service/internal/domain"
	"github.com/Kovalyovv/auth-service/internal/pkg/hash"
	"github.com/Kovalyovv/auth-service/internal/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Error(0)
}

func (m *MockUserRepository) ConsumeRefreshToken(ctx context.Context, token string) (int64, error) {
	args := m.Called(ctx, token)
	return int64(args.Int(0)), args.Error(1)
}

func TestAuthUseCase_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := jwt.NewTokenManager("secret")
	uc := NewAuthUseCase(mockRepo, tokenManager)

	password := "password123"
	hashedPassword, _ := hash.HashPassword(password)

	t.Run("Given valid credentials", func(t *testing.T) {
		ctx := context.Background()
		user := &domain.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
		}

		mockRepo.On("GetByEmail", ctx, user.Email).Return(user, nil).Once()
		mockRepo.On("SaveRefreshToken", ctx, user.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil).Once()

		pair, err := uc.Login(ctx, user.Email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Given non-existent user", func(t *testing.T) {
		ctx := context.Background()
		email := "notfound@example.com"
		mockRepo.On("GetByEmail", ctx, email).Return(nil, domain.ErrUserNotFound).Once()

		_, err := uc.Login(ctx, email, password)

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Given incorrect password", func(t *testing.T) {
		ctx := context.Background()
		user := &domain.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
		}
		mockRepo.On("GetByEmail", ctx, user.Email).Return(user, nil).Once()

		_, err := uc.Login(ctx, user.Email, "wrongpassword")

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_Refresh(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := jwt.NewTokenManager("secret")
	uc := NewAuthUseCase(mockRepo, tokenManager)

	t.Run("Given valid refresh token", func(t *testing.T) {
		ctx := context.Background()
		refreshToken := "valid-token"
		userID := int64(1)

		mockRepo.On("ConsumeRefreshToken", ctx, refreshToken).Return(int(userID), nil).Once()
		mockRepo.On("SaveRefreshToken", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil).Once()

		pair, err := uc.Refresh(ctx, refreshToken)

		assert.NoError(t, err)
		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Given invalid refresh token", func(t *testing.T) {
		ctx := context.Background()
		refreshToken := "invalid-token"

		mockRepo.On("ConsumeRefreshToken", ctx, refreshToken).Return(0, domain.ErrRefreshTokenNotFound).Once()

		_, err := uc.Refresh(ctx, refreshToken)

		assert.ErrorIs(t, err, domain.ErrRefreshTokenNotFound)
		mockRepo.AssertExpectations(t)
	})
}
