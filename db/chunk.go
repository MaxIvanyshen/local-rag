package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SaveChunk(ctx context.Context, db *gorm.DB, documentID string, chunkIndex int, data []byte, embedding []float32) error {
	chunk := Chunk{
		ID:         uuid.New().String(),
		DocumentID: documentID,
		ChunkIndex: chunkIndex,
		Data:       data,
	}
	if err := db.WithContext(ctx).Create(&chunk).Error; err != nil {
		return fmt.Errorf("failed to insert chunk: %w", err)
	}

	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// Insert into chunk_embeddings virtual table using raw SQL
	err = db.WithContext(ctx).Exec(`
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

func SearchChunks(ctx context.Context, db *gorm.DB, queryEmbedding []float32, limit int) ([]SearchResult, error) {
	queryJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	// KNN query with raw SQL using GORM
	var results []SearchResult
	err = db.WithContext(ctx).Raw(`
        SELECT c.id, c.document_id, c.chunk_index, c.data, knn.distance
        FROM chunks c
        JOIN (SELECT rowid, distance FROM chunk_embeddings WHERE embedding MATCH ? ORDER BY distance LIMIT ?) knn
        ON c.rowid = knn.rowid`, string(queryJSON), limit).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	return results, nil
}
