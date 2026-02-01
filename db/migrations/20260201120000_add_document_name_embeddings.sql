-- +goose Up
-- +goose StatementBegin
CREATE VIRTUAL TABLE document_name_embeddings USING vec0(
    document_id TEXT,
    embedding float[768]
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE document_name_embeddings;
-- +goose StatementEnd
