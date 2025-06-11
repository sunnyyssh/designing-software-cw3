-- +goose Up
CREATE TABLE accounts (
    user_id UUID PRIMARY KEY,
    amount BIGINT NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE accounts;
