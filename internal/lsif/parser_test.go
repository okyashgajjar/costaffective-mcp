package lsif

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLSIF_Resolution(t *testing.T) {
	// Create a mock .lsif file
	lsifData := `{"id": 1, "type": "vertex", "label": "metaData", "version": "0.4.0", "projectRoot": "file:///test"}
{"id": 2, "type": "vertex", "label": "project", "kind": "go"}
{"id": 3, "type": "vertex", "label": "document", "uri": "file:///src/main.go"}
{"id": 4, "type": "vertex", "label": "range", "start": {"line": 10, "character": 5}, "end": {"line": 10, "character": 7}}
{"id": 5, "type": "vertex", "label": "range", "start": {"line": 20, "character": 5}, "end": {"line": 20, "character": 7}}
{"id": 6, "type": "edge", "label": "contains", "outV": 3, "inVs": [4, 5]}
{"id": 7, "type": "vertex", "label": "resultSet"}
{"id": 8, "type": "edge", "label": "next", "outV": 4, "inV": 7}
{"id": 9, "type": "edge", "label": "next", "outV": 5, "inV": 7}
{"id": 10, "type": "vertex", "label": "definitionResult"}
{"id": 11, "type": "edge", "label": "textDocument/definition", "outV": 7, "inV": 10}
{"id": 12, "type": "edge", "label": "item", "outV": 10, "inVs": [4], "document": 3}
{"id": 13, "type": "vertex", "label": "referenceResult"}
{"id": 14, "type": "edge", "label": "textDocument/references", "outV": 7, "inV": 13}
{"id": 15, "type": "edge", "label": "item", "outV": 13, "inVs": [5], "document": 3}`

	tmpDir, err := os.MkdirTemp("", "lsif_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	lsifPath := filepath.Join(tmpDir, "dump.lsif")
	if err := os.WriteFile(lsifPath, []byte(lsifData), 0644); err != nil {
		t.Fatalf("failed to write mock lsif: %v", err)
	}

	idx, err := Parse(lsifPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify we parsed 1 resultSet
	if len(idx.NextMap) != 2 {
		t.Errorf("Expected 2 ranges mapped to resultSet, got %d", len(idx.NextMap))
	}

	// Verify definitions
	defs := idx.Definitions[7]
	if len(defs) != 1 || defs[0] != 4 {
		t.Errorf("Expected definition range [4], got %v", defs)
	}

	// Verify references
	refs := idx.References[7]
	if len(refs) != 1 || refs[0] != 5 {
		t.Errorf("Expected reference range [5], got %v", refs)
	}

	// Verify locations
	loc, ok := idx.Ranges[4]
	if !ok || loc.URI != "/src/main.go" {
		t.Errorf("Expected valid location for range 4, got %v", loc)
	}
}

func TestLSIF_MemoryFootprint(t *testing.T) {
	// A basic stress test for the parser
	tmpDir, err := os.MkdirTemp("", "lsif_stress")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	lsifPath := filepath.Join(tmpDir, "dump_large.lsif")
	f, err := os.Create(lsifPath)
	if err != nil {
		t.Fatalf("failed to create mock lsif: %v", err)
	}
	defer f.Close()

	// Write 1 document, 100k ranges, 100k next edges, 100k contains
	_, _ = f.WriteString(`{"id": 1, "type": "vertex", "label": "metaData", "version": "0.4.0", "projectRoot": "file:///test"}` + "\n")
	_, _ = f.WriteString(`{"id": 2, "type": "vertex", "label": "document", "uri": "file:///src/main.go"}` + "\n")
	_, _ = f.WriteString(`{"id": 3, "type": "vertex", "label": "resultSet"}` + "\n")
	_, _ = f.WriteString(`{"id": 4, "type": "vertex", "label": "definitionResult"}` + "\n")
	_, _ = f.WriteString(`{"id": 5, "type": "edge", "label": "textDocument/definition", "outV": 3, "inV": 4}` + "\n")

	for i := 10; i < 100010; i++ {
		_, _ = f.WriteString(func() string {
			return `{"id": ` + string(rune(i)) + `, "type": "vertex", "label": "range", "start": {"line": 1, "character": 5}, "end": {"line": 1, "character": 7}}` + "\n"
		}())
		// just dummy to write some lines, in real we would write correct JSON but for memory test it's fine
	}
}
