package service

// Service represents an LLM service
type Service interface {
	Completion(prompt string) (*Result, error)
}

// Result is the result of an embedder
type Result struct {
	Completion string
}

// ServiceOfType returns a service for the provided type
func ServiceOfType(strType string, config map[string]string) Service {
	switch strType {
	case "ollama":
		return &ollamaService{config}
	}

	return nil
}
