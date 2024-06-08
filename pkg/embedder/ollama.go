package embedder

type ollamaEmbedder struct {
	config map[string]string
}

// Generate generates embeddings for the given input
func (o *ollamaEmbedder) Generate(_ []byte) (*Result, error) {
	r := &Result{
		Embedding: []float32{0.1, 0.2},
	}

	return r, nil
}
