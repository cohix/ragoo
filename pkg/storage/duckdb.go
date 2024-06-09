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

func (d *duckDBStorage) InsertEmbedding(embedding []float32, collection string, ref string, batch string) (*Result, error) {
	db, err := d.ensureDB()
	if err != nil {
		return nil, fmt.Errorf("failed to ensureDB: %w", err)
	}

	if _, exists := d.created[collection]; !exists {
		if _, err := db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS collection_%s (embedding DOUBLE[], ref VARCHAR, batch VARCHAR);", collection)); err != nil {
			return nil, fmt.Errorf("failed to Exec: %w", err)
		}

		d.created[collection] = true
	}

	if _, err := db.Exec(fmt.Sprintf("INSERT INTO collection_%s (embedding, ref, batch) VALUES (?, ?, ?);", collection), pgvector.NewVector(embedding), ref, batch); err != nil {
		return nil, fmt.Errorf("failed to Exec: %w", err)
	}

	return &Result{}, nil
}

func (d *duckDBStorage) LookupCosine(embedding []float32, collection string, limit int, threshold float32) (*Result, error) {
	db, err := d.ensureDB()
	if err != nil {
		return nil, fmt.Errorf("failed to ensureDB: %w", err)
	}

	res, err := db.Query(fmt.Sprintf(`
	SELECT ref, MAX(cosine)
		FROM(
				SELECT ref, list_cosine_similarity(embedding, ?) as cosine 
				FROM collection_%s 
				WHERE cosine > ? 
				ORDER BY cosine DESC)
		GROUP BY ref		
		LIMIT ?;`, collection), pgvector.NewVector(embedding), threshold, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to Exec: %w", err)
	}

	defer res.Close()

	result := &Result{
		Refs:    []string{},
		Cosines: []float32{},
	}

	for res.Next() {
		var ref string
		var cosine float32
		if err := res.Scan(&ref, &cosine); err != nil {
			return nil, fmt.Errorf("failed to res.Scan: %w", err)
		}

		result.Refs = append(result.Refs, ref)
		result.Cosines = append(result.Cosines, cosine)
	}

	return result, nil
}

// Cleanup cleans up old data
func (d *duckDBStorage) Cleanup(collection string, currentBatch string) error {
	db, err := d.ensureDB()
	if err != nil {
		return fmt.Errorf("failed to ensureDB: %w", err)
	}

	if _, err := db.Exec(fmt.Sprintf("DELETE FROM collection_%s WHERE batch != ?;", collection), currentBatch); err != nil {
		return fmt.Errorf("failed to Exec: %w", err)
	}

	return nil
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
