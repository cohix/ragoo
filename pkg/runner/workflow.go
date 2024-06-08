package runner

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/cohix/ragoo/pkg/config"
	"github.com/cohix/ragoo/pkg/embedder"
	"github.com/cohix/ragoo/pkg/service"
	"github.com/cohix/ragoo/pkg/storage"
	"github.com/cohix/ragoo/pkg/tool"
)

const (
	inputKey    = "_input"
	responseKey = "_response"
)

// Multivar represents one of several types of variables
type Multivar struct {
	String    string           `json:"string,omitempty"`
	Bytes     []byte           `json:"bytes,omitempty"`
	Any       any              `json:"obj,omitempty"`
	Embedding *embedder.Result `json:"embedding,omitempty"`
	Tool      *tool.Result     `json:"tool,omitempty"`
	Service   *service.Result  `json:"service,omitempty"`
	Storage   *storage.Result  `json:"storage,omitempty"`
}

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

func (r *Runner) runEmbedder(stp config.Step, vars map[string]Multivar) (*Multivar, string, error) {
	var mult *Multivar

	emb := r.embedder(stp.Ref)
	if emb == nil {
		return nil, "", fmt.Errorf("embedder with ref %s not found", stp.Ref)
	}

	switch stp.Action {
	case "generate":
		inputKey, exists := stp.Params["input"]
		if !exists {
			return nil, "", fmt.Errorf("embedder with ref %s missing param: input", stp.Ref)
		}

		input, err := varSubst(inputKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		res, err := emb.Generate(input.Bytes)
		if err != nil {
			return nil, "", fmt.Errorf("embedder with ref %s resulted in error: %w", stp.Ref, err)
		}

		mult = &Multivar{Embedding: res}
	default:
		return nil, "", fmt.Errorf("embedder with ref %s called with invalid action %s", stp.Ref, stp.Action)
	}

	key := "embedder"
	if stp.Var != "" {
		key = stp.Var
	}

	return mult, key, nil
}

func (r *Runner) runStorage(stp config.Step, vars map[string]Multivar) (*Multivar, string, error) {
	var mult *Multivar

	str := r.storage(stp.Ref)
	if str == nil {
		return nil, "", fmt.Errorf("storage with ref %s not found", stp.Ref)
	}

	switch stp.Action {
	case "lookup.vectorsimilarity":
		embeddingKey, exists := stp.Params["embedding"]
		if !exists {
			return nil, "", fmt.Errorf("embedder with ref %s missing param: embedding", stp.Ref)
		}

		embedding, err := varSubst(embeddingKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		collectionKey, exists := stp.Params["collection"]
		if !exists {
			return nil, "", fmt.Errorf("embedder with ref %s missing param: collection", stp.Ref)
		}

		collection, err := varSubst(collectionKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		limitKey, exists := stp.Params["limit"]
		if !exists {
			return nil, "", fmt.Errorf("embedder with ref %s missing param: limit", stp.Ref)
		}

		limit, err := varSubst(limitKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		limitInt, err := strconv.Atoi(limit.String)
		if err != nil {
			return nil, "", fmt.Errorf("failed to strconv.Atoi for param: limit (must be integer): %w", err)
		}

		res, err := str.VectorSimilarity(embedding.Embedding.Embedding, collection.String, limitInt)
		if err != nil {
			return nil, "", fmt.Errorf("embedder with ref %s resulted in error: %w", stp.Ref, err)
		}

		mult = &Multivar{Storage: res}
	default:
		return nil, "", fmt.Errorf("storage with ref %s called with invalid action %s", stp.Ref, stp.Action)
	}

	key := "storage"
	if stp.Var != "" {
		key = stp.Var
	}

	return mult, key, nil
}

func (r *Runner) runService(stp config.Step, vars map[string]Multivar) (*Multivar, string, error) {
	var mult *Multivar

	srv := r.service(stp.Ref)
	if srv == nil {
		return nil, "", fmt.Errorf("service with ref %s not found", stp.Ref)
	}

	switch stp.Action {
	case "completion":
		promptKey, exists := stp.Params["prompt"]
		if !exists {
			return nil, "", fmt.Errorf("embedder with ref %s missing param: input", stp.Ref)
		}

		prompt, err := varSubst(promptKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		res, err := srv.Completion(prompt.String)
		if err != nil {
			return nil, "", fmt.Errorf("embedder with ref %s resulted in error: %w", stp.Ref, err)
		}

		mult = &Multivar{Service: res}
	default:
		return nil, "", fmt.Errorf("service with ref %s called with invalid action %s", stp.Ref, stp.Action)
	}

	key := "service"
	if stp.Var != "" {
		key = stp.Var
	}

	return mult, key, nil
}
