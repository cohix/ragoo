package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/pgvector/pgvector-go"

	_ "github.com/marcboeker/go-duckdb"
)

type duckDBStorage struct {
	config  map[string]string
	db      *sql.DB
	created map[string]bool
}

func (d *duckDBStorage) InsertEmbedding(collection string, ref string, embedding []float32, batch string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn, err := d.ensureDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ensureDB: %w", err)
	}

	if _, exists := d.created[collection]; !exists {
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS collection_%s (embedding DOUBLE[], ref VARCHAR, batch VARCHAR);", collection)); err != nil {
			return nil, fmt.Errorf("failed to Exec: %w", err)
		}

		d.created[collection] = true
	}

	if _, err := conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO collection_%s (embedding, ref, batch) VALUES (?, ?, ?);", collection), pgvector.NewVector(embedding), ref, batch); err != nil {
		return nil, fmt.Errorf("failed to Exec: %w", err)
	}

	return &Result{}, nil
}

func (d *duckDBStorage) LookupCosine(collection string, embedding []float32, limit int, threshold float32) (*Result, error) {
	slog.Info("cosine lookup", "storage", "duckdb")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn, err := d.ensureDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ensureDB: %w", err)
	}

	res, err := conn.QueryContext(ctx, fmt.Sprintf(`
	SELECT ref, MAX(cosine) as max_cosine
		FROM(
				SELECT ref, list_cosine_similarity(embedding, ?) as cosine 
				FROM collection_%s 
				WHERE cosine > ? 
			)
		GROUP BY ref
		ORDER BY max_cosine DESC
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
func (d *duckDBStorage) Cleanup(collection string, batch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn, err := d.ensureDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensureDB: %w", err)
	}

	if _, err := conn.ExecContext(ctx, fmt.Sprintf("DELETE FROM collection_%s WHERE batch != ?;", collection), batch); err != nil {
		return fmt.Errorf("failed to Exec: %w", err)
	}

	return nil
}

func (d *duckDBStorage) ensureDB(ctx context.Context) (*sql.Conn, error) {
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

	conn, err := d.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to db.Conn: %w", err)
	}

	return conn, nil
}
