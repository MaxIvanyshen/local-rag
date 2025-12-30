package embedding

import (
	"context"
	"net/http"
)

type TextEmbedder interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	SetHttpClient(*http.Client)
}

type HttpRequestEmbedder struct {
	httpClient *http.Client
}

func (he *HttpRequestEmbedder) SetHttpClient(client *http.Client) {
	he.httpClient = client
}

type Option func(TextEmbedder)

func WithHttpClient(client *http.Client) Option {
	return func(te TextEmbedder) {
		te.SetHttpClient(client)
	}
}
