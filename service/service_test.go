package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/MaxIvanyshen/local-rag/chunker"
	"github.com/MaxIvanyshen/local-rag/config"
	"github.com/MaxIvanyshen/local-rag/db"
	"github.com/MaxIvanyshen/local-rag/embedding"
	"gorm.io/gorm"
)

func createEmbedder(cfg *config.Config) (embedding.Embedder, error) {
	switch cfg.Embedder.Type {
	case "ollama":
		return embedding.NewOllamaEmbedder(cfg.Embedder.Model, embedding.WithBaseURL(cfg.Embedder.BaseURL)), nil
	case "http":
		return embedding.NewHTTPEmbedder(cfg.Embedder.BaseURL), nil
	default:
		return nil, fmt.Errorf("unknown embedder type: %s", cfg.Embedder.Type)
	}
}

func createChunker(cfg *config.Config) (chunker.Chunker, error) {
	switch cfg.Chunker.Type {
	case "paragraph":
		return chunker.NewParagraphChunker(cfg.Chunker.OverlapBytes), nil
	case "fixed":
		return &chunker.FixedSizeChunker{ChunkSize: cfg.Chunker.ChunkSize}, nil
	default:
		return nil, fmt.Errorf("unknown chunker type: %s", cfg.Chunker.Type)
	}
}

var (
	svc    *Service
	testDB *gorm.DB
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	cfg := config.GetConfig(ctx)

	testDB = db.SetupTestDB()

	embedder, err := createEmbedder(cfg)
	if err != nil {
		panic(err)
	}

	chunker, err := createChunker(cfg)
	if err != nil {
		panic(err)
	}

	svc = NewService(&ServiceParameters{
		DB:       testDB,
		Embedder: embedder,
		Chunker:  chunker,
		Cfg:      cfg,
	})

	m.Run()
}

func TestDocProcessingAndSearch(t *testing.T) {
	ctx := context.Background()

	documentName := "Test Document"
	documentData := []byte("This is a test document. It contains several sentences. The purpose is to test document processing and search functionality.")

	s, err := svc.ProcessDocument(ctx, &ProcessDocumentRequest{
		DocumentName: documentName,
		DocumentData: documentData,
	})
	if err != nil {
		t.Fatalf("failed to process document: %v", err)
	}

	if !s.Success {
		t.Fatalf("document processing reported failure")
	}

	searchReq := &SearchRequest{
		Query: "test document",
	}

	results, err := svc.Search(ctx, searchReq)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected search results, got none")
	}
}

func TestBatchProcessDocuments(t *testing.T) {
	ctx := context.Background()

	reqs := []*ProcessDocumentRequest{
		{
			DocumentName: "Batch Doc 1",
			DocumentData: []byte("This is the first batch document. It has unique content for testing."),
		},
		{
			DocumentName: "Batch Doc 2",
			DocumentData: []byte("This is the second batch document. It contains different text."),
		},
		{
			DocumentName: "Batch Doc 3",
			DocumentData: []byte("Third document in batch. Testing concurrent processing."),
		},
	}

	s, err := svc.BatchProcessDocuments(ctx, &BatchProcessDocumentsRequest{Documents: reqs})
	if err != nil {
		t.Fatalf("failed to batch process documents: %v", err)
	}

	if !s.Success {
		t.Fatalf("batch processing reported failure")
	}

	// Verify documents were processed by searching for content from one
	searchReq := &SearchRequest{
		Query: "unique content",
	}

	results, err := svc.Search(ctx, searchReq)
	if err != nil {
		t.Fatalf("search failed after batch processing: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected search results after batch processing, got none")
	}
}

func TestBatchProcessDocumentsEmpty(t *testing.T) {
	ctx := context.Background()

	reqs := []*ProcessDocumentRequest{}

	s, err := svc.BatchProcessDocuments(ctx, &BatchProcessDocumentsRequest{Documents: reqs})
	if err != nil {
		t.Fatalf("failed to batch process empty list: %v", err)
	}

	if !s.Success {
		t.Fatalf("batch processing empty list reported failure")
	}
}

func TestProcessOpenAIGuideNotes(t *testing.T) {
	ctx := context.Background()

	data, err := os.ReadFile("../test_data/openai_agent_guide_notes.md")
	if err != nil {
		t.Fatalf("failed to read test data file: %v", err)
	}

	documentName := "OpenAI Agent Guide Notes"

	s, err := svc.ProcessDocument(ctx, &ProcessDocumentRequest{
		DocumentName: documentName,
		DocumentData: data,
	})
	if err != nil {
		t.Fatalf("failed to process document: %v", err)
	}

	if !s.Success {
		t.Fatalf("document processing reported failure")
	}

	// Search for a specific term from the document
	searchReq := &SearchRequest{
		Query: "agent design foundations",
	}

	results, err := svc.Search(ctx, searchReq)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	for _, res := range results {
		t.Logf("Found chunk: %s", res.Content)
	}

	if len(results) == 0 {
		t.Fatalf("expected search results for processed OpenAI guide, got none")
	}
}
