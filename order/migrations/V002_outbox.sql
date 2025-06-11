-- +goose Up
CREATE TABLE outbox (
    id SERIAL,
    message JSONB NOT NULL
)

-- +goose Down
DROP TABLE outbox;
