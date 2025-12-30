package db

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SaveChunk inserts a chunk and its embedding into the database.
// Assumes the document already exists.
func SaveChunk(db *gorm.DB, documentID string, chunkIndex int, data []byte, embedding []float32) error {
	// Generate UUID for chunk
	chunkID := uuid.New().String()

	// Create chunk using GORM
	chunk := Chunk{
		ID:         chunkID,
		DocumentID: documentID,
		ChunkIndex: chunkIndex,
		Data:       data,
	}
	if err := db.Create(&chunk).Error; err != nil {
		return fmt.Errorf("failed to insert chunk: %w", err)
	}

	// Convert embedding to JSON string
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// Insert into chunk_embeddings virtual table using raw SQL
	err = db.Exec(`
		INSERT INTO chunk_embeddings (rowid, embedding)
		VALUES ((SELECT last_insert_rowid()), ?)`, string(embeddingJSON)).Error
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

func SearchChunks(db *gorm.DB, queryEmbedding []float32, limit int) ([]SearchResult, error) {
	// Serialize query embedding to JSON
	queryJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	// KNN query with raw SQL using GORM
	var results []SearchResult
	err = db.Raw(`
        SELECT c.id, c.document_id, c.chunk_index, c.data, knn.distance
        FROM chunks c
        JOIN (SELECT rowid, distance FROM chunk_embeddings WHERE embedding MATCH ? ORDER BY distance LIMIT ?) knn
        ON c.rowid = knn.rowid`, string(queryJSON), limit).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	return results, nil
}
