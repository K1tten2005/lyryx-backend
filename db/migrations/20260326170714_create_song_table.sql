-- +goose Up
-- +goose StatementBegin
CREATE TABLE song (
    id           SERIAL       PRIMARY KEY,
    title        VARCHAR(255) NOT NULL,
    artist_id    INTEGER      NOT NULL REFERENCES artist(id) ON DELETE CASCADE,
    lyrics       TEXT         NOT NULL,
    cover_url    VARCHAR(255),
    release_date DATE,
    views        INTEGER      DEFAULT 0,
    created_at   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска всех песен артиста
CREATE INDEX idx_song_artist_id ON song(artist_id);

-- Триггер для автообновления updated_at
CREATE TRIGGER update_song_updated_at
    BEFORE UPDATE ON song
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_song_updated_at ON song;
DROP INDEX IF EXISTS idx_song_artist_id;
DROP TABLE IF EXISTS song;
-- +goose StatementEnd