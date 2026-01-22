CREATE TABLE users
(
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(50)  NOT NULL,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users (email);

CREATE TABLE refresh_tokens
(
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER   NOT NULL REFERENCES users ON DELETE CASCADE,
    token      TEXT      NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens (token);