-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE SCHEMA IF NOT EXISTS link;

CREATE TABLE link.links
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    original_url VARCHAR(255) UNIQUE NOT NULL,
    short_url VARCHAR(10) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS link.links;
DROP SCHEMA IF EXISTS link;
-- +goose StatementEnd