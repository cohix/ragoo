package importer

import "github.com/cohix/ragoo/pkg/storage"

// Importer represents an importer for a given data source
type Importer interface {
	Run(string, chan Result) error
	ResolveRefs(storage.Result) (*Result, error)
}

// Result is the result of an importer
type Result struct {
	Ref       string   `json:"ref"`
	Chunks    []string `json:"chunks"`
	Documents []string `json:"documents"`
	Batch     string   `json:"batch"`
}

// ImporterOfType provides an importer for the given type
func ImporterOfType(imType string, config map[string]string) Importer {
	switch imType {
	case "file":
		return &fileImporter{config}
	}

	return nil
}
