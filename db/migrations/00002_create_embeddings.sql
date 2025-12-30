-- +goose Up
CREATE TABLE IF NOT EXISTS embeddings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	document_id INTEGER,
	embedding BLOB,
	FOREIGN KEY (document_id) REFERENCES documents(id)
);