package provider

import "context"

// EmbeddingProvider defines an interface for fetching vector embeddings.
type EmbeddingProvider interface {
	EmbedTexts(ctx context.Context, texts []string) ([][]float32, error)
}
