package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/cohix/ragoo/pkg/config"
	"github.com/cohix/ragoo/pkg/runner"
)

func (s *Server) handlerForRoute(route config.Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rn, err := runner.New(s.config)
		if err != nil {
			slog.Error(fmt.Errorf("failed to runner.New: %w", err).Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer func() {
			if err := r.Body.Close(); err != nil {
				slog.Error(fmt.Errorf("failed to Body.Close: %w", err).Error())
			}
		}()

		inputBuf, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error(fmt.Errorf("failed to ReadAll: %w", err).Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		params := route.Workflow.Params
		if params == nil {
			params = map[string]string{}
		}

		params["_input"] = string(inputBuf)

		result, err := rn.RunWorkflow(route.Workflow.Ref, params)
		if err != nil {
			slog.Error(fmt.Errorf("failed to RunWorkflow: %w", err).Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		slog.Info("workflow completed", "name", route.Workflow.Ref)

		if err := json.NewEncoder(w).Encode(result.Response); err != nil {
			slog.Error(fmt.Errorf("failed to NewEncoder.Encode: %w", err).Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
