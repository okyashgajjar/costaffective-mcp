package cache

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresBackend implements StorageBackend using PostgreSQL
type PostgresBackend struct {
	db *sql.DB
}

// NewPostgresBackend creates a new Postgres cache backend
func NewPostgresBackend(connStr string) (*PostgresBackend, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres DB: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping postgres DB: %w", err)
	}

	if err := createPostgresTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &PostgresBackend{db: db}, nil
}

func createPostgresTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cache_entries (
			key_hash TEXT PRIMARY KEY,
			repo_hash TEXT NOT NULL,
			query TEXT NOT NULL,
			retriever TEXT NOT NULL,
			context_level TEXT NOT NULL,
			token_budget INTEGER NOT NULL,
			results_json BYTEA,
			context_text TEXT,
			tokens INTEGER,
			created_at TIMESTAMP,
			last_accessed TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_cache_repo ON cache_entries(repo_hash);
	`)
	return err
}

func (p *PostgresBackend) Get(keyHash string) (*CacheEntry, error) {
	row := p.db.QueryRow(`
		SELECT results_json, context_text, tokens, created_at 
		FROM cache_entries WHERE key_hash = $1`, keyHash)

	var entry CacheEntry
	if err := row.Scan(&entry.ResultsJSON, &entry.Context, &entry.Tokens, &entry.CreatedAt); err != nil {
		return nil, err
	}
	_, _ = p.db.Exec("UPDATE cache_entries SET last_accessed = $1 WHERE key_hash = $2", time.Now(), keyHash)
	return &entry, nil
}

func (p *PostgresBackend) Put(keyHash string, key CacheKey, entry *CacheEntry, maxSize int) error {
	now := time.Now()
	// Postgres upsert syntax
	_, err := p.db.Exec(`
		INSERT INTO cache_entries
			(key_hash, repo_hash, query, retriever, context_level, token_budget,
			 results_json, context_text, tokens, created_at, last_accessed)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (key_hash) DO UPDATE SET
			repo_hash = EXCLUDED.repo_hash,
			query = EXCLUDED.query,
			retriever = EXCLUDED.retriever,
			context_level = EXCLUDED.context_level,
			token_budget = EXCLUDED.token_budget,
			results_json = EXCLUDED.results_json,
			context_text = EXCLUDED.context_text,
			tokens = EXCLUDED.tokens,
			created_at = EXCLUDED.created_at,
			last_accessed = EXCLUDED.last_accessed
	`, keyHash, key.RepoHash, key.Query, key.Retriever, key.ContextLevel, key.TokenBudget,
		entry.ResultsJSON, entry.Context, entry.Tokens, entry.CreatedAt, now)
	if err != nil {
		return err
	}

	// Postgres-based eviction
	_, err = p.db.Exec(`
		DELETE FROM cache_entries WHERE key_hash NOT IN (
			SELECT key_hash FROM cache_entries ORDER BY last_accessed DESC LIMIT $1
		)
	`, maxSize)
	return err
}

func (p *PostgresBackend) Invalidate(repoHash string) error {
	_, err := p.db.Exec("DELETE FROM cache_entries WHERE repo_hash = $1", repoHash)
	return err
}

func (p *PostgresBackend) Size() (int, error) {
	var size int
	err := p.db.QueryRow("SELECT COUNT(*) FROM cache_entries").Scan(&size)
	return size, err
}

func (p *PostgresBackend) Close() error {
	return p.db.Close()
}
