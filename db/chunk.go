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

	// Get the rowid of the inserted chunk
	var chunkRowID int64
	err = db.QueryRow("SELECT last_insert_rowid()").Scan(&chunkRowID)
	if err != nil {
		return fmt.Errorf("failed to get chunk rowid: %w", err)
	}

	// Convert embedding to JSON string
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// Insert into chunk_embeddings virtual table
	_, err = db.Exec(`
		INSERT INTO chunk_embeddings (rowid, embedding)
		VALUES (?, ?)`,
		chunkRowID, string(embeddingJSON))
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
	queryJSON, _ := json.Marshal(queryEmbedding)

	// KNN query
	rows, _ := db.Query(`
        SELECT rowid, distance 
        FROM chunk_embeddings 
        WHERE embedding MATCH ? 
        ORDER BY distance 
        LIMIT ?`, string(queryJSON), limit)

	var results []SearchResult
	for rows.Next() {
		var rowid int64
		var distance float64
		rows.Scan(&rowid, &distance)

		// Fetch chunk data
		var chunkID, docID string
		var index int
		var data []byte
		db.QueryRow(`
            SELECT id, document_id, chunk_index, data 
            FROM chunks 
            WHERE rowid = ?`, rowid).Scan(&chunkID, &docID, &index, &data)

		results = append(results, SearchResult{
			ChunkID: chunkID, DocumentID: docID, ChunkIndex: index, Data: data, Distance: distance,
		})
	}
	return results, nil
}
