package kbcore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Embedder interface {
	Embed(text string) ([]float64, error)
}

type openaiEmbedder struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func NewEmbedder() Embedder {
	model := os.Getenv("NAUTIKIT_EMBEDDING_MODEL")
	if model == "" {
		model = "text-embedding-3-small"
	}
	baseURL := os.Getenv("NAUTIKIT_EMBEDDING_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/embeddings"
	}
	return &openaiEmbedder{
		apiKey:  os.Getenv("NAUTIKIT_EMBEDDING_API_KEY"),
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

func (e *openaiEmbedder) Embed(text string) ([]float64, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("NAUTIKIT_EMBEDDING_API_KEY is not set")
	}

	body, err := json.Marshal(embeddingRequest{Model: e.model, Input: text})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", e.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	return result.Data[0].Embedding, nil
}
