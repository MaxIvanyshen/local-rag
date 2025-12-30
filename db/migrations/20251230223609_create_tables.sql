-- +goose Up
-- +goose StatementBegin
CREATE TABLE document (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chunks (
    id TEXT PRIMARY KEY,
    document_id TEXT,
    chunk_index INTEGER NOT NULL,
    data BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE VIRTUAL TABLE chunk_embeddings USING vec0(
    embedding float[768]
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE chunk_embeddings;
DROP TABLE chunks;
DROP TABLE document;
-- +goose StatementEnd
