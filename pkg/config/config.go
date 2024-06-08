package config

// Config represents the full config for the ragoo app
type Config struct {
	Routes    []Route    `json:"routes" yaml:"routes"`
	Workflows []Workflow `json:"workflows" yaml:"workflows"`
	Services  []Service  `json:"services" yaml:"services"`
	Importers []Importer `json:"importers" yaml:"importers"`
	Embedders []Embedder `json:"embedders" yaml:"embedders"`
	Storage   []Storage  `json:"storage" yaml:"storage"`
	Tools     []Tool     `json:"tools" yaml:"tools"`
}

type Ref struct {
	Ref    string            `json:"ref" yaml:"ref"`
	Params map[string]string `json:"params" yaml:"params"`
}

// Route represents a route made available on the server and the workflow that gets triggered
type Route struct {
	Path     string `json:"path" yaml:"path"`
	Workflow Ref    `json:"workflow" yaml:"workflow"`
}

type Workflow struct {
	Name   string  `json:"name" yaml:"name"`
	Stages []Stage `json:"stages" yaml:"stages"`
}

type Stage struct {
	Name  string `json:"name" yaml:"name"`
	Steps []Step `json:"steps" yaml:"steps"`
}

type Step struct {
	Type   string            `json:"type" yaml:"type"`
	Action string            `json:"action" yaml:"action"`
	Ref    string            `json:"ref" yaml:"ref"`
	Params map[string]string `json:"params" yaml:"params"`
	Var    string            `json:"var" yaml:"var"`
}

// Service represents an LLM service and its configuration
type Service struct {
	Name   string            `json:"name" yaml:"name"`
	Type   string            `json:"type" yaml:"type"`
	Config map[string]string `json:"config" yaml:"config"`
}

// Importer represents an import source and its configuration
type Importer struct{}

// Embedder represents an embedding provider and its configuration
type Embedder struct {
	Name   string            `json:"name" yaml:"name"`
	Type   string            `json:"type" yaml:"type"`
	Config map[string]string `json:"config" yaml:"config"`
}

// Storage represents a vector db and its configuration
type Storage struct {
	Name   string            `json:"name" yaml:"name"`
	Type   string            `json:"type" yaml:"type"`
	Config map[string]string `json:"config" yaml:"config"`
}

// Tool represents a tool available to a service or workflow
type Tool struct{}
