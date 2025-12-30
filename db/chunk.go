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
	ChunkID      string  `json:"chunk_id" gorm:"column:chunk_id"`
	DocumentID   string  `json:"document_id" gorm:"column:document_id"`
	DocumentName string  `json:"document_name" gorm:"column:document_name"`
	ChunkIndex   int     `json:"chunk_index" gorm:"column:chunk_index"`
	Data         []byte  `json:"data" gorm:"column:data"`
	Distance     float64 `json:"distance" gorm:"column:distance"`
}

func SearchChunks(ctx context.Context, db *gorm.DB, queryEmbedding []float32, limit int) ([]SearchResult, error) {
	queryJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	var results []SearchResult
	err = db.WithContext(ctx).Raw(`SELECT 
		c.id as chunk_id,
		c.document_id as document_id,
		d.name as document_name, 
		c.chunk_index as chunk_index,
		c.data as data, 
		knn.distance as distance 
		FROM chunks c 
		JOIN document d ON d.id = c.document_id 
		JOIN (
			SELECT rowid, distance 
			FROM chunk_embeddings 
			WHERE embedding MATCH ? 
			ORDER BY distance 
			LIMIT ?
		) knn ON c.rowid = knn.rowid`, string(queryJSON), limit).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	return results, nil
}
