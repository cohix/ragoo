package importer

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jonathanhecl/chunker"
)

type fileImporter struct {
	config map[string]string
}

func (f *fileImporter) Run(resChan chan Result) error {
	dir, exists := f.config["directory"]
	if !exists {
		return errors.New("file importer missing config key: directory")
	}

	err := filepath.WalkDir(filepath.Clean(dir), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error(fmt.Errorf("error from WalkDir: %w", err).Error())
			return nil
		}

		if d.IsDir() {
			return nil
		}

		slog.Info("importing file", "path", path)

		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to ReadFile: %w", err)
		}

		// TODO: this is a random chunker found with a quick Google search, may need to revisit using it.
		// another option is the package from langchain-go: https://pkg.go.dev/github.com/tmc/langchaingo/textsplitter
		c := chunker.NewChunker(265, 24, chunker.DefaultSeparators, true, false)
		chunks := c.Chunk(string(fileBytes))

		r := Result{
			Ref:    path,
			Chunks: chunks,
		}

		resChan <- r

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to WalkDir: %w", err)
	}

	return nil
}
