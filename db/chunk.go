package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type Chunk struct {
	ID         string
	DocumentID string
	ChunkIndex int
	Data       []byte
	Embedding  []float32
}

// SaveChunk inserts a chunk and its embedding into the database.
// Assumes the document already exists.
func SaveChunk(db *sql.DB, documentID string, chunkIndex int, data []byte, embedding []float32) error {
	// Generate UUID for chunk
	chunkID := uuid.New().String()

	// Insert into chunks table
	_, err := db.Exec(`
		INSERT INTO chunks (id, document_id, chunk_index, data)
		VALUES (?, ?, ?, ?)`,
		chunkID, documentID, chunkIndex, data)
	if err != nil {
		return fmt.Errorf("failed to insert chunk: %w", err)
	}

	// Convert embedding to JSON string
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// Insert into chunk_embeddings virtual table
	_, err = db.Exec(`
		INSERT INTO chunk_embeddings (rowid, embedding)
		VALUES ((SELECT last_insert_rowid()), ?)`, string(embeddingJSON))
	if err != nil {
		return fmt.Errorf("failed to insert embedding: %w", err)
	}

	return nil
}

type SearchResult struct {
	ChunkID    string
	DocumentID string
	ChunkIndex int
	Data       []byte
	Distance   float64
}

func SearchChunks(db *sql.DB, queryEmbedding []float32, limit int) ([]SearchResult, error) {
	// Serialize query embedding to JSON
	queryJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	// KNN query with JOIN using subquery
	rows, err := db.Query(`
        SELECT c.id, c.document_id, c.chunk_index, c.data, knn.distance
        FROM chunks c
        JOIN (SELECT rowid, distance FROM chunk_embeddings WHERE embedding MATCH ? ORDER BY distance LIMIT ?) knn
        ON c.rowid = knn.rowid`, string(queryJSON), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query embeddings: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var chunkID, docID string
		var index int
		var data []byte
		var distance float64
		if err := rows.Scan(&chunkID, &docID, &index, &data, &distance); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		results = append(results, SearchResult{
			ChunkID: chunkID, DocumentID: docID, ChunkIndex: index, Data: data, Distance: distance,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return results, nil
}
