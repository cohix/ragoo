package embedder

import "log/slog"

type ollamaEmbedder struct {
	config map[string]string
}

// Generate generates embeddings for the given input
func (o *ollamaEmbedder) Generate(input string) (*Result, error) {
	slog.Info("generating embedding", "input", input)

	r := &Result{
		Embedding: []float32{0.1, 0.2},
	}

	return r, nil
}
