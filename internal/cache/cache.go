package cache

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// CacheKey identifies a unique retrieval+context combination.
type CacheKey struct {
	RepoHash     string `json:"repo_hash"`
	Query        string `json:"query"`
	Retriever    string `json:"retriever"`
	ContextLevel string `json:"context_level"`
	TokenBudget  int    `json:"token_budget"`
}

// CacheEntry holds the cached retrieval and context results.
type CacheEntry struct {
	ResultsJSON []byte    `json:"results_json"`
	Context     string    `json:"context"`
	Tokens      int       `json:"tokens"`
	CreatedAt   time.Time `json:"created_at"`
}

// CacheStats holds cache performance metrics.
type CacheStats struct {
	Hits    int64   `json:"hits"`
	Misses  int64   `json:"misses"`
	HitRate float64 `json:"hit_rate"`
	Size    int     `json:"size"`
	MaxSize int     `json:"max_size"`
}

type StorageBackend interface {
	Get(keyHash string) (*CacheEntry, error)
	Put(keyHash string, key CacheKey, entry *CacheEntry, maxSize int) error
	Invalidate(repoHash string) error
	Size() (int, error)
	Close() error
}

// Cache is an LRU cache persistent (SQLite or Postgres) layers.
type Cache struct {
	maxSize int
	backend StorageBackend
	hits    int64
	misses  int64
}

// SQLiteBackend implements StorageBackend using SQLite
type SQLiteBackend struct {
	db *sql.DB
}

func NewSQLiteBackend(dbPath string) (*SQLiteBackend, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite DB: %w", err)
	}

	if err := createCacheTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteBackend{db: db}, nil
}

func (s *SQLiteBackend) Get(keyHash string) (*CacheEntry, error) {
	row := s.db.QueryRow(`
		SELECT results_json, context_text, tokens, created_at 
		FROM cache_entries WHERE key_hash = ?`, keyHash)

	var entry CacheEntry
	if err := row.Scan(&entry.ResultsJSON, &entry.Context, &entry.Tokens, &entry.CreatedAt); err != nil {
		return nil, err
	}
	_, _ = s.db.Exec("UPDATE cache_entries SET last_accessed = ? WHERE key_hash = ?", time.Now(), keyHash)
	return &entry, nil
}

func (s *SQLiteBackend) Put(keyHash string, key CacheKey, entry *CacheEntry, maxSize int) error {
	now := time.Now()
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO cache_entries
			(key_hash, repo_hash, query, retriever, context_level, token_budget,
			 results_json, context_text, tokens, created_at, last_accessed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, keyHash, key.RepoHash, key.Query, key.Retriever, key.ContextLevel, key.TokenBudget,
		entry.ResultsJSON, entry.Context, entry.Tokens, entry.CreatedAt, now)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		DELETE FROM cache_entries WHERE key_hash NOT IN (
			SELECT key_hash FROM cache_entries ORDER BY last_accessed DESC LIMIT ?
		)
	`, maxSize)
	return err
}

func (s *SQLiteBackend) Invalidate(repoHash string) error {
	_, err := s.db.Exec("DELETE FROM cache_entries WHERE repo_hash = ?", repoHash)
	return err
}

func (s *SQLiteBackend) Size() (int, error) {
	var size int
	err := s.db.QueryRow("SELECT COUNT(*) FROM cache_entries").Scan(&size)
	return size, err
}

func (s *SQLiteBackend) Close() error {
	return s.db.Close()
}

// NewCache creates a new LRU cache with SQLite or Postgres persistence.
func NewCache(repoRoot string, maxSize int) (*Cache, error) {
	if maxSize <= 0 {
		maxSize = 100
	}

	var backend StorageBackend
	var err error

	pgURL := os.Getenv("COSTWISE_PG_URL")
	if pgURL != "" {
		backend, err = NewPostgresBackend(pgURL)
		if err != nil {
			return nil, fmt.Errorf("failed to init postgres backend: %w", err)
		}
	} else {
		dbDir := filepath.Join(repoRoot, ".mycli-fts")
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create cache directory: %w", err)
		}
		dbPath := filepath.Join(dbDir, "cache.db")
		backend, err = NewSQLiteBackend(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to init sqlite backend: %w", err)
		}
	}

	return &Cache{
		maxSize: maxSize,
		backend: backend,
	}, nil
}

func createCacheTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cache_entries (
			key_hash TEXT PRIMARY KEY,
			repo_hash TEXT NOT NULL,
			query TEXT NOT NULL,
			retriever TEXT NOT NULL,
			context_level TEXT NOT NULL,
			token_budget INTEGER NOT NULL,
			results_json BLOB,
			context_text TEXT,
			tokens INTEGER,
			created_at TIMESTAMP,
			last_accessed TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_cache_repo ON cache_entries(repo_hash);
	`)
	return err
}

func RepoHash(repoRoot string) string {
	h := sha256.Sum256([]byte(repoRoot))
	return hex.EncodeToString(h[:8])
}

func (k CacheKey) hash() string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%d", k.RepoHash, k.Query, k.Retriever, k.ContextLevel, k.TokenBudget)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func (c *Cache) Get(key CacheKey) (*CacheEntry, bool) {
	keyHash := key.hash()
	entry, err := c.backend.Get(keyHash)
	if err == nil && entry != nil {
		c.hits++
		return entry, true
	}
	c.misses++
	return nil, false
}

func (c *Cache) Put(key CacheKey, entry *CacheEntry) {
	keyHash := key.hash()
	_ = c.backend.Put(keyHash, key, entry, c.maxSize)
}

func (c *Cache) Invalidate(repoHash string) {
	_ = c.backend.Invalidate(repoHash)
}

func (c *Cache) Stats() CacheStats {
	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	size, _ := c.backend.Size()

	return CacheStats{
		Hits:    c.hits,
		Misses:  c.misses,
		HitRate: hitRate,
		Size:    size,
		MaxSize: c.maxSize,
	}
}

func (c *Cache) Close() error {
	if c.backend != nil {
		return c.backend.Close()
	}
	return nil
}

func MarshalResults(results interface{}) []byte {
	data, err := json.Marshal(results)
	if err != nil {
		return nil
	}
	return data
}

func UnmarshalResults(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}
