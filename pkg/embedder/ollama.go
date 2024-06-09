package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ollamaEmbedder struct {
	config map[string]string
}

type embeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// Generate generates embeddings for the given input
func (o *ollamaEmbedder) Generate(input string) (*Result, error) {
	url := "http://localhost:11434/api/embeddings"

	reqBody := &embeddingRequest{
		Model:  "snowflake-arctic-embed",
		Prompt: input,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to json.Marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to NewRequest: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to Do: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	emb := &embeddingResponse{}
	if err := json.NewDecoder(resp.Body).Decode(emb); err != nil {
		return nil, fmt.Errorf("failed to NewDecoder.Decode: %w", err)
	}

	r := &Result{
		Embedding: emb.Embedding,
	}

	return r, nil
}
