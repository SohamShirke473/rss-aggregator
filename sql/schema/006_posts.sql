-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMP NOT NULL DEFAULT NOW(),
    url TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE posts;
