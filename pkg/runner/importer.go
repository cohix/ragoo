package runner

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/cohix/ragoo/pkg/config"
	"github.com/cohix/ragoo/pkg/importer"
)

// StartImporter starts the provided importer on a goroutine
func (r *Runner) StartImporter(imp config.Importer) error {
	im := importer.ImporterOfType(imp.Type, imp.Config)
	if im == nil {
		return fmt.Errorf("importer of type %s not found", imp.Type)
	}

	resultChan := make(chan importer.Result, 1)

	// first of two goroutines to catch any results generated by the importer
	// and run the defined steps on each chunk
	go func() {
		for res := range resultChan {
			for _, ch := range res.Chunks {
				vars := map[string]Multivar{
					chunkKey: {String: ch},
					refKey:   {String: res.Ref},
					batchKey: {String: res.Batch},
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

	// second of two goroutines to run the importer and its defined
	// cleanup steps on a regular schedule
	go func() {
		for {
			batchID, err := batchID()
			if err != nil {
				slog.Error(fmt.Errorf("failed to batchID: %w", err).Error())
				time.Sleep(time.Minute * 5)
				continue
			}

			if err := im.Run(batchID, resultChan); err != nil {
				slog.Error(fmt.Errorf("failed to Run importer %s: %w", imp.Name, err).Error())
				time.Sleep(time.Minute * 5)
				continue
			} else {
				slog.Info("ran importer successfully", "name", imp.Name)
			}

			vars := map[string]Multivar{
				batchKey: {String: batchID},
			}

			switch imp.Cleanup.Type {
			case "storage":
				_, _, err := r.runStorage(imp.Cleanup, vars)
				if err != nil {
					slog.Error(fmt.Errorf("failed to runStorage: %w", err).Error())
				} else {
					slog.Info("ran importer cleanup successfully", "name", imp.Name)
				}
			default:
				slog.Error(fmt.Errorf("encountered cleanup with unsupported type: %s", imp.Cleanup.Type).Error())
			}

			time.Sleep(time.Minute * 5)
		}
	}()

	return nil
}

func batchID() (string, error) {
	reader := io.LimitReader(rand.Reader, 24)
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to ReadAll: %w", err)
	}

	randStr := base64.StdEncoding.EncodeToString(bytes)

	return randStr, nil
}
