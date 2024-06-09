package importer

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cohix/ragoo/pkg/storage"
	"github.com/jonathanhecl/chunker"
)

type fileImporter struct {
	config map[string]string
}

func (f *fileImporter) Run(batch string, resChan chan Result) error {
	dir, exists := f.config["directory"]
	if !exists {
		return errors.New("file importer missing config key: directory")
	}

	count := 0

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
		c := chunker.NewChunker(512, 24, chunker.DefaultSeparators, true, false)
		chunks := c.Chunk(string(fileBytes))

		r := Result{
			Ref:    path,
			Chunks: chunks,
			Batch:  batch,
		}

		resChan <- r

		count++

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to WalkDir: %w", err)
	}

	slog.Info("file import complete", "count", count)

	return nil
}

// ResolveRefs resolves the provided references to their contents
func (f *fileImporter) ResolveRefs(refs storage.Result) (*Result, error) {
	r := &Result{
		Documents: make([]string, len(refs.Refs)),
	}

	for i, ref := range refs.Refs {
		slog.Info("resolving document", "ref", ref, "cosine", refs.Cosines[i])

		fileBytes, err := os.ReadFile(filepath.Clean(ref))
		if err != nil {
			return nil, fmt.Errorf("failed to ReadFile: %w", err)
		}

		r.Documents[i] = string(fileBytes)
	}

	return r, nil
}
