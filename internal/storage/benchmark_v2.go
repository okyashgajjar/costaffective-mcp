package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type RetrievalBenchmark struct {
	ID            string    `json:"id"`
	TaskID        string    `json:"task_id"`
	Repository    string    `json:"repository"`
	Retriever     string    `json:"retriever"`
	Query         string    `json:"query"`
	FilesScanned  int       `json:"files_scanned"`
	FilesLoaded   int       `json:"files_loaded"`
	TokensContext int       `json:"tokens_context"`
	TokensInput   int       `json:"tokens_input"`
	TokensOutput  int       `json:"tokens_output"`
	LatencyMs     int64     `json:"latency_ms"`
	Provider      string    `json:"provider"`
	Model         string    `json:"model"`
	Cost          float64   `json:"cost"`
	ContextLevel  string    `json:"context_level"`
	Timestamp     time.Time `json:"timestamp"`
}

type BenchmarkStore struct {
	db *sql.DB
}

func NewBenchmarkStore(dataDir string) (*BenchmarkStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "benchmarks.db")
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open benchmark database: %w", err)
	}

	if err := createBenchmarkTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create benchmark tables: %w", err)
	}

	return &BenchmarkStore{db: db}, nil
}

func createBenchmarkTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS retrieval_benchmarks (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL DEFAULT '',
			repository TEXT NOT NULL DEFAULT '',
			retriever TEXT NOT NULL DEFAULT '',
			query TEXT NOT NULL DEFAULT '',
			files_scanned INTEGER NOT NULL DEFAULT 0,
			files_loaded INTEGER NOT NULL DEFAULT 0,
			tokens_context INTEGER NOT NULL DEFAULT 0,
			tokens_input INTEGER NOT NULL DEFAULT 0,
			tokens_output INTEGER NOT NULL DEFAULT 0,
			latency_ms INTEGER NOT NULL DEFAULT 0,
			provider TEXT NOT NULL DEFAULT '',
			model TEXT NOT NULL DEFAULT '',
			cost REAL NOT NULL DEFAULT 0.0,
			context_level TEXT NOT NULL DEFAULT 'snippets',
			timestamp TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_benchmark_retriever
		ON retrieval_benchmarks(retriever)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_benchmark_task
		ON retrieval_benchmarks(task_id)
	`)
	return err
}

func (bs *BenchmarkStore) Close() error {
	return bs.db.Close()
}

func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	ts := time.Now().UnixMilli()
	return fmt.Sprintf("bm_%d_%s", ts, hex.EncodeToString(bytes))
}

func (bs *BenchmarkStore) Save(ctx context.Context, b *RetrievalBenchmark) error {
	if b.ID == "" {
		b.ID = generateID()
	}
	if b.Timestamp.IsZero() {
		b.Timestamp = time.Now()
	}

	_, err := bs.db.ExecContext(ctx, `
		INSERT INTO retrieval_benchmarks (
			id, task_id, repository, retriever, query,
			files_scanned, files_loaded, tokens_context,
			tokens_input, tokens_output, latency_ms,
			provider, model, cost, context_level, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		b.ID, b.TaskID, b.Repository, b.Retriever, b.Query,
		b.FilesScanned, b.FilesLoaded, b.TokensContext,
		b.TokensInput, b.TokensOutput, b.LatencyMs,
		b.Provider, b.Model, b.Cost, b.ContextLevel, b.Timestamp,
	)
	return err
}

func (bs *BenchmarkStore) List(ctx context.Context, limit int) ([]*RetrievalBenchmark, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := bs.db.QueryContext(ctx, `
		SELECT id, task_id, repository, retriever, query,
			files_scanned, files_loaded, tokens_context,
			tokens_input, tokens_output, latency_ms,
			provider, model, cost, context_level, timestamp
		FROM retrieval_benchmarks
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*RetrievalBenchmark
	for rows.Next() {
		rec := &RetrievalBenchmark{}
		if err := rows.Scan(
			&rec.ID, &rec.TaskID, &rec.Repository, &rec.Retriever, &rec.Query,
			&rec.FilesScanned, &rec.FilesLoaded, &rec.TokensContext,
			&rec.TokensInput, &rec.TokensOutput, &rec.LatencyMs,
			&rec.Provider, &rec.Model, &rec.Cost, &rec.ContextLevel, &rec.Timestamp,
		); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (bs *BenchmarkStore) GetByID(ctx context.Context, id string) (*RetrievalBenchmark, error) {
	rec := &RetrievalBenchmark{}
	err := bs.db.QueryRowContext(ctx, `
		SELECT id, task_id, repository, retriever, query,
			files_scanned, files_loaded, tokens_context,
			tokens_input, tokens_output, latency_ms,
			provider, model, cost, context_level, timestamp
		FROM retrieval_benchmarks
		WHERE id = ?
	`, id).Scan(
		&rec.ID, &rec.TaskID, &rec.Repository, &rec.Retriever, &rec.Query,
		&rec.FilesScanned, &rec.FilesLoaded, &rec.TokensContext,
		&rec.TokensInput, &rec.TokensOutput, &rec.LatencyMs,
		&rec.Provider, &rec.Model, &rec.Cost, &rec.ContextLevel, &rec.Timestamp,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("benchmark not found: %s", id)
	}
	return rec, err
}

func (bs *BenchmarkStore) Export(ctx context.Context) ([]*RetrievalBenchmark, error) {
	return bs.List(ctx, 10000)
}

func (bs *BenchmarkStore) GetByTaskID(ctx context.Context, taskID string) ([]*RetrievalBenchmark, error) {
	rows, err := bs.db.QueryContext(ctx, `
		SELECT id, task_id, repository, retriever, query,
			files_scanned, files_loaded, tokens_context,
			tokens_input, tokens_output, latency_ms,
			provider, model, cost, context_level, timestamp
		FROM retrieval_benchmarks
		WHERE task_id = ?
		ORDER BY retriever, timestamp
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*RetrievalBenchmark
	for rows.Next() {
		rec := &RetrievalBenchmark{}
		if err := rows.Scan(
			&rec.ID, &rec.TaskID, &rec.Repository, &rec.Retriever, &rec.Query,
			&rec.FilesScanned, &rec.FilesLoaded, &rec.TokensContext,
			&rec.TokensInput, &rec.TokensOutput, &rec.LatencyMs,
			&rec.Provider, &rec.Model, &rec.Cost, &rec.ContextLevel, &rec.Timestamp,
		); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}
