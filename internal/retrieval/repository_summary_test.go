package retrieval

import (
	"testing"
)

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"foo_test.go", true},
		{"foo_test.py", true},
		{"test_foo.py", true},
		{"foo_spec.rb", true},
		{"FooTest.java", true},
		{"main.go", false},
		{"app.py", false},
		{"index.js", false},
		{"helper.ts", false},
		{"README.md", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			got := isTestFile(tc.path)
			if got != tc.want {
				t.Errorf("isTestFile(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestBuildRepositorySummaryNil(t *testing.T) {
	summary, text := BuildRepositorySummary(nil)
	if summary != nil {
		t.Error("expected nil summary for nil KnowledgeStore")
	}
	if text != "" {
		t.Errorf("expected empty text for nil KnowledgeStore, got %q", text)
	}
}

func TestArchitectureMapEntries(t *testing.T) {
	modules := []ModuleInfo{
		{
			Name:  "cmd",
			Path:  "cmd/",
			Files: []string{"cmd/main.go", "cmd/helper.go"},
		},
		{
			Name:  "internal",
			Path:  "internal/",
			Files: []string{"internal/app.go", "internal/server.go"},
		},
	}

	entries := extractArchitectureEntries(modules)
	if len(entries) == 0 {
		t.Error("expected at least one architecture entry for cmd/main.go")
	}

	found := false
	for _, e := range entries {
		if e == "cmd/main.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected cmd/main.go in entries, got %v", entries)
	}
}

func TestExtractLayers(t *testing.T) {
	modules := []ModuleInfo{
		{Name: "cmd", Path: "cmd/"},
		{Name: "internal", Path: "internal/"},
		{Name: "pkg", Path: "pkg/"},
	}
	layers := extractLayers(modules)
	if len(layers) != 3 {
		t.Errorf("expected 3 layers, got %d", len(layers))
	}
}

func TestRepositorySummaryFormat(t *testing.T) {
	rs := &RepositorySummary{
		FileCount:   10,
		SymbolCount: 50,
		TestFiles:   3,
		LanguageMix: map[string]int{"Go": 7, "Python": 3},
		Modules: []ModuleSummary{
			{Name: "cmd", Path: "cmd/", FileCount: 2, SymbolCount: 10, Language: "Go"},
			{Name: "internal", Path: "internal/", FileCount: 5, SymbolCount: 30, Language: "Go"},
		},
		Architecture: ArchitectureMap{
			Layers:  []string{"cmd/", "internal/"},
			Entries: []string{"cmd/main.go"},
		},
	}

	text := rs.Format()
	if !contains(text, "Files: 10") {
		t.Errorf("expected 'Files: 10' in output, got:\n%s", text)
	}
	if !contains(text, "Symbols: 50") {
		t.Errorf("expected 'Symbols: 50' in output, got:\n%s", text)
	}
	if !contains(text, "Test Files: 3") {
		t.Errorf("expected 'Test Files: 3' in output, got:\n%s", text)
	}
	if !contains(text, "Go") {
		t.Errorf("expected 'Go' in language mix, got:\n%s", text)
	}
	if !contains(text, "cmd") || !contains(text, "internal") {
		t.Errorf("expected module names in output, got:\n%s", text)
	}
}

func TestRepositorySummaryEmpty(t *testing.T) {
	rs := &RepositorySummary{
		Modules: []ModuleSummary{},
	}
	text := rs.Format()
	if text == "" {
		t.Error("expected non-empty format even for empty summary")
	}
}
