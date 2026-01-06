package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/MaxIvanyshen/local-rag/chunker"
	"github.com/MaxIvanyshen/local-rag/config"
	"github.com/MaxIvanyshen/local-rag/db"
	"github.com/MaxIvanyshen/local-rag/embedding"
	"gorm.io/gorm"
)

type Service struct {
	db       *gorm.DB
	embedder embedding.Embedder
	chunker  chunker.Chunker
	cfg      *config.Config
}

type ServiceParameters struct {
	DB       *gorm.DB
	Embedder embedding.Embedder
	Chunker  chunker.Chunker
	Cfg      *config.Config
}

func NewService(params *ServiceParameters) *Service {
	return &Service{
		db:       params.DB,
		embedder: params.Embedder,
		chunker:  params.Chunker,
		cfg:      params.Cfg,
	}
}

type SearchRequest struct {
	Query string `json:"query"`
}

func (s *Service) Search(ctx context.Context, req *SearchRequest) ([]db.SearchResult, error) {
	slog.Info("received search request", slog.String("query", req.Query))

	// Generate embedding for the query
	queryEmbedding, err := s.embedder.GenerateEmbedding(ctx, []byte(req.Query))
	if err != nil {
		slog.Error("failed to generate embedding for the query", slog.String("error", err.Error()), slog.String("query", req.Query))
		return nil, err
	}

	// Search chunks using the generated embedding
	results, err := db.SearchChunks(ctx, s.db, queryEmbedding, s.cfg.Search.TopK)
	if err != nil {
		return nil, err
	}

	return results, nil
}

type ProcessDocumentRequest struct {
	DocumentName string `json:"document_name"`
	DocumentData []byte `json:"document_data"`
}

type DeleteDocumentRequest struct {
	DocumentName string `json:"document_name"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

func Success(s bool) *SuccessResponse {
	return &SuccessResponse{Success: s}
}

func (s *Service) ProcessDocument(ctx context.Context, req *ProcessDocumentRequest) (*SuccessResponse, error) {
	slog.Info("received process document request", slog.String("document_name", req.DocumentName))

	// Delete existing document with this name if it exists
	err := db.DeleteDocumentByName(ctx, s.db, req.DocumentName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("failed to delete existing document", slog.String("error", err.Error()), slog.String("document_name", req.DocumentName))
		return Success(false), err
	}

	// Save document to the database
	documentID, err := db.SaveDocument(ctx, s.db, req.DocumentName)
	if err != nil {
		slog.Error("failed to save document", slog.String("error", err.Error()), slog.String("document_name", req.DocumentName))
		return Success(false), err
	}

	// Chunk the document
	chunkResults := s.chunker.Chunk(req.DocumentData)

	for i, chunkResult := range chunkResults {
		// Generate embedding for the chunk
		embedding, err := s.embedder.GenerateEmbedding(ctx, chunkResult.Data)
		if err != nil {
			slog.Error("failed to generate embedding for chunk", slog.String("error", err.Error()), slog.String("document_name", req.DocumentName), slog.Int("chunk_index", i))
			return Success(false), err
		}

		// Save chunk and its embedding to the database
		err = db.SaveChunk(ctx, s.db, documentID, i, chunkResult.StartLine, chunkResult.EndLine, chunkResult.Data, embedding)
		if err != nil {
			slog.Error("failed to save chunk", slog.String("error", err.Error()), slog.String("document_name", req.DocumentName), slog.Int("chunk_index", i))
			return Success(false), err
		}
	}

	slog.Info("successfully processed document", slog.String("document_name", req.DocumentName))

	return Success(true), nil
}

func (s *Service) DeleteDocument(ctx context.Context, req *DeleteDocumentRequest) (*SuccessResponse, error) {
	err := db.DeleteDocumentByName(ctx, s.db, req.DocumentName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("failed to delete document", slog.String("error", err.Error()), slog.String("document_name", req.DocumentName))
		return Success(false), err
	}
	return Success(true), nil
}

type BatchProcessDocumentsRequest struct {
	Documents []*ProcessDocumentRequest `json:"documents"`
}

type BatchProcessResponse struct {
	FailedDocuments []string `json:"failed_documents"` // Names of documents that failed to process
}

func (s *Service) BatchProcessDocuments(ctx context.Context, req *BatchProcessDocumentsRequest) (*BatchProcessResponse, error) {
	if len(req.Documents) == 0 {
		slog.Info("nothing to process in batch request. Returning....")
		return &BatchProcessResponse{
			FailedDocuments: []string{},
		}, nil
	}

	reqChan := make(chan *ProcessDocumentRequest)

	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(s.cfg.BatchProcessing.WorkerCount)

	failed := make(chan string, len(req.Documents))

	for range s.cfg.BatchProcessing.WorkerCount {
		go func() {
			defer wg.Done()
			for req := range reqChan {
				s, err := s.ProcessDocument(ctx, req)
				if err != nil {
					slog.Error("failed to process document in batch", slog.String("error", err.Error()), slog.String("document_name", req.DocumentName))
				}

				if !s.Success {
					slog.Error("processing document in batch was not successful", slog.String("document_name", req.DocumentName))
					failed <- req.DocumentName
				}
			}
		}()
	}

	for _, req := range req.Documents {
		reqChan <- req
	}
	close(reqChan)

	failedDocuments := make([]string, 0)
	collectorDone := make(chan struct{})
	go func() {
		defer close(collectorDone)
		for docName := range failed {
			mu.Lock()
			failedDocuments = append(failedDocuments, docName)
			mu.Unlock()
		}
	}()

	wg.Wait()
	close(failed)
	<-collectorDone

	slog.Info("batch processing completed", slog.Int("total_documents", len(req.Documents)), slog.Int("failed_documents", len(failedDocuments)))

	return &BatchProcessResponse{
		FailedDocuments: failedDocuments,
	}, nil
}
