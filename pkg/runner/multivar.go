package runner

import (
	"fmt"
	"strings"

	"github.com/cohix/ragoo/pkg/embedder"
	"github.com/cohix/ragoo/pkg/importer"
	"github.com/cohix/ragoo/pkg/service"
	"github.com/cohix/ragoo/pkg/storage"
	"github.com/cohix/ragoo/pkg/tool"
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

func resolveParam(key string, params map[string]string, vars map[string]Multivar, optional bool) (*Multivar, error) {
	val, exists := params[key]
	if !exists && !optional {
		return nil, fmt.Errorf("param not found: %s", key)
	} else if !exists && optional {
		return nil, nil
	}

	multi, err := varSubst(val, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to varSubst for params key: %s, val: %s: %w", key, val, err)
	}

	return multi, nil
}

// varSubst returns val if it is not a "variable" (i.e. starts with $), otherwise returns
// the value of the variable in vars with the matching name (sans $)
func varSubst(val string, vars map[string]Multivar) (*Multivar, error) {
	if !strings.HasPrefix(val, "$") || strings.Count(val, "$") > 1 {
		return &Multivar{String: val}, nil
	}

	varVal, exists := vars[strings.TrimPrefix(val, "$")]
	if !exists {
		return nil, fmt.Errorf("workflow vars missing value for key %s", val)
	}

	return &varVal, nil
}

func promptSubst(prompt string, vars map[string]Multivar) (string, error) {
	for k, mv := range vars {
		key := fmt.Sprintf("$%s", k)

		val := mv.String
		if val == "" {
			val = string(mv.Bytes)
			if val == "" {
				continue
			}
		}

		prompt = strings.Replace(prompt, key, val, -1)
	}

	return prompt, nil
}
