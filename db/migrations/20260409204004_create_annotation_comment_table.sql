-- +goose Up
-- +goose StatementBegin
CREATE TABLE annotation_comment (
    id BIGSERIAL PRIMARY KEY,
    annotation_id BIGINT NOT NULL REFERENCES annotation(id) ON DELETE CASCADE,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS annotation_comment;
-- +goose StatementEnd
