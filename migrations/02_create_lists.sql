-- +goose Up
-- +goose StatementBegin
CREATE TABLE lists (
                       id UUID PRIMARY KEY,
                       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                       image TEXT,
                       title TEXT NOT NULL,
                       notes TEXT,
                       is_public BOOLEAN NOT NULL DEFAULT TRUE,
                       share_token VARCHAR(32) NOT NULL UNIQUE,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       CONSTRAINT lists_share_token_len CHECK (char_length(share_token) = 32)
);

CREATE INDEX idx_lists_user_id_created_at_desc ON lists (user_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_lists_user_id_created_at_desc;
DROP TABLE IF EXISTS lists;
-- +goose StatementEnd
