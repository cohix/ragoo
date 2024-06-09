package runner

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/cohix/ragoo/pkg/config"
	"github.com/cohix/ragoo/pkg/embedder"
	"github.com/cohix/ragoo/pkg/importer"
	"github.com/cohix/ragoo/pkg/service"
	"github.com/cohix/ragoo/pkg/storage"
	"github.com/cohix/ragoo/pkg/tool"
)

const (
	inputKey    = "_input"
	responseKey = "_response"
	chunkKey    = "_chunk"
	refKey      = "_ref"
	batchKey    = "_batch"
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
	Importer  *importer.Result `json:"importer,omitempty"`
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

		res, err := emb.Generate(input.String)
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
	case "lookup.cosine":
		embeddingKey, exists := stp.Params["embedding"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: embedding", stp.Ref)
		}

		embedding, err := varSubst(embeddingKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		collectionKey, exists := stp.Params["collection"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: collection", stp.Ref)
		}

		collection, err := varSubst(collectionKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		limitKey, exists := stp.Params["limit"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: limit", stp.Ref)
		}

		limit, err := varSubst(limitKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		limitInt, err := strconv.Atoi(limit.String)
		if err != nil {
			return nil, "", fmt.Errorf("failed to strconv.Atoi for param: limit (must be integer): %w", err)
		}

		thresholdKey, exists := stp.Params["threshold"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: threshold", stp.Ref)
		}

		threshold, err := varSubst(thresholdKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		thresholdFloat, err := strconv.ParseFloat(threshold.String, 32)
		if err != nil {
			return nil, "", fmt.Errorf("failed to ParseFloat for param: threshold (must be decimal): %w", err)
		}

		res, err := str.LookupCosine(embedding.Embedding.Embedding, collection.String, limitInt, float32(thresholdFloat))
		if err != nil {
			return nil, "", fmt.Errorf("storage with ref %s resulted in error: %w", stp.Ref, err)
		}

		mult = &Multivar{Storage: res}
	case "insert.embedding":
		embeddingKey, exists := stp.Params["embedding"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: embedding", stp.Ref)
		}

		embedding, err := varSubst(embeddingKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		collectionKey, exists := stp.Params["collection"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: collection", stp.Ref)
		}

		collection, err := varSubst(collectionKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		refKey, exists := stp.Params["ref"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: limit", stp.Ref)
		}

		ref, err := varSubst(refKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		batchKey, exists := stp.Params["batch"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: batch", stp.Ref)
		}

		batch, err := varSubst(batchKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		res, err := str.InsertEmbedding(embedding.Embedding.Embedding, collection.String, ref.String, batch.String)
		if err != nil {
			return nil, "", fmt.Errorf("storage with ref %s resulted in error: %w", stp.Ref, err)
		}

		mult = &Multivar{Storage: res}
	case "cleanup":
		batchKey, exists := stp.Params["batch"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: batch", stp.Ref)
		}

		batch, err := varSubst(batchKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		collectionKey, exists := stp.Params["collection"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: collection", stp.Ref)
		}

		collection, err := varSubst(collectionKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		if err := str.Cleanup(collection.String, batch.String); err != nil {
			return nil, "", fmt.Errorf("failed to Cleanup: %w", err)
		}
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

		augmented, err := promptSubst(prompt.String, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to promptSubst: %w", err)
		}

		res, err := srv.Completion(augmented)
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

func (r *Runner) runImporter(stp config.Step, vars map[string]Multivar) (*Multivar, string, error) {
	var mult *Multivar

	imp := r.importer(stp.Ref)
	if imp == nil {
		return nil, "", fmt.Errorf("importer with ref %s not found", stp.Ref)
	}

	switch stp.Action {
	case "resolve.refs":
		refsKey, exists := stp.Params["refs"]
		if !exists {
			return nil, "", fmt.Errorf("importer with ref %s missing param: input", stp.Ref)
		}

		refs, err := varSubst(refsKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		res, err := imp.ResolveRefs(*refs.Storage)
		if err != nil {
			return nil, "", fmt.Errorf("importer with ref %s resulted in error: %w", stp.Ref, err)
		}

		seperator := " "
		sep, exists := stp.Params["seperator"]
		if exists {
			seperator = sep
		}

		combined := strings.Join(res.Documents, seperator)

		mult = &Multivar{Importer: res, String: combined, Bytes: []byte(combined)}
	default:
		return nil, "", fmt.Errorf("importer with ref %s called with invalid action %s", stp.Ref, stp.Action)
	}

	key := "importer"
	if stp.Var != "" {
		key = stp.Var
	}

	return mult, key, nil
}
