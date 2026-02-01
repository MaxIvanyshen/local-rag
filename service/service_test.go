package service

import (
	"context"
	"errors"
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
		t.Fatalf("expected to find new content, got none")
	}
}

func TestDeleteDocument(t *testing.T) {
	ctx := context.Background()

	documentName := "Delete Test Document"
	documentData := []byte("This is a document to be deleted.")

	// Process document
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

	// Verify it exists
	doc, err := db.GetDocumentByName(ctx, testDB, documentName)
	if err != nil {
		t.Fatalf("failed to get document: %v", err)
	}
	if doc == nil {
		t.Fatalf("document was not found after processing")
	}
	initialID := doc.ID

	// Delete document
	delReq := &DeleteDocumentRequest{
		DocumentName: documentName,
	}
	s, err = svc.DeleteDocument(ctx, delReq)
	if err != nil {
		t.Fatalf("failed to delete document: %v", err)
	}
	if !s.Success {
		t.Fatalf("document deletion reported failure")
	}

	// Verify it no longer exists
	doc, err = db.GetDocumentByName(ctx, testDB, documentName)
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected document not found, but got: %v", err)
	}

	// Verify chunks are deleted
	var chunkCount int64
	testDB.Model(&db.Chunk{}).Where("document_id = ?", initialID).Count(&chunkCount)
	if chunkCount != 0 {
		t.Fatalf("expected 0 chunks for deleted document, but found %d", chunkCount)
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

func TestReprocessDocument(t *testing.T) {
	ctx := context.Background()

	documentName := "Reprocess Test Document"

	// Process initial document
	initialData := []byte("This document contains a unique phrase: xyz123 old content.")
	s, err := svc.ProcessDocument(ctx, &ProcessDocumentRequest{
		DocumentName: documentName,
		DocumentData: initialData,
	})
	if err != nil {
		t.Fatalf("failed to process initial document: %v", err)
	}
	if !s.Success {
		t.Fatalf("initial document processing reported failure")
	}

	// Get initial document ID
	initialDoc, err := db.GetDocumentByName(ctx, testDB, documentName)
	if err != nil {
		t.Fatalf("failed to get initial document: %v", err)
	}
	if initialDoc == nil {
		t.Fatalf("initial document was not found")
	}
	initialID := initialDoc.ID

	// Search for unique old phrase to confirm it's there
	searchReq := &SearchRequest{
		Query: "xyz123",
	}
	results, err := svc.Search(ctx, searchReq)
	if err != nil {
		t.Fatalf("search failed for initial content: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected to find initial content, got none")
	}

	// Reprocess with new content
	newData := []byte("This document has a different unique phrase: abc456 new content.")
	s, err = svc.ProcessDocument(ctx, &ProcessDocumentRequest{
		DocumentName: documentName,
		DocumentData: newData,
	})
	if err != nil {
		t.Fatalf("failed to reprocess document: %v", err)
	}
	if !s.Success {
		t.Fatalf("reprocess document reported failure")
	}

	// Verify document was replaced (different ID)
	newDoc, err := db.GetDocumentByName(ctx, testDB, documentName)
	if err != nil {
		t.Fatalf("failed to get new document: %v", err)
	}
	if newDoc == nil {
		t.Fatalf("new document was not found")
	}
	if newDoc.ID == initialID {
		t.Fatalf("document ID was not changed, expected replacement")
	}

	// Search for new unique phrase - should find
	searchReq = &SearchRequest{
		Query: "abc456",
	}
	results, err = svc.Search(ctx, searchReq)
	if err != nil {
		t.Fatalf("search failed for new content: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected to find new content, got none")
	}
}

func TestDocumentNameSemanticSearch(t *testing.T) {
	ctx := context.Background()

	// Test that documents with similar names are findable and marked correctly
	// Process a document with a distinctive name
	documentName1 := "QuantumComputingResearch2024"
	documentData1 := []byte("This document contains information about traditional computing systems. CPUs process instructions sequentially. Memory storage uses traditional binary representation.")

	s, err := svc.ProcessDocument(ctx, &ProcessDocumentRequest{
		DocumentName: documentName1,
		DocumentData: documentData1,
	})
	if err != nil {
		t.Fatalf("failed to process document: %v", err)
	}
	if !s.Success {
		t.Fatalf("document processing reported failure")
	}

	// Verify that document name embedding was saved
	doc, err := db.GetDocumentByName(ctx, testDB, documentName1)
	if err != nil {
		t.Fatalf("failed to get document: %v", err)
	}

	// Directly check the database for name embedding
	var count int64
	testDB.Raw("SELECT COUNT(*) FROM document_name_embeddings WHERE document_id = ?", doc.ID).Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 name embedding for document, got %d", count)
	}

	// Now search - the document should be findable
	searchReq := &SearchRequest{
		Query: "Quantum Computing",
	}

	results, err := svc.Search(ctx, searchReq)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	// The document should be in results
	found := false
	for _, result := range results {
		if result.DocumentName == documentName1 {
			found = true
			t.Logf("Found document '%s' with IsNameMatch=%v, distance=%.4f", result.DocumentName, result.IsNameMatch, result.Distance)
			break
		}
	}

	if !found {
		t.Logf("Document not found in results. Results count: %d", len(results))
		t.Logf("Results: %+v", results)
		t.Fatalf("expected to find document '%s' in search results", documentName1)
	}

	// The test passes if the document is found in results
	// (it may be found as a name match or content match depending on semantic similarity)
}

func TestDocumentNameEmbeddingCleanup(t *testing.T) {
	ctx := context.Background()

	documentName := "Cleanup Test Document"
	documentData := []byte("Test content for cleanup.")

	// Process document
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

	// Get document ID
	doc, err := db.GetDocumentByName(ctx, testDB, documentName)
	if err != nil {
		t.Fatalf("failed to get document: %v", err)
	}
	if doc == nil {
		t.Fatalf("document was not found")
	}

	// Verify name embedding exists
	var count int64
	testDB.Raw("SELECT COUNT(*) FROM document_name_embeddings WHERE document_id = ?", doc.ID).Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 name embedding for document, got %d", count)
	}

	// Delete document
	delReq := &DeleteDocumentRequest{
		DocumentName: documentName,
	}
	s, err = svc.DeleteDocument(ctx, delReq)
	if err != nil {
		t.Fatalf("failed to delete document: %v", err)
	}
	if !s.Success {
		t.Fatalf("document deletion reported failure")
	}

	// Verify name embedding was deleted
	testDB.Raw("SELECT COUNT(*) FROM document_name_embeddings WHERE document_id = ?", doc.ID).Scan(&count)
	if count != 0 {
		t.Fatalf("expected 0 name embeddings after deletion, but found %d", count)
	}
}
