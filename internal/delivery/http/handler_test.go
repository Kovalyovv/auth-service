package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kovalyovv/auth-service/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Register(ctx context.Context, username, email, password string) error {
	args := m.Called(ctx, username, email, password)
	return args.Error(0)
}

func (m *MockAuthUseCase) Login(ctx context.Context, email, password string) (domain.TokenPair, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(domain.TokenPair), args.Error(1)
}

func (m *MockAuthUseCase) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(domain.TokenPair), args.Error(1)
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Given valid credentials", func(t *testing.T) {
		mockUC := new(MockAuthUseCase)
		handler := NewAuthHandler(mockUC)

		expectedPair := domain.TokenPair{AccessToken: "access", RefreshToken: "refresh"}
		loginReq := loginReq{Email: "test@example.com", Password: "password"}
		mockUC.On("Login", mock.Anything, loginReq.Email, loginReq.Password).Return(expectedPair, nil).Once()

		router := gin.New()
		router.POST("/login", handler.Login)

		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var respPair domain.TokenPair
		err := json.Unmarshal(rr.Body.Bytes(), &respPair)
		assert.NoError(t, err)
		assert.Equal(t, expectedPair, respPair)

		mockUC.AssertExpectations(t)
	})

	t.Run("Given invalid json", func(t *testing.T) {
		mockUC := new(MockAuthUseCase)
		handler := NewAuthHandler(mockUC)

		router := gin.New()
		router.POST("/login", handler.Login)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email": "test@test.com"`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
