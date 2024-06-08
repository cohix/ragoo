package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pgvector/pgvector-go"

	_ "github.com/marcboeker/go-duckdb"
)

type duckDBStorage struct {
	config  map[string]string
	db      *sql.DB
	created map[string]bool
}

func (d *duckDBStorage) InsertEmbedding(embedding []float32, collection string, ref string) (*Result, error) {
	db, err := d.ensureDB()
	if err != nil {
		return nil, fmt.Errorf("failed to ensureDB: %w", err)
	}

	if _, exists := d.created[collection]; !exists {
		if _, err := db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS collection_%s (embedding DOUBLE[], ref varchar);", collection)); err != nil {
			return nil, fmt.Errorf("failed to Exec: %w", err)
		}

		d.created[collection] = true
	}

	if _, err := db.Exec(fmt.Sprintf("INSERT INTO collection_%s (embedding, ref) VALUES (?, ?)", collection), pgvector.NewVector(embedding), ref); err != nil {
		return nil, fmt.Errorf("failed to Exec: %w", err)
	}

	return &Result{}, nil
}

func (d *duckDBStorage) LookupCosine(embedding []float32, collection string, limit int, threshold float32) (*Result, error) {
	db, err := d.ensureDB()
	if err != nil {
		return nil, fmt.Errorf("failed to ensureDB: %w", err)
	}

	res, err := db.Query(fmt.Sprintf("SELECT DISTINCT ref FROM collection_%s c WHERE list_cosine_similarity(c.embedding, ?) > ? LIMIT ?;", collection), pgvector.NewVector(embedding), threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to Exec: %w", err)
	}

	defer res.Close()

	result := &Result{
		Refs: []string{},
	}

	for res.Next() {
		var ref string
		if err := res.Scan(&ref); err != nil {
			return nil, fmt.Errorf("failed to res.Scan: %w", err)
		}

		result.Refs = append(result.Refs, ref)
	}

	return result, nil
}

func (d *duckDBStorage) ensureDB() (*sql.DB, error) {
	if d.db == nil {
		dbFile, exists := d.config["dbFilePath"]
		if !exists {
			return nil, errors.New("storage of type duckdb missing config key: dbFilePath")
		}

		if err := os.MkdirAll(filepath.Dir(filepath.Clean(dbFile)), os.FileMode(0o700)); err != nil {
			return nil, fmt.Errorf("failed to MkdirAll: %w", err)
		}

		db, err := sql.Open("duckdb", filepath.Clean(dbFile))
		if err != nil {
			return nil, fmt.Errorf("failed to sql.Open: %w", err)
		}

		d.db = db
	}

	return d.db, nil
}
