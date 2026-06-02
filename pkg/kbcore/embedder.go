package kbcore

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

func newEmbedder() (embedding.Embedder, error) {
	model := os.Getenv("NAUTIKIT_EMBEDDING_MODEL")
	if model == "" {
		model = "text-embedding-3-small"
	}

	baseURL := os.Getenv("NAUTIKIT_EMBEDDING_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	baseURL = strings.TrimSuffix(baseURL, "/embeddings")

	return openai.NewEmbedder(context.Background(), &openai.EmbeddingConfig{
		APIKey:  os.Getenv("NAUTIKIT_EMBEDDING_API_KEY"),
		Model:   model,
		BaseURL: baseURL,
		Timeout: 30 * time.Second,
	})
}
