package runner

import (
	"fmt"
	"strings"
)

// varSubst returns val if it is not a "variable" (i.e. starts with $), otherwise returns
// the value of the variable in vars with the matching name (sans $)
func varSubst(val string, vars map[string]Multivar) (*Multivar, error) {
	if !strings.HasPrefix(val, "$") {
		return &Multivar{String: val}, nil
	}

	varVal, exists := vars[strings.TrimPrefix(val, "$")]
	if !exists {
		return nil, fmt.Errorf("workflow vars missing value for key %s", val)
	}

	return &varVal, nil
}
