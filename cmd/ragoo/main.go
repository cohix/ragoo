package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cohix/ragoo/pkg/config"
	"github.com/cohix/ragoo/pkg/server"
)

func main() {
	if len(os.Args) != 2 {
		slog.Error("missing argument: <config file path>")
		os.Exit(1)
	}

	config, err := config.ReadConfigFromFile(os.Args[1])
	if err != nil {
		slog.Error(fmt.Errorf("failed to ReadConfigFromFile: %w", err).Error())
		os.Exit(1)
	}

	srv, err := server.New(config)
	if err != nil {
		slog.Error(fmt.Errorf("failed to server.New: %w", err).Error())
		os.Exit(1)
	}

	if err := srv.Start(); err != nil {
		slog.Error(fmt.Errorf("failed to srv.Start: %w", err).Error())
		os.Exit(1)
	}
}
