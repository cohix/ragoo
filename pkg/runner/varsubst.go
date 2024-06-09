package runner

import (
	"fmt"
	"strings"
)

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
