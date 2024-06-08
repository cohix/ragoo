package storage

// Storage represents an embedder
type Storage interface {
	VectorSimilarity(embedding []float32, collection string, limit int) (*Result, error)
}

// Result is the result of an embedder
type Result struct {
	Entries []string
}

// StorageOfType returns storage for the provided type
func StorageOfType(strType string, config map[string]string) Storage {
	switch strType {
	case "duckdb":
		return &duckDBStorage{config}
	}

	return nil
}
