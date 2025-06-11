-- +goose Up
CREATE TYPE order_status AS ENUM ('new', 'finished', 'cancelled');

-- +goose Down
DROP TYPE order_status;

-- +goose Up
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    description TEXT,
    amount BIGINT NOT NULL,
    status order_status NOT NULL DEFAULT 'new'
);

-- +goose Down
DROP TABLE orders;
