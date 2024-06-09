package runner

import (
	"fmt"

	"github.com/cohix/ragoo/pkg/config"
	"github.com/cohix/ragoo/pkg/embedder"
)

func (r *Runner) runEmbedder(stp config.Step, vars map[string]Multivar) (*Multivar, string, error) {
	var mult *Multivar

	emb := r.embedder(stp.Ref)
	if emb == nil {
		return nil, "", fmt.Errorf("embedder with ref %s not found", stp.Ref)
	}

	switch stp.Action {
	case "generate":
		input, err := resolveParam("input", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'input' for embedder: %w", err)
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

func (r *Runner) embedder(ref string) embedder.Embedder {
	for _, emb := range r.config.Embedders {
		if emb.Name == ref {
			return embedder.EmbedderOfType(emb.Type, emb.Config)
		}
	}

	return nil
}
