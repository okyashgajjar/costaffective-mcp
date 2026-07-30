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

// Cache is an LRU cache persistent (SQLite) layers.
type Cache struct {
	maxSize int
	db      *sql.DB
	hits    int64
	misses  int64
}

// NewCache creates a new LRU cache with SQLite persistence.
func NewCache(repoRoot string, maxSize int) (*Cache, error) {
	if maxSize <= 0 {
		maxSize = 100
	}

	dbDir := filepath.Join(repoRoot, ".mycli-fts")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	dbPath := filepath.Join(dbDir, "cache.db")

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open cache DB: %w", err)
	}

	if err := createCacheTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Cache{
		maxSize: maxSize,
		db:      db,
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

	row := c.db.QueryRow(`
		SELECT results_json, context_text, tokens, created_at 
		FROM cache_entries WHERE key_hash = ?`, keyHash)

	var entry CacheEntry
	if err := row.Scan(&entry.ResultsJSON, &entry.Context, &entry.Tokens, &entry.CreatedAt); err == nil {
		c.hits++
		_, _ = c.db.Exec("UPDATE cache_entries SET last_accessed = ? WHERE key_hash = ?", time.Now(), keyHash)
		return &entry, true
	}

	c.misses++
	return nil, false
}

func (c *Cache) Put(key CacheKey, entry *CacheEntry) {
	keyHash := key.hash()
	now := time.Now()

	_, _ = c.db.Exec(`
		INSERT OR REPLACE INTO cache_entries
			(key_hash, repo_hash, query, retriever, context_level, token_budget,
			 results_json, context_text, tokens, created_at, last_accessed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, keyHash, key.RepoHash, key.Query, key.Retriever, key.ContextLevel, key.TokenBudget,
		entry.ResultsJSON, entry.Context, entry.Tokens, entry.CreatedAt, now)

	// SQLite-based eviction
	_, _ = c.db.Exec(`
		DELETE FROM cache_entries WHERE key_hash NOT IN (
			SELECT key_hash FROM cache_entries ORDER BY last_accessed DESC LIMIT ?
		)
	`, c.maxSize)
}

func (c *Cache) Invalidate(repoHash string) {
	_, _ = c.db.Exec("DELETE FROM cache_entries WHERE repo_hash = ?", repoHash)
}

func (c *Cache) Stats() CacheStats {
	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	var size int
	_ = c.db.QueryRow("SELECT COUNT(*) FROM cache_entries").Scan(&size)

	return CacheStats{
		Hits:    c.hits,
		Misses:  c.misses,
		HitRate: hitRate,
		Size:    size,
		MaxSize: c.maxSize,
	}
}

func (c *Cache) Close() error {
	if c.db != nil {
		return c.db.Close()
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
