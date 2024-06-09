package runner

import (
	"fmt"
	"strconv"

	"github.com/cohix/ragoo/pkg/config"
	"github.com/cohix/ragoo/pkg/storage"
)

func (r *Runner) runStorage(stp config.Step, vars map[string]Multivar) (*Multivar, string, error) {
	var mult *Multivar

	str := r.storage(stp.Ref)
	if str == nil {
		return nil, "", fmt.Errorf("storage with ref %s not found", stp.Ref)
	}

	switch stp.Action {
	case "lookup.cosine":
		embedding, err := resolveParam("embedding", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'embedding' for storage: %w", err)
		}

		collection, err := resolveParam("collection", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'collection' for storage: %w", err)
		}

		limit, err := resolveParam("limit", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'limit' for storage: %w", err)
		}

		limitInt, err := strconv.Atoi(limit.String)
		if err != nil {
			return nil, "", fmt.Errorf("failed to strconv.Atoi for param 'limit' (must be integer): %w", err)
		}

		threshold, err := resolveParam("threshold", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'threshold' for storage: %w", err)
		}

		thresholdFloat, err := strconv.ParseFloat(threshold.String, 32)
		if err != nil {
			return nil, "", fmt.Errorf("failed to ParseFloat for param: threshold (must be decimal): %w", err)
		}

		res, err := str.LookupCosine(collection.String, embedding.Embedding.Embedding, limitInt, float32(thresholdFloat))
		if err != nil {
			return nil, "", fmt.Errorf("storage with ref %s resulted in error: %w", stp.Ref, err)
		}

		mult = &Multivar{Storage: res}
	case "insert.embedding":
		embedding, err := resolveParam("embedding", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'embedding' for storage: %w", err)
		}

		collection, err := resolveParam("collection", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'collection' for storage: %w", err)
		}

		ref, err := resolveParam("ref", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'ref' for storage: %w", err)
		}

		batchKey, exists := stp.Params["batch"]
		if !exists {
			return nil, "", fmt.Errorf("storage with ref %s missing param: batch", stp.Ref)
		}

		batch, err := varSubst(batchKey, vars)
		if err != nil {
			return nil, "", fmt.Errorf("failed to varSubst: %w", err)
		}

		res, err := str.InsertEmbedding(collection.String, ref.String, embedding.Embedding.Embedding, batch.String)
		if err != nil {
			return nil, "", fmt.Errorf("storage with ref %s resulted in error: %w", stp.Ref, err)
		}

		mult = &Multivar{Storage: res}
	case "cleanup":
		batch, err := resolveParam("batch", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'batch' for storage: %w", err)
		}

		collection, err := resolveParam("collection", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'collection' for storage: %w", err)
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

// storage instances are persistent and are reused, unlike other object types (for the time being)
func (r *Runner) storage(ref string) storage.Storage {
	for _, str := range r.config.Storage {
		if str.Name == ref {
			return storage.StorageOfType(str.Name, str.Type, str.Config)
		}
	}

	return nil
}
