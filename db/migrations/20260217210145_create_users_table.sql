-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE users (
    id               SERIAL       PRIMARY KEY,
    username         VARCHAR(255) NOT NULL UNIQUE,
    email            VARCHAR(255) NOT NULL UNIQUE,
    password_hash    VARCHAR(255) NOT NULL,
    avatar_url       VARCHAR(255),
    bio              TEXT,
    role             VARCHAR(50)  NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'moderator')),
    reputation_score INTEGER      DEFAULT 0,
    refresh_token    TEXT,
    created_at       TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);

-- При миграции добавляем модератора с паролем moder
INSERT INTO users (username, email, password_hash, role)
VALUES ('moder', 'moder@mail.ru', '$2a$12$sxI.ki6dKpRYBOBAlo5KZ.WFW7gdU3F/3tHdNmiywm25FAC8u8ecy', 'moderator');

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';
-- +goose StatementEnd

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column;
DROP TABLE IF EXISTS users;