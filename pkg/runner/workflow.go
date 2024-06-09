package runner

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/cohix/ragoo/pkg/config"
)

const (
	inputKey    = "_input"
	responseKey = "_response"
	chunkKey    = "_chunk"
	refKey      = "_ref"
	batchKey    = "_batch"
)

// RunWorkflow runs the named workflow with the given params
func (r *Runner) RunWorkflow(ref string, params map[string]string) (*Result, error) {
	wrk := r.workflowFromConfig(ref)
	if wrk == nil {
		return nil, fmt.Errorf("workflow with ref %s not found", ref)
	}

	input, exists := params[inputKey]
	if !exists {
		return nil, fmt.Errorf("no input provided in workflow params")
	}

	// seed the vars with the "built in" _input var
	vars := map[string]Multivar{
		inputKey: {String: input, Bytes: []byte(input)},
	}

	if len(wrk.Stages) == 0 {
		return nil, fmt.Errorf("workflow with ref %s contains no stages", wrk.Name)
	}

	for _, stg := range wrk.Stages {
		if len(stg.Steps) == 0 {
			slog.Warn("workflow stage contains no steps, skipping", "name", stg.Name)
			continue
		}

		for _, stp := range stg.Steps {
			switch stp.Type {
			case "embedder":
				mult, key, err := r.runEmbedder(stp, vars)
				if err != nil {
					return nil, fmt.Errorf("failed to runEmbedder: %w", err)
				}

				vars[key] = *mult

			case "storage":
				mult, key, err := r.runStorage(stp, vars)
				if err != nil {
					return nil, fmt.Errorf("failed to runStorage: %w", err)
				}

				vars[key] = *mult

			case "service":
				mult, key, err := r.runService(stp, vars)
				if err != nil {
					return nil, fmt.Errorf("failed to runService: %w", err)
				}

				vars[key] = *mult

			case "importer":
				mult, key, err := r.runImporter(stp, vars)
				if err != nil {
					return nil, fmt.Errorf("failed to runImporter: %w", err)
				}

				vars[key] = *mult
			default:
				return nil, fmt.Errorf("workflow step has invalid type %s", stp.Type)
			}
		}
	}

	resp, exists := vars[responseKey]
	if !exists {
		return nil, errors.New("workflow did not produce a result (missing _result workflow var)")
	}

	res := &Result{
		Response: resp,
		Vars:     vars,
	}

	return res, nil
}

func (r *Runner) workflowFromConfig(ref string) *config.Workflow {
	for i, w := range r.config.Workflows {
		if w.Name == ref {
			return &r.config.Workflows[i]
		}
	}

	return nil
}
