-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

-- =========================================================
-- IMMUTABLE WRAPPER FOR UNACCENT
-- Создаем строгую неизменяемую обертку, чтобы Postgres 
-- разрешил использовать ее в генерируемых колонках
-- =========================================================
CREATE OR REPLACE FUNCTION f_unaccent(text)
RETURNS text
LANGUAGE sql
IMMUTABLE STRICT
AS $$
    SELECT unaccent('unaccent', $1);
$$;

-- =========================================================
-- SONG SEARCH VECTOR
-- =========================================================
ALTER TABLE song
ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
    -- TITLE
    setweight(to_tsvector('simple'::regconfig, f_unaccent(coalesce(title, ''))), 'A') ||
    setweight(to_tsvector('english'::regconfig, f_unaccent(coalesce(title, ''))), 'A') ||
    setweight(to_tsvector('russian'::regconfig, f_unaccent(coalesce(title, ''))), 'A') ||
    -- LYRICS
    setweight(to_tsvector('simple'::regconfig, f_unaccent(substr(coalesce(lyrics, ''), 1, 15000))), 'C') ||
    setweight(to_tsvector('english'::regconfig, f_unaccent(substr(coalesce(lyrics, ''), 1, 15000))), 'B') ||
    setweight(to_tsvector('russian'::regconfig, f_unaccent(substr(coalesce(lyrics, ''), 1, 15000))), 'B')
) STORED;

-- Просто индексируем уже посчитанную колонку!
CREATE INDEX idx_song_search_vector ON song USING GIN (search_vector);
CREATE INDEX idx_song_title_trgm ON song USING GIN(title gin_trgm_ops);

-- =========================================================
-- ARTIST SEARCH VECTOR
-- =========================================================
ALTER TABLE artist
ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
    -- NAME
    setweight(to_tsvector('simple'::regconfig, f_unaccent(coalesce(name, ''))), 'A') ||
    setweight(to_tsvector('english'::regconfig, f_unaccent(coalesce(name, ''))), 'A') ||
    setweight(to_tsvector('russian'::regconfig, f_unaccent(coalesce(name, ''))), 'A') ||
    -- BIO
    setweight(to_tsvector('simple'::regconfig, f_unaccent(substr(coalesce(bio, ''), 1, 5000))), 'C') ||
    setweight(to_tsvector('english'::regconfig, f_unaccent(substr(coalesce(bio, ''), 1, 5000))), 'B') ||
    setweight(to_tsvector('russian'::regconfig, f_unaccent(substr(coalesce(bio, ''), 1, 5000))), 'B')
) STORED;

CREATE INDEX idx_artist_search_vector ON artist USING GIN (search_vector);
CREATE INDEX idx_artist_name_trgm ON artist USING GIN(name gin_trgm_ops);

-- =========================================================
-- USER SEARCH VECTOR
-- =========================================================
ALTER TABLE users
ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
    -- USERNAME
    setweight(to_tsvector('simple'::regconfig, f_unaccent(coalesce(username, ''))), 'A') ||
    setweight(to_tsvector('english'::regconfig, f_unaccent(coalesce(username, ''))), 'A') ||
    setweight(to_tsvector('russian'::regconfig, f_unaccent(coalesce(username, ''))), 'A') ||
    -- BIO
    setweight(to_tsvector('simple'::regconfig, f_unaccent(substr(coalesce(bio, ''), 1, 3000))), 'C') ||
    setweight(to_tsvector('english'::regconfig, f_unaccent(substr(coalesce(bio, ''), 1, 3000))), 'B') ||
    setweight(to_tsvector('russian'::regconfig, f_unaccent(substr(coalesce(bio, ''), 1, 3000))), 'B')
) STORED;

CREATE INDEX idx_users_search_vector ON users USING GIN (search_vector);
CREATE INDEX idx_users_username_trgm ON users USING GIN(username gin_trgm_ops);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_username_trgm;
DROP INDEX IF EXISTS idx_users_search_vector;
ALTER TABLE users DROP COLUMN IF EXISTS search_vector;

DROP INDEX IF EXISTS idx_artist_name_trgm;
DROP INDEX IF EXISTS idx_artist_search_vector;
ALTER TABLE artist DROP COLUMN IF EXISTS search_vector;

DROP INDEX IF EXISTS idx_song_title_trgm;
DROP INDEX IF EXISTS idx_song_search_vector;
ALTER TABLE song DROP COLUMN IF EXISTS search_vector;

DROP FUNCTION IF EXISTS build_search_query(text);
DROP FUNCTION IF EXISTS f_unaccent(text);
DROP EXTENSION IF EXISTS unaccent;
DROP EXTENSION IF EXISTS pg_trgm;
-- +goose StatementEnd