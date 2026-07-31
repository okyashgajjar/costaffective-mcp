package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSemanticSearch_Synonyms(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "semantic_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	idx, err := NewSemanticIndexer(tmpDir)
	if err != nil {
		t.Fatalf("failed to create indexer: %v", err)
	}
	defer idx.Close()

	// Index a file
	err = idx.IndexFile("auth/login.go", "func Login() {}", "Handles user authentication and login_flow")
	if err != nil {
		t.Fatalf("failed to index file: %v", err)
	}

	// Search for something that matches
	results, err := idx.Search("authentication", 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected to find auth/login.go, got 0 results")
	}

	if results[0].Path != "auth/login.go" {
		t.Errorf("expected path auth/login.go, got %s", results[0].Path)
	}
}

func TestSemanticIndex_HeavyLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "semantic_heavy")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	idx, err := NewSemanticIndexer(tmpDir)
	if err != nil {
		t.Fatalf("failed to create indexer: %v", err)
	}
	defer idx.Close()

	// Index 5000 files
	for i := 0; i < 5000; i++ {
		path := filepath.Join("src", "pkg", string(rune('a'+(i%26))), "file.go")
		err := idx.IndexFile(path, "func Test() {}", "this is a test document")
		if err != nil {
			t.Fatalf("failed to index file at iter %d: %v", i, err)
		}
	}

	results, err := idx.Search("document", 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected results, got 0")
	}
}
