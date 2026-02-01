package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SaveChunk(ctx context.Context, db *gorm.DB, documentID string, chunkIndex int, startLine, endLine int, data []byte, embedding []float32) error {
	chunk := Chunk{
		ID:         uuid.New().String(),
		DocumentID: documentID,
		ChunkIndex: chunkIndex,
		StartLine:  startLine,
		EndLine:    endLine,
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
	if err := db.WithContext(ctx).Exec(`
		INSERT INTO chunk_embeddings (chunk_id, embedding)
		VALUES (?, ?)`, chunk.ID, string(embeddingJSON)).Error; err != nil {
		return fmt.Errorf("failed to insert embedding: %w", err)
	}

	// Get the rowid of the inserted embedding
	var embeddingRowID int64
	if err := db.WithContext(ctx).Raw("SELECT last_insert_rowid()").Scan(&embeddingRowID).Error; err != nil {
		return fmt.Errorf("failed to get embedding rowid: %w", err)
	}

	// Update the chunk with the embedding rowid
	if err := db.WithContext(ctx).Model(&chunk).Update("embedding_rowid", embeddingRowID).Error; err != nil {
		return fmt.Errorf("failed to update chunk with embedding rowid: %w", err)
	}

	return nil
}

type SearchResult struct {
	ChunkID      string  `json:"chunk_id" gorm:"column:chunk_id"`
	DocumentID   string  `json:"document_id" gorm:"column:document_id"`
	DocumentName string  `json:"document_name" gorm:"column:document_name"`
	ChunkIndex   int     `json:"chunk_index" gorm:"column:chunk_index"`
	StartLine    int     `json:"start_line" gorm:"column:start_line"`
	EndLine      int     `json:"end_line" gorm:"column:end_line"`
	Content      string  `json:"data" gorm:"column:data"`
	Distance     float64 `json:"distance" gorm:"column:distance"`
	IsNameMatch  bool    `json:"is_name_match"`
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
		c.start_line as start_line,
		c.end_line as end_line,
		c.data as data,
		knn.distance as distance
		FROM chunks c
		JOIN documents d ON d.id = c.document_id
		JOIN (
			SELECT rowid, distance
			FROM chunk_embeddings
			WHERE embedding MATCH ?
			ORDER BY distance
			LIMIT ?
		) knn ON c.embedding_rowid = knn.rowid`, string(queryJSON), limit).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	return results, nil
}

type DocumentNameSearchResult struct {
	DocumentID string  `json:"document_id" gorm:"column:document_id"`
	Distance   float64 `json:"distance" gorm:"column:distance"`
}

func SaveDocumentNameEmbedding(ctx context.Context, db *gorm.DB, documentID string, embedding []float32) error {
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal document name embedding: %w", err)
	}

	if err := db.WithContext(ctx).Exec(`
		INSERT INTO document_name_embeddings (document_id, embedding)
		VALUES (?, ?)`, documentID, string(embeddingJSON)).Error; err != nil {
		return fmt.Errorf("failed to insert document name embedding: %w", err)
	}

	return nil
}

func SearchDocumentNames(ctx context.Context, db *gorm.DB, queryEmbedding []float32, limit int) ([]DocumentNameSearchResult, error) {
	queryJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	var results []DocumentNameSearchResult
	err = db.WithContext(ctx).Raw(`SELECT
		document_id as document_id,
		distance as distance
		FROM document_name_embeddings
		WHERE embedding MATCH ?
		ORDER BY distance
		LIMIT ?`, string(queryJSON), limit).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query document name embeddings: %w", err)
	}
	return results, nil
}
