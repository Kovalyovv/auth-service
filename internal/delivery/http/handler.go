package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/Kovalyovv/auth-service/internal/domain"
	"github.com/gin-gonic/gin"
	"log/slog"
)

type AuthUseCase interface {
	Register(ctx context.Context, username, email, password string) error
	Login(ctx context.Context, email, password string) (domain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error)
}

type AuthHandler struct {
	uc AuthUseCase
}

func NewAuthHandler(uc AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

type registerReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
type apiError struct {
	Error string `json:"error"`
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	slog.Error("http handler error", "path", c.Request.URL.Path, "error", err)

	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.AbortWithStatusJSON(http.StatusUnauthorized, apiError{Error: err.Error()})
	case errors.Is(err, domain.ErrRefreshTokenNotFound):
		c.AbortWithStatusJSON(http.StatusUnauthorized, apiError{Error: err.Error()})
	case errors.Is(err, domain.ErrEmailExists):
		c.AbortWithStatusJSON(http.StatusConflict, apiError{Error: err.Error()})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, apiError{Error: "an internal server error occurred"})
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "invalid request body"})
		return
	}

	if err := h.uc.Register(c.Request.Context(), req.Username, req.Email, req.Password); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "invalid request body"})
		return
	}

	pair, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pair)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "invalid request body"})
		return
	}

	pair, err := h.uc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pair)
}
