package embedding

import (
	"context"
	"net/http"
)

type Embedder interface {
	GenerateEmbedding(ctx context.Context, input []byte) ([]float32, error)
}

type TextEmbedder interface {
	Embedder

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
