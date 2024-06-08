package runner

import "github.com/cohix/ragoo/pkg/config"

// Runner is an orchestrator for workflows and importers
type Runner struct {
	config *config.Config
}

// Result is the result of a workflow
type Result struct {
	Response any                 `json:"response"`
	Vars     map[string]Multivar `json:"vars"`
}

func New(config *config.Config) (*Runner, error) {
	r := &Runner{
		config: config,
	}

	return r, nil
}
