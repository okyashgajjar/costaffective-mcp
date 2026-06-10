package retrieval

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type RepositorySummary struct {
	Purpose      string          `json:"-"`
	Modules      []ModuleSummary `json:"modules"`
	FileCount    int             `json:"file_count"`
	LanguageMix  map[string]int  `json:"language_mix"`
	SymbolCount  int             `json:"symbol_count"`
	TestFiles    int             `json:"test_files"`
	Architecture ArchitectureMap `json:"architecture"`
}

type ModuleSummary struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	FileCount   int      `json:"file_count"`
	SymbolCount int      `json:"symbol_count"`
	Language    string   `json:"language"`
	TopSymbols  []string `json:"top_symbols,omitempty"`
}

type ArchitectureMap struct {
	Layers  []string `json:"layers,omitempty"`
	Entries []string `json:"entries,omitempty"`
}

func BuildRepositorySummary(ks *KnowledgeStore) (*RepositorySummary, string) {
	if ks == nil {
		return nil, ""
	}

	modules, err := ks.GetModules()
	if err != nil {
		return nil, ""
	}

	summaries, err := ks.GetAllFileSummaries()
	if err != nil {
		return nil, ""
	}

	summary := &RepositorySummary{
		Modules:     make([]ModuleSummary, 0, len(modules)),
		LanguageMix: make(map[string]int),
		Architecture: ArchitectureMap{
			Layers:  extractLayers(modules),
			Entries: extractArchitectureEntries(modules),
		},
	}

	totalSymbols := 0
	testCount := 0

	for _, mod := range modules {
		topSymbols := extractTopSymbols(ks, mod)
		m := ModuleSummary{
			Name:        mod.Name,
			Path:        mod.Path,
			FileCount:   len(mod.Files),
			SymbolCount: mod.Symbols,
			Language:    mod.Language,
			TopSymbols:  topSymbols,
		}
		summary.Modules = append(summary.Modules, m)
		totalSymbols += mod.Symbols
		if mod.Language != "" {
			summary.LanguageMix[mod.Language] += len(mod.Files)
		}
	}

	for _, s := range summaries {
		if isTestFile(s.Path) {
			testCount++
		}
	}

	summary.FileCount = len(summaries)
	summary.SymbolCount = totalSymbols
	summary.TestFiles = testCount

	sort.Slice(summary.Modules, func(i, j int) bool {
		return summary.Modules[i].SymbolCount > summary.Modules[j].SymbolCount
	})

	text := summary.Format()
	return summary, text
}

func (rs *RepositorySummary) Format() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Files: %d\n", rs.FileCount)
	fmt.Fprintf(&b, "Symbols: %d\n", rs.SymbolCount)
	fmt.Fprintf(&b, "Test Files: %d\n", rs.TestFiles)

	if len(rs.LanguageMix) > 0 {
		var langs []string
		for lang, count := range rs.LanguageMix {
			langs = append(langs, fmt.Sprintf("%s:%d", lang, count))
		}
		sort.Strings(langs)
		fmt.Fprintf(&b, "Languages: %s\n", strings.Join(langs, ", "))
	}

	b.WriteString("Modules:\n")
	for _, m := range rs.Modules {
		fmt.Fprintf(&b, "  %s (%s) - %d files, %d symbols\n", m.Name, m.Language, m.FileCount, m.SymbolCount)
		if len(m.TopSymbols) > 0 {
			top := m.TopSymbols
			if len(top) > 3 {
				top = top[:3]
			}
			fmt.Fprintf(&b, "    Symbols: %s\n", strings.Join(top, ", "))
		}
	}

	if len(rs.Architecture.Layers) > 0 {
		fmt.Fprintf(&b, "Layers: %s\n", strings.Join(rs.Architecture.Layers, " -> "))
	}
	if len(rs.Architecture.Entries) > 0 {
		fmt.Fprintf(&b, "Entry Points: %s\n", strings.Join(rs.Architecture.Entries, ", "))
	}

	return b.String()
}

func extractLayers(modules []ModuleInfo) []string {
	var layers []string
	seen := make(map[string]bool)
	for _, m := range modules {
		dir := m.Path
		if !seen[dir] {
			seen[dir] = true
			layers = append(layers, dir)
		}
	}
	return layers
}

func extractArchitectureEntries(modules []ModuleInfo) []string {
	var entries []string
	for _, m := range modules {
		for _, f := range m.Files {
			base := strings.ToLower(filepath.Base(f))
			if base == "main.go" || base == "main.py" || base == "index.js" || base == "app.go" || base == "cmd.go" || strings.HasPrefix(base, "main.") {
				entries = append(entries, f)
			}
		}
	}
	return entries
}

func extractTopSymbols(ks *KnowledgeStore, mod ModuleInfo) []string {
	var symbols []string
	for _, f := range mod.Files {
		fs, err := ks.GetFileSummary(f)
		if err != nil {
			continue
		}
		symbols = append(symbols, fs.Functions...)
		symbols = append(symbols, fs.Classes...)
	}
	if len(symbols) > 5 {
		symbols = symbols[:5]
	}
	return symbols
}

func isTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, "_test.py") || strings.HasSuffix(base, "_test.js") || strings.HasSuffix(base, "_test.ts") || strings.HasSuffix(base, "test.java") || strings.HasSuffix(base, "spec.rb") || strings.Contains(base, "test_") || strings.Contains(base, "_spec.")
}
