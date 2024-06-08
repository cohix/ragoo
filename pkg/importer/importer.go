package importer

// Importer represents an importer for a given data source
type Importer interface {
	Run(chan Result) error
}

// Result is the result of an importer
type Result struct {
	Ref    string   `json:"ref"`
	Chunks []string `json:"chunks"`
}

// ImporterForType provides an importer for the given type
func ImporterForType(imType string, config map[string]string) Importer {
	switch imType {
	case "file":
		return &fileImporter{config}
	}

	return nil
}
