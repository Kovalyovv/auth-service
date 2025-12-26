package domain

import "errors"

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserNotFound         = errors.New("user not found")
	ErrRefreshTokenNotFound = errors.New("invalid or expired refresh token")
	ErrTokenExpired         = errors.New("token has expired")
	ErrEmailExists          = errors.New("email already exists")
)
