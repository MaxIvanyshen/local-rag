package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveChunk(t *testing.T) {
	db := SetupTestDB()
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Insert a document first
	docID := "test-doc-1"
	err := db.Exec("INSERT INTO documents (id, name) VALUES (?, ?)", docID, "test document").Error
	require.NoError(t, err)

	// Test data
	chunkIndex := 0
	data := []byte("test chunk data")
	embedding := make([]float32, 768)
	for i := range embedding {
		embedding[i] = 0.1
	}

	// Save chunk
	err = SaveChunk(t.Context(), db, docID, chunkIndex, 1, 10, data, embedding)
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
	db := SetupTestDB()
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Insert a document
	docID := "test-doc-2"
	err := db.Exec("INSERT INTO documents (id, name) VALUES (?, ?)", docID, "test document").Error
	require.NoError(t, err)

	// Verify document is inserted
	var name string
	err = sqlDB.QueryRow("SELECT name FROM documents WHERE id = ?", docID).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "test document", name)

	// Save two chunks
	embedding1 := make([]float32, 768)
	embedding1[0] = 1.0
	embedding2 := make([]float32, 768)
	embedding2[1] = 1.0
	err = SaveChunk(t.Context(), db, docID, 0, 1, 10, []byte("chunk 0"), embedding1)
	require.NoError(t, err)
	err = SaveChunk(t.Context(), db, docID, 1, 11, 20, []byte("chunk 1"), embedding2)
	require.NoError(t, err)

	// Search with the same embedding as embedding1
	results, err := SearchChunks(t.Context(), db, embedding1, 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// First result should be the closest (chunk 0)
	assert.Equal(t, 0, results[0].ChunkIndex)
	assert.Equal(t, "chunk 0", results[0].Content)
	assert.Equal(t, docID, results[0].DocumentID)
	assert.Equal(t, "test document", results[0].DocumentName)
}
