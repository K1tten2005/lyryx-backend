-- +goose Up
-- +goose StatementBegin
CREATE TABLE annotation (
    id BIGSERIAL PRIMARY KEY,
    song_id BIGINT NOT NULL REFERENCES song(id) ON DELETE CASCADE,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    start_index INT NOT NULL, -- начальный символ выделения в lyrics
    end_index INT NOT NULL,   -- конечный символ выделения в lyrics
    rating INT NOT NULL DEFAULT 0, -- сумма плюсов и минусов
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_annotations_song_id ON annotation(song_id);

CREATE TABLE annotation_vote (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    annotation_id BIGINT NOT NULL REFERENCES annotation(id) ON DELETE CASCADE,
    vote_value SMALLINT NOT NULL CHECK (vote_value IN (-1, 1)),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, annotation_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS annotation_vote;
DROP TABLE IF EXISTS annotation;
-- +goose StatementEnd
