package architecture

import (
	"context"
	"testing"
)

func TestExtractTopics(t *testing.T) {
	tests := []struct {
		name      string
		classes   []string
		functions []string
		imports   []string
		filePath  string
		wantAny   []string
	}{
		{
			name:      "exceptions file",
			classes:   []string{"AiderError", "APIError"},
			functions: []string{"handle_exception", "retry_request"},
			filePath:  "aider/exceptions.py",
			wantAny:   []string{"error handling", "failures", "retry logic"},
		},
		{
			name:      "caching file",
			classes:   []string{"ChatChunks"},
			functions: []string{"get_cached_messages", "cache_lookup"},
			filePath:  "aider/caching.py",
			wantAny:   []string{"caching", "performance"},
		},
		{
			name:      "repo mapping file",
			classes:   []string{"RepoMap"},
			functions: []string{"get_repo_map", "build_graph"},
			filePath:  "aider/repomap.py",
			wantAny:   []string{"repository", "git", "source control", "graph", "indexing"},
		},
		{
			name:      "LLM model file",
			functions: []string{"send_chat", "completion"},
			imports:   []string{"litellm"},
			filePath:  "aider/models.py",
			wantAny:   []string{"llm", "language model", "messaging", "communication"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topics := ExtractTopics(tt.classes, tt.functions, tt.imports, tt.filePath)
			topicSet := make(map[string]bool)
			for _, topic := range topics {
				topicSet[topic] = true
			}
			for _, want := range tt.wantAny {
				if !topicSet[want] {
					t.Errorf("expected topic %q in %v", want, topics)
				}
			}
		})
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		lang     string
		expected string
	}{
		{
			name: "python docstring",
			content: `"""This module handles exceptions in the system."""
import os
`,
			lang:     "python",
			expected: "This module handles exceptions in the system.",
		},
		{
			name: "go package comment",
			content: `// Package exceptions provides error handling utilities.
package exceptions
`,
			lang:     "go",
			expected: "Package exceptions provides error handling utilities.",
		},
		{
			name: "no docstring",
			content: `import os
print("hello")
`,
			lang:     "python",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := ExtractDescription(tt.content, "test.py", tt.lang)
			if desc != tt.expected {
				t.Errorf("got %q, want %q", desc, tt.expected)
			}
		})
	}
}

func TestExtractImports(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		lang     string
		expected []string
	}{
		{
			name:     "python imports",
			content:  "import os\nimport sys\nfrom pathlib import Path\nfrom collections import OrderedDict",
			lang:     "python",
			expected: []string{"os", "sys", "pathlib", "collections"},
		},
		{
			name:     "go imports",
			content:  "import (\n\t\"fmt\"\n\t\"os\"\n)\n",
			lang:     "go",
			expected: []string{"fmt", "os"},
		},
		{
			name:     "js imports",
			content:  "import {x} from 'lodash'\nimport y from './local'\nconst z = require('express')",
			lang:     "javascript",
			expected: []string{"lodash", "local", "express"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractImports(tt.content, tt.lang)
			if len(got) != len(tt.expected) {
				t.Errorf("got %d imports %v, want %d %v", len(got), got, len(tt.expected), tt.expected)
				return
			}
			gotMap := make(map[string]bool)
			for _, g := range got {
				gotMap[g] = true
			}
			for _, want := range tt.expected {
				if !gotMap[want] {
					t.Errorf("missing import %q in %v", want, got)
				}
			}
		})
	}
}

func TestHashContent(t *testing.T) {
	h1 := HashContent([]byte("hello"))
	h2 := HashContent([]byte("hello"))
	h3 := HashContent([]byte("world"))
	if h1 != h2 {
		t.Error("same content should produce same hash")
	}
	if h1 == h3 {
		t.Error("different content should produce different hash")
	}
}

func TestIndexerRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	idx, err := NewIndexer(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	content := `"""Test module for error handling."""
class MyError(Exception):
    pass

def handle_error():
    pass
`
	if err := writeFileBytes(tmpDir+"/test_exc.py", []byte(content)); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	count, _, err := idx.IndexRepo(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Error("expected at least 1 file indexed")
	}

	all, err := idx.db.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("no summaries loaded")
	}

	found := false
	for _, s := range all {
		if s.FilePath == "test_exc.py" {
			found = true
			if s.Description != "Test module for error handling." {
				t.Errorf("unexpected description: %q", s.Description)
			}
		}
	}
	if !found {
		t.Error("test_exc.py not found in summaries")
	}
}
