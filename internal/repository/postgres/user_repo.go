package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Kovalyovv/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := r.pool.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return domain.ErrEmailExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = $1`
	err := r.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("GetByEmail query failed: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

func (r *UserRepo) ConsumeRefreshToken(ctx context.Context, token string) (int64, error) {
	var userID int64

	query := `
		DELETE FROM refresh_tokens
		WHERE token = $1 AND expires_at > now()
		RETURNING user_id
	`
	err := r.pool.QueryRow(ctx, query, token).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrRefreshTokenNotFound
		}
		return 0, fmt.Errorf("consume refresh token failed: %w", err)
	}
	return userID, nil
}

func (r *UserRepo) GetRefreshToken(ctx context.Context, token string) (int64, time.Time, error) {
	var userID int64
	var expiresAt time.Time
	query := `SELECT user_id, expires_at FROM refresh_tokens WHERE token = $1`
	err := r.pool.QueryRow(ctx, query, token).Scan(&userID, &expiresAt)
	return userID, expiresAt, err
}
