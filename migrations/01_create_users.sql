-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       avatar TEXT,
                       name TEXT NOT NULL,
                       username TEXT NOT NULL UNIQUE,
                       email TEXT UNIQUE,
                       email_verified BOOLEAN NOT NULL DEFAULT FALSE,
                       password TEXT NOT NULL,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
