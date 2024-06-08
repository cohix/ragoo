package embedder

// Embedder represents an embedder
type Embedder interface {
	Generate(input string) (*Result, error)
}

// Result is the result of an embedder
type Result struct {
	Embedding []float32
}

func EmbedderOfType(embType string, config map[string]string) Embedder {
	switch embType {
	case "ollama":
		return &ollamaEmbedder{config}
	}

	return nil
}
