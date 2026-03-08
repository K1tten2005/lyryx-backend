-- +goose Up
-- +goose StatementBegin
CREATE TABLE artist (
    id          SERIAL        PRIMARY KEY,
    name        VARCHAR(255)  NOT NULL     UNIQUE,
    bio         TEXT,
    avatar_url  VARCHAR(255),
    created_at  TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP     DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS artist;
-- +goose StatementEnd
