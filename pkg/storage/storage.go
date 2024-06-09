package storage

// Storage represents an embedder
type Storage interface {
	InsertEmbedding(embedding []float32, collection string, ref string, batch string) (*Result, error)
	LookupCosine(embedding []float32, collection string, limit int, threshold float32) (*Result, error)
	Cleanup(collection string, currentBatch string) error
}

// Result is the result of an embedder
type Result struct {
	Refs    []string
	Cosines []float32
}

// StorageOfType returns storage for the provided type
func StorageOfType(strType string, config map[string]string) Storage {
	switch strType {
	case "duckdb":
		return &duckDBStorage{config, nil, map[string]bool{}}
	}

	return nil
}
