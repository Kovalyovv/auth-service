package postgres

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/Kovalyovv/auth-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("could not start postgres container: %s", err)
	}

	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("could not stop postgres container: %s", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("could not get connection string: %s", err)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("could not connect to test database: %s", err)
	}

	m.Run()
}

func setupTables(t *testing.T, ctx context.Context) {
	_, err := testPool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(50) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            created_at TIMESTAMPTZ DEFAULT NOW()
        );
        CREATE TABLE IF NOT EXISTS refresh_tokens (
            id SERIAL PRIMARY KEY,
            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            token TEXT NOT NULL UNIQUE,
            expires_at TIMESTAMPTZ NOT NULL,
            created_at TIMESTAMPTZ DEFAULT NOW()
        );
    `)
	require.NoError(t, err)
}

func cleanupTables(t *testing.T, ctx context.Context) {
	_, err := testPool.Exec(ctx, "DROP TABLE IF EXISTS refresh_tokens, users;")
	require.NoError(t, err)
}

func TestUserRepo_ConsumeRefreshToken(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepo(testPool)

	setupTables(t, ctx)
	defer cleanupTables(t, ctx)

	user := &domain.User{Username: "test", Email: "test@test.com", PasswordHash: "hash"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("Given a valid and unexpired token", func(t *testing.T) {
		token := "valid-token"
		expiresAt := time.Now().Add(time.Hour)
		err := repo.SaveRefreshToken(ctx, user.ID, token, expiresAt)
		require.NoError(t, err)

		userID, err := repo.ConsumeRefreshToken(ctx, token)

		assert.NoError(t, err)
		assert.Equal(t, user.ID, userID)

		_, _, err = repo.GetRefreshToken(ctx, token)
		assert.Error(t, err, "token should have been deleted")
	})

	t.Run("Given a non-existent token", func(t *testing.T) {
		_, err := repo.ConsumeRefreshToken(ctx, "non-existent-token")

		assert.ErrorIs(t, err, domain.ErrRefreshTokenNotFound)
	})

	t.Run("Given an expired token", func(t *testing.T) {
		token := "expired-token"
		expiresAt := time.Now().Add(-time.Hour)
		err := repo.SaveRefreshToken(ctx, user.ID, token, expiresAt)
		require.NoError(t, err)

		_, err = repo.ConsumeRefreshToken(ctx, token)

		assert.ErrorIs(t, err, domain.ErrRefreshTokenNotFound)
	})
}
