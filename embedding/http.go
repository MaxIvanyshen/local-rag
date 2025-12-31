package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type HTTPEmbeddingRequest struct {
	Text string `json:"text"`
}

type HTTPEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type HTTPEmbedder struct {
	HttpRequestEmbedder

	baseURL string
}

func NewHTTPEmbedder(baseURL string) *HTTPEmbedder {
	return &HTTPEmbedder{
		baseURL: baseURL,
	}
}

func (he *HTTPEmbedder) GenerateEmbedding(ctx context.Context, input []byte) ([]float32, error) {
	req := HTTPEmbeddingRequest{
		Text: string(input),
	}

	body, err := json.Marshal(req)
	if err != nil {
		slog.Error("failed to marshal request", slog.String("error", err.Error()))
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", he.baseURL, bytes.NewBuffer(body))
	if err != nil {
		slog.Error("failed to create HTTP request", slog.String("error", err.Error()))
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := he.httpClient.Do(httpReq)
	if err != nil {
		slog.Error("HTTP request failed", slog.String("error", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	var embeddingResp HTTPEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		slog.Error("failed to decode response", slog.String("error", err.Error()))
		return nil, err
	}

	return embeddingResp.Embedding, nil
}
