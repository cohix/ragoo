package server

import (
	"fmt"
	"net/http"

	"github.com/cohix/ragoo/pkg/config"
)

type Server struct {
	config *config.Config
}

// New returns a new server
func New(config *config.Config) (*Server, error) {
	s := &Server{
		config: config,
	}

	return s, nil
}

// Start starts the server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	for _, r := range s.config.Routes {
		mux.HandleFunc(r.Path, handlerForRoute(r))
	}

	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to ListenAndServe: %w", err)
	}

	return nil
}
