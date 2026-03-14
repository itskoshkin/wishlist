-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       avatar TEXT,
                       name TEXT NOT NULL,
                       username TEXT NOT NULL,
                       email TEXT UNIQUE,
                       email_verified BOOLEAN NOT NULL DEFAULT FALSE,
                       password TEXT NOT NULL,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       CONSTRAINT users_username_format_check CHECK (lower(username) ~ '^[a-z0-9а-я_-]+$')
);

CREATE UNIQUE INDEX users_username_lower_key ON users ((lower(username)));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
