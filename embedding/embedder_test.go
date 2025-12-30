package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOllamaEmbedder_GenerateEmbedding_Success(t *testing.T) {
	// Mock Ollama server
	expectedEmbedding := make([]float32, 768)
	for i := range expectedEmbedding {
		expectedEmbedding[i] = float32(i) * 0.01
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/embeddings", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req OllamaEmbeddingRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "text-embedding-3-small", req.Model)
		assert.Equal(t, "test prompt", req.Prompt)

		resp := OllamaEmbeddingResponse{Embedding: expectedEmbedding}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create embedder with test client and base URL
	client := server.Client()
	embedder := NewOllamaEmbedder("text-embedding-3-small", WithHttpClient(client), WithBaseURL(server.URL))

	// Test the embedding generation
	embedding, err := embedder.GenerateEmbedding(context.Background(), "test prompt")
	require.NoError(t, err)
	assert.Equal(t, expectedEmbedding, embedding)
}

func TestOllamaEmbedder_GenerateEmbedding_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := server.Client()
	embedder := NewOllamaEmbedder("text-embedding-3-small", WithHttpClient(client), WithBaseURL(server.URL))

	_, err := embedder.GenerateEmbedding(context.Background(), "test")
	assert.Error(t, err)
}

func TestOllamaEmbedder_GenerateEmbedding_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := server.Client()
	embedder := NewOllamaEmbedder("text-embedding-3-small", WithHttpClient(client), WithBaseURL(server.URL))

	_, err := embedder.GenerateEmbedding(context.Background(), "test")
	assert.Error(t, err)
}

func TestOllamaEmbedder_IntegrationWithActualOllama(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test with actual Ollama")
	}

	embedder := NewOllamaEmbedder("nomic-embed-text")

	embedding, err := embedder.GenerateEmbedding(context.Background(), "This is a test sentence for embedding.")
	t.Log(embedding)

	require.NoError(t, err)
	assert.Len(t, embedding, 768)   // nomic-embed-text outputs 768 dimensions
	assert.NotZero(t, embedding[0]) // Ensure it's not all zeros
}
