package db

import (
	"os"
	"testing"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	os.Remove("test.db")
	sqlite_vec.Auto()
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables manually for tests
	err = db.Exec(`
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
	`).Error
	require.NoError(t, err)

	return db
}

func TestSaveChunk(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Insert a document first
	docID := "test-doc-1"
	err := db.Exec("INSERT INTO document (id, name) VALUES (?, ?)", docID, "test document").Error
	require.NoError(t, err)

	// Test data
	chunkIndex := 0
	data := []byte("test chunk data")
	embedding := make([]float32, 768)
	for i := range embedding {
		embedding[i] = 0.1
	}

	// Save chunk
	err = SaveChunk(db, docID, chunkIndex, data, embedding)
	require.NoError(t, err)

	// Verify chunk was inserted
	var count int
	err = sqlDB.QueryRow("SELECT COUNT(*) FROM chunks WHERE document_id = ? AND chunk_index = ?", docID, chunkIndex).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify data was stored correctly
	var retrievedData []byte
	err = sqlDB.QueryRow("SELECT data FROM chunks WHERE document_id = ? AND chunk_index = ?", docID, chunkIndex).Scan(&retrievedData)
	require.NoError(t, err)
	assert.Equal(t, data, retrievedData)

	// Verify embedding was inserted (check if rowid exists in chunk_embeddings)
	var embeddingCount int
	err = sqlDB.QueryRow("SELECT COUNT(*) FROM chunk_embeddings").Scan(&embeddingCount)
	require.NoError(t, err)
	assert.Equal(t, 1, embeddingCount)
}

func TestSearchChunks(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Insert a document
	docID := "test-doc-2"
	err := db.Exec("INSERT INTO document (id, name) VALUES (?, ?)", docID, "test document").Error
	require.NoError(t, err)

	// Save two chunks
	embedding1 := make([]float32, 768)
	embedding1[0] = 1.0
	embedding2 := make([]float32, 768)
	embedding2[1] = 1.0
	err = SaveChunk(db, docID, 0, []byte("chunk 0"), embedding1)
	require.NoError(t, err)
	err = SaveChunk(db, docID, 1, []byte("chunk 1"), embedding2)
	require.NoError(t, err)

	// Search with the same embedding as embedding1
	results, err := SearchChunks(db, embedding1, 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// First result should be the closest (chunk 0)
	assert.Equal(t, 0, results[0].ChunkIndex)
	assert.Equal(t, []byte("chunk 0"), results[0].Data)
	assert.Equal(t, docID, results[0].DocumentID)
}
