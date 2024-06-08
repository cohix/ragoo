package storage

type duckDBStorage struct {
	config map[string]string
}

func (d *duckDBStorage) VectorSimilarity(embedding []float32, collection string, limit int) (*Result, error) {
	r := &Result{
		Entries: []string{"a relevant document"},
	}

	return r, nil
}
