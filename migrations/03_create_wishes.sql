-- +goose Up
-- +goose StatementBegin
CREATE TABLE wishes (
                        id UUID PRIMARY KEY,
                        list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
                        image TEXT,
                        title TEXT NOT NULL,
                        notes TEXT,
                        link TEXT,
                        price BIGINT,
                        currency VARCHAR(8),
                        reserved_by UUID REFERENCES users(id) ON DELETE SET NULL,
                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        CONSTRAINT wishes_price_non_negative CHECK (price IS NULL OR price >= 0)
);

CREATE INDEX idx_wishes_list_id_created_at_asc ON wishes (list_id, created_at ASC);
CREATE INDEX idx_wishes_reserved_by ON wishes (reserved_by);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_wishes_reserved_by;
DROP INDEX IF EXISTS idx_wishes_list_id_created_at_asc;
DROP TABLE IF EXISTS wishes;
-- +goose StatementEnd
