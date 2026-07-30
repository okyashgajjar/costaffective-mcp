package cache

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func TestPostgresBackend_Integration(t *testing.T) {
	pgURL := os.Getenv("COSTWISE_TEST_PG_URL")
	if pgURL == "" {
		t.Skip("Skipping Postgres integration test; set COSTWISE_TEST_PG_URL")
	}

	backend, err := NewPostgresBackend(pgURL)
	if err != nil {
		t.Fatalf("failed to init postgres backend: %v", err)
	}
	defer backend.Close()

	// Clear table for clean test
	_, err = backend.db.Exec("DELETE FROM cache_entries")
	if err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}

	key := CacheKey{
		RepoHash:     "repo123",
		Query:        "test query",
		Retriever:    "auto",
		ContextLevel: "detailed",
		TokenBudget:  1000,
	}
	keyHash := key.hash()

	entry := &CacheEntry{
		ResultsJSON: []byte(`{"test": true}`),
		Context:     "test context",
		Tokens:      42,
		CreatedAt:   time.Now(),
	}

	// 1. Test Put
	err = backend.Put(keyHash, key, entry, 100)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// 2. Test Get
	got, err := backend.Get(keyHash)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(got.ResultsJSON) != string(entry.ResultsJSON) {
		t.Errorf("expected %s, got %s", entry.ResultsJSON, got.ResultsJSON)
	}

	// 3. Test Concurrency (100 concurrent Puts to simulate enterprise load)
	var wg sync.WaitGroup
	errs := make(chan error, 100)
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			
			// Simulate slightly different timestamps and token budgets to trigger upserts
			iterKey := key
			iterKey.TokenBudget = 1000 + idx
			
			err := backend.Put(iterKey.hash(), iterKey, entry, 100)
			if err != nil {
				errs <- err
			}
		}(i)
	}
	
	wg.Wait()
	close(errs)
	
	for err := range errs {
		t.Errorf("Concurrent Put failed: %v", err)
	}
	
	// Verify Size
	size, err := backend.Size()
	if err != nil {
		t.Fatalf("Size failed: %v", err)
	}
	
	if size == 0 {
		t.Errorf("Expected size > 0, got 0")
	}
	
	// Test Invalidate
	err = backend.Invalidate("repo123")
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}
	
	size, _ = backend.Size()
	if size != 0 {
		t.Errorf("Expected size 0 after invalidate, got %d", size)
	}
}
