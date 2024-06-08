package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cohix/ragoo/pkg/config"
	"github.com/cohix/ragoo/pkg/runner"
	"github.com/cohix/ragoo/pkg/server"
)

func main() {
	if len(os.Args) != 2 {
		slog.Error("missing argument: <config file path>")
		os.Exit(1)
	}

	slog.Info("--- Starting Ragoo --- ")

	config, err := config.ReadConfigFromFile(os.Args[1])
	if err != nil {
		slog.Error(fmt.Errorf("failed to ReadConfigFromFile: %w", err).Error())
		os.Exit(1)
	}

	if err := startImporters(config); err != nil {
		slog.Error(fmt.Errorf("failed to startImporters: %w", err).Error())
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

func startImporters(config *config.Config) error {
	runner, err := runner.New(config)
	if err != nil {
		return fmt.Errorf("failed to runner.New: %w", err)
	}

	for _, imp := range config.Importers {
		slog.Info("starting importer", "name", imp.Name)

		if err := runner.StartImporter(imp); err != nil {
			return fmt.Errorf("failed to StartImporter %s: %w", imp.Name, err)
		}
	}

	return nil
}
