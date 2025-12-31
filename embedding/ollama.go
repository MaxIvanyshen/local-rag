package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type OllamaEmbedder struct {
	HttpRequestEmbedder

	modelName  string
	baseURL    string
	httpClient http.Client
}

func WithBaseURL(baseURL string) Option {
	return func(te TextEmbedder) {
		if oe, ok := te.(*OllamaEmbedder); ok {
			oe.baseURL = baseURL
		}
	}
}

func NewOllamaEmbedder(modelName string, opts ...Option) *OllamaEmbedder {
	oe := &OllamaEmbedder{
		modelName: modelName,
		baseURL:   "http://localhost:11434",
	}
	for _, opt := range opts {
		opt(oe)
	}
	return oe
}

func (oe *OllamaEmbedder) GenerateEmbedding(ctx context.Context, input []byte) ([]float32, error) {
	req := OllamaEmbeddingRequest{
		Model:  oe.modelName,
		Prompt: string(input),
	}

	body, err := json.Marshal(req)
	if err != nil {
		slog.Error("failed to marshal request", slog.String("error", err.Error()))
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", oe.baseURL+"/api/embeddings", bytes.NewBuffer(body))
	if err != nil {
		slog.Error("failed to create HTTP request", slog.String("error", err.Error()))
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := oe.httpClient.Do(httpReq)
	if err != nil {
		slog.Error("HTTP request failed", slog.String("error", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	var embeddingResp OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		slog.Error("failed to decode response", slog.String("error", err.Error()))
		return nil, err
	}

	return embeddingResp.Embedding, nil
}
