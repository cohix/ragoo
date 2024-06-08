package runner

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/cohix/ragoo/pkg/config"
	"github.com/cohix/ragoo/pkg/importer"
)

// StartImporter starts the provided importer on a goroutine
func (r *Runner) StartImporter(imp config.Importer) error {
	im := importer.ImporterForType(imp.Type, imp.Config)
	if im == nil {
		return fmt.Errorf("importer of type %s not found", imp.Type)
	}

	resultChan := make(chan importer.Result, 1)

	go func() {
		for res := range resultChan {
			for _, ch := range res.Chunks {
				vars := map[string]Multivar{
					chunkKey: {String: ch},
					refKey:   {String: res.Ref},
				}

				for _, stp := range imp.Steps {
					switch stp.Type {
					case "embedder":
						mult, key, err := r.runEmbedder(stp, vars)
						if err != nil {
							slog.Error(fmt.Errorf("failed to runEmbedder: %w", err).Error())
							return
						}

						vars[key] = *mult

					case "storage":
						mult, key, err := r.runStorage(stp, vars)
						if err != nil {
							slog.Error(fmt.Errorf("failed to runStorage: %w", err).Error())
							return
						}

						vars[key] = *mult
					}
				}
			}
		}
	}()

	go func() {
		for {
			if err := im.Run(resultChan); err != nil {
				slog.Error(fmt.Errorf("failed to Run importer %s: %w", imp.Name, err).Error())
			} else {
				slog.Info("ran importer successfully", "name", imp.Name)
			}

			time.Sleep(time.Minute)
		}
	}()

	return nil
}
