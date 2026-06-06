package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// BenchmarkRecord represents a single benchmark entry.
type BenchmarkRecord struct {
	Task       string    `json:"task"`
	Model      string    `json:"model"`
	TokensIn   int       `json:"tokens_in"`
	TokensOut  int       `json:"tokens_out"`
	LatencyMs  int64     `json:"latency_ms"`
	Cost       float64   `json:"cost"`
	Provider   string    `json:"provider"`
	Timestamp  time.Time `json:"timestamp"`
}

// Store handles persistence of benchmark records and other storage needs.
type Store struct {
	db *sql.DB
}

// NewStore creates a new storage store with SQLite.
func NewStore(dataDir string) (*Store, error) {
	// Ensure directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "storage.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// createTables creates the necessary tables if they don't exist.
func createTables(db *sql.DB) error {
	// Benchmark records table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS benchmark_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task TEXT NOT NULL,
			model TEXT NOT NULL,
			tokens_in INTEGER NOT NULL,
			tokens_out INTEGER NOT NULL,
			latency_ms INTEGER NOT NULL,
			cost REAL NOT NULL,
			provider TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// SaveBenchmarkRecord saves a benchmark record.
func (s *Store) SaveBenchmarkRecord(ctx context.Context, record *BenchmarkRecord) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO benchmark_records (task, model, tokens_in, tokens_out, latency_ms, cost, provider, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, record.Task, record.Model, record.TokensIn, record.TokensOut, record.LatencyMs, record.Cost, record.Provider, record.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to save benchmark record: %w", err)
	}
	return nil
}

// GetBenchmarkRecords retrieves benchmark records with optional filtering.
func (s *Store) GetBenchmarkRecords(ctx context.Context, limit int) ([]*BenchmarkRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT task, model, tokens_in, tokens_out, latency_ms, cost, provider, timestamp
		FROM benchmark_records
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get benchmark records: %w", err)
	}
	defer rows.Close()

	var records []*BenchmarkRecord
	for rows.Next() {
		var record BenchmarkRecord
		if err := rows.Scan(&record.Task, &record.Model, &record.TokensIn, &record.TokensOut, &record.LatencyMs, &record.Cost, &record.Provider, &record.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan benchmark record: %w", err)
		}
		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating benchmark records: %w", err)
	}

	return records, nil
}