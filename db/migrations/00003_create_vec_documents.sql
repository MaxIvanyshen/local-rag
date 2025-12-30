-- +goose Up
CREATE VIRTUAL TABLE IF NOT EXISTS vec_documents USING vec0(
	id INTEGER PRIMARY KEY,
	embedding FLOAT[1536]
);