-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE SCHEMA IF NOT EXISTS link;

CREATE TABLE link.links
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    original_link VARCHAR(255) NOT NULL,
    short_link VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_original_link UNIQUE (original_link),
    CONSTRAINT unique_short_link UNIQUE (short_link)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS link.links;
DROP SCHEMA IF EXISTS link;
-- +goose StatementEnd