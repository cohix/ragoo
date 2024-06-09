package storage

import (
	"log/slog"
	"sync"
)

var (
	active = map[string]map[string]Storage{} // store active storage by type and name
	lock   = sync.Mutex{}                    // protect access to active storage for concurrent access
)

// Storage represents an embedder
type Storage interface {
	InsertEmbedding(collection string, ref string, embedding []float32, batch string) (*Result, error)
	LookupCosine(collection string, embedding []float32, limit int, threshold float32) (*Result, error)
	Cleanup(collection string, batch string) error
}

// Result is the result of an embedder
type Result struct {
	Refs    []string
	Cosines []float32
}

// StorageOfType returns storage for the provided type
func StorageOfType(name, stType string, config map[string]string) Storage {
	lock.Lock()
	defer lock.Unlock()

	// first check to see if storage of the given type and name is already active
	if actType, exists := active[stType]; exists {
		if str, exists := actType[name]; exists {
			return str
		}
	}

	var str Storage
	switch stType {
	case "duckdb":
		str = &duckDBStorage{config, nil, map[string]bool{}}
	}

	if str == nil {
		slog.Warn("no storage found for", "type", stType, "name", name)
		return nil
	}

	if actType, exists := active[stType]; exists {
		actType[name] = str
	} else {
		active[stType] = map[string]Storage{name: str}
	}

	return str
}
