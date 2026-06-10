package kmemory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type KnowledgeType int

const (
	SymbolKnowledge KnowledgeType = iota
	CallerKnowledge
	ReferenceKnowledge
	GrepKnowledge
	GlobKnowledge
	ArchitectureKnowledge
	RepoSummary
	ModuleOwnership
	FileOwnership
)

func (kt KnowledgeType) String() string {
	switch kt {
	case SymbolKnowledge:
		return "symbol"
	case CallerKnowledge:
		return "caller"
	case ReferenceKnowledge:
		return "reference"
	case GrepKnowledge:
		return "grep"
	case GlobKnowledge:
		return "glob"
	case ArchitectureKnowledge:
		return "architecture"
	case RepoSummary:
		return "repo_summary"
	case ModuleOwnership:
		return "module_ownership"
	case FileOwnership:
		return "file_ownership"
	default:
		return "unknown"
	}
}

type KnowledgeEntry struct {
	Type       KnowledgeType `json:"type"`
	Key        string        `json:"key"`
	Value      string        `json:"value"`
	Metadata   []string      `json:"metadata,omitempty"`
	File       string        `json:"file,omitempty"`
	Line       int           `json:"line,omitempty"`
	Confidence float64       `json:"confidence"`
	CreatedAt  time.Time     `json:"created_at"`
	LastUsedAt time.Time     `json:"last_used_at"`
	HitCount   int           `json:"hit_count"`
}

type KnowledgeMemory struct {
	mu      sync.RWMutex
	entries map[string]*KnowledgeEntry
	index   map[string][]string
}

func NewKnowledgeMemory() *KnowledgeMemory {
	return &KnowledgeMemory{
		entries: make(map[string]*KnowledgeEntry),
		index:   make(map[string][]string),
	}
}

func entryKey(kt KnowledgeType, key string) string {
	return fmt.Sprintf("%s:%s", kt.String(), strings.ToLower(key))
}

func (km *KnowledgeMemory) Store(kt KnowledgeType, key string, entry *KnowledgeEntry) {
	km.mu.Lock()
	defer km.mu.Unlock()

	ek := entryKey(kt, key)
	entry.Key = key
	entry.Type = kt
	entry.CreatedAt = time.Now()
	entry.LastUsedAt = time.Now()
	entry.HitCount = 1
	km.entries[ek] = entry

	indexKey := strings.ToLower(key)
	km.index[indexKey] = append(km.index[indexKey], ek)

	for _, token := range tokenizeKey(key) {
		km.index[token] = append(km.index[token], ek)
	}
}

func (km *KnowledgeMemory) Lookup(kt KnowledgeType, key string) *KnowledgeEntry {
	km.mu.Lock()
	defer km.mu.Unlock()

	ek := entryKey(kt, key)
	if e, ok := km.entries[ek]; ok {
		e.LastUsedAt = time.Now()
		e.HitCount++
		return e
	}
	return nil
}

func (km *KnowledgeMemory) Search(kt KnowledgeType, query string) []*KnowledgeEntry {
	km.mu.RLock()
	defer km.mu.RUnlock()

	lower := strings.ToLower(query)
	seen := make(map[string]bool)
	var results []*KnowledgeEntry

	prefix := kt.String() + ":"

	if entries, ok := km.index[lower]; ok {
		for _, ek := range entries {
			if strings.HasPrefix(ek, prefix) && !seen[ek] {
				if e, ok2 := km.entries[ek]; ok2 {
					seen[ek] = true
					results = append(results, e)
				}
			}
		}
		return results
	}

	for _, ek := range km.entries {
		if !strings.HasPrefix(ek.Key, prefix) {
			continue
		}
		if strings.Contains(strings.ToLower(ek.Key), lower) ||
			strings.Contains(strings.ToLower(ek.Value), lower) ||
			containsAny(strings.ToLower(ek.File), lower) {
			if !seen[ek.Key] {
				seen[ek.Key] = true
				results = append(results, ek)
			}
		}
	}

	return results
}

func (km *KnowledgeMemory) SearchAll(query string) []*KnowledgeEntry {
	km.mu.RLock()
	defer km.mu.RUnlock()

	lower := strings.ToLower(query)
	seen := make(map[string]bool)
	var results []*KnowledgeEntry

	for _, ek := range km.entries {
		if strings.Contains(strings.ToLower(ek.Key), lower) ||
			strings.Contains(strings.ToLower(ek.Value), lower) ||
			containsAny(strings.ToLower(ek.File), lower) {
			if !seen[ek.Key] {
				seen[ek.Key] = true
				results = append(results, ek)
			}
		}
	}

	return results
}

func (km *KnowledgeMemory) Stats() map[string]int {
	km.mu.RLock()
	defer km.mu.RUnlock()

	stats := make(map[string]int)
	for _, e := range km.entries {
		stats[e.Type.String()]++
	}
	return stats
}

func (km *KnowledgeMemory) Snapshot() []*KnowledgeEntry {
	km.mu.RLock()
	defer km.mu.RUnlock()

	entries := make([]*KnowledgeEntry, 0, len(km.entries))
	for _, e := range km.entries {
		entries = append(entries, e)
	}
	return entries
}

func (km *KnowledgeMemory) SaveToFile(path string) error {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	snapshot := struct {
		UpdatedAt time.Time         `json:"updated_at"`
		Entries   []*KnowledgeEntry `json:"entries"`
	}{
		UpdatedAt: time.Now(),
	}

	for _, e := range km.entries {
		snapshot.Entries = append(snapshot.Entries, e)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (km *KnowledgeMemory) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var snapshot struct {
		UpdatedAt time.Time         `json:"updated_at"`
		Entries   []*KnowledgeEntry `json:"entries"`
	}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	for _, e := range snapshot.Entries {
		ek := entryKey(e.Type, e.Key)
		e.HitCount = 1
		km.entries[ek] = e

		indexKey := strings.ToLower(e.Key)
		km.index[indexKey] = append(km.index[indexKey], ek)
		for _, token := range tokenizeKey(e.Key) {
			km.index[token] = append(km.index[token], ek)
		}
	}

	return nil
}

func tokenizeKey(key string) []string {
	key = strings.ToLower(key)
	words := strings.FieldsFunc(key, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})

	var tokens []string
	seen := make(map[string]bool)
	for _, w := range words {
		if len(w) >= 2 && !seen[w] {
			seen[w] = true
			tokens = append(tokens, w)
		}
	}
	return tokens
}

func containsAny(s, substr string) bool {
	return s != "" && strings.Contains(s, substr)
}

func NewSymbolEntry(symbolName, file, definition string, callers []string) *KnowledgeEntry {
	return &KnowledgeEntry{
		Type:       SymbolKnowledge,
		Key:        symbolName,
		Value:      definition,
		File:       file,
		Confidence: 1.0,
	}
}

func NewReferenceEntry(symbol string, files []string) *KnowledgeEntry {
	return &KnowledgeEntry{
		Type:       ReferenceKnowledge,
		Key:        symbol,
		Value:      strings.Join(files, ", "),
		Metadata:   files,
		Confidence: 0.8,
	}
}

func NewCallerEntry(symbol string, callers []string) *KnowledgeEntry {
	return &KnowledgeEntry{
		Type:       CallerKnowledge,
		Key:        symbol,
		Value:      strings.Join(callers, ", "),
		Metadata:   callers,
		Confidence: 0.85,
	}
}
