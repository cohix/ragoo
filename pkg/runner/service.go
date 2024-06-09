package runner

import (
	"fmt"

	"github.com/cohix/ragoo/pkg/config"
	"github.com/cohix/ragoo/pkg/service"
)

func (r *Runner) runService(stp config.Step, vars map[string]Multivar) (*Multivar, string, error) {
	var mult *Multivar

	srv := r.service(stp.Ref)
	if srv == nil {
		return nil, "", fmt.Errorf("service with ref %s not found", stp.Ref)
	}

	switch stp.Action {
	case "completion":
		prompt, err := resolveParam("prompt", stp.Params, vars, false)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolveParam 'prompt' for service: %w", err)
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

func (r *Runner) service(ref string) service.Service {
	for _, srv := range r.config.Services {
		if srv.Name == ref {
			return service.ServiceOfType(srv.Type, srv.Config)
		}
	}

	return nil
}
