package treesitter

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SymbolDB struct {
	db       *sql.DB
	queries  *Queries
	repoRoot string
}

func NewSymbolDB(repoRoot string) (*SymbolDB, error) {
	hash := sha256.Sum256([]byte(repoRoot))
	dbName := fmt.Sprintf("symbols_%x.db", hash[:8])
	dbDir := filepath.Join(repoRoot, ".mycli-fts")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dbDir, dbName)

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open symbol DB: %w", err)
	}

	if err := createSymbolTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &SymbolDB{db: db, queries: New(db), repoRoot: repoRoot}, nil
}

const schemaVersion = 5

func createSymbolTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS symbols (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			language TEXT NOT NULL DEFAULT '',
			file TEXT NOT NULL,
			start_line INTEGER NOT NULL DEFAULT 0,
			end_line INTEGER NOT NULL DEFAULT 0,
			signature TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_symbols_name ON symbols(name);
		CREATE INDEX IF NOT EXISTS idx_symbols_kind ON symbols(kind);
		CREATE INDEX IF NOT EXISTS idx_symbols_file ON symbols(file);
		CREATE INDEX IF NOT EXISTS idx_symbols_name_kind ON symbols(name, kind);

		CREATE TABLE IF NOT EXISTS symbol_files (
			file_path TEXT PRIMARY KEY,
			file_hash TEXT NOT NULL DEFAULT '',
			last_indexed TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS references_t (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol_name TEXT NOT NULL,
			file TEXT NOT NULL,
			line INTEGER NOT NULL DEFAULT 0,
			col INTEGER NOT NULL DEFAULT 0,
			ref_type TEXT NOT NULL DEFAULT 'reference',
			context TEXT NOT NULL DEFAULT ''
		);

		CREATE INDEX IF NOT EXISTS idx_refs_name ON references_t(symbol_name);
		CREATE INDEX IF NOT EXISTS idx_refs_file ON references_t(file);
		CREATE INDEX IF NOT EXISTS idx_refs_type ON references_t(ref_type);
		CREATE INDEX IF NOT EXISTS idx_refs_name_type ON references_t(symbol_name, ref_type);

		CREATE TABLE IF NOT EXISTS call_edges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			caller_name TEXT NOT NULL,
			caller_file TEXT NOT NULL DEFAULT '',
			callee_name TEXT NOT NULL,
			file TEXT NOT NULL,
			line INTEGER NOT NULL DEFAULT 0,
			language TEXT NOT NULL DEFAULT ''
		);

		CREATE INDEX IF NOT EXISTS idx_calls_callee ON call_edges(callee_name);
		CREATE INDEX IF NOT EXISTS idx_calls_caller ON call_edges(caller_name);
		CREATE INDEX IF NOT EXISTS idx_calls_file ON call_edges(file);

		CREATE TABLE IF NOT EXISTS ontology_tags (
			symbol_id TEXT NOT NULL,
			tag TEXT NOT NULL,
			domain TEXT NOT NULL DEFAULT '',
			file TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_ontology_tags_tag ON ontology_tags(tag);
		CREATE INDEX IF NOT EXISTS idx_ontology_tags_domain ON ontology_tags(domain);
		CREATE INDEX IF NOT EXISTS idx_ontology_tags_symbol ON ontology_tags(symbol_id);
		CREATE INDEX IF NOT EXISTS idx_ontology_tags_file ON ontology_tags(file);
	`)
	if err != nil {
		return err
	}

	var ver int
	_ = db.QueryRow("SELECT version FROM schema_version").Scan(&ver)
	if ver != schemaVersion {
		if _, err := db.Exec("DELETE FROM schema_version"); err != nil {
			return err
		}
		if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (?)", schemaVersion); err != nil {
			return err
		}
		if _, err := db.Exec("DELETE FROM symbols"); err != nil {
			return err
		}
		if _, err := db.Exec("DELETE FROM symbol_files"); err != nil {
			return err
		}
		if _, err := db.Exec("DELETE FROM references_t"); err != nil {
			return err
		}
		if _, err := db.Exec("DELETE FROM call_edges"); err != nil {
			return err
		}
		if _, err := db.Exec("DELETE FROM ontology_tags"); err != nil {
			return err
		}
	}
	return nil
}

func (s *SymbolDB) Close() error {
	return s.db.Close()
}

func SymbolID(name, kind, file string, line int) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%d", name, kind, file, line)))
	return hex.EncodeToString(h[:8])
}

func (s *SymbolDB) StoreSymbols(symbols []Symbol) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	ctx := context.Background()

	for _, sym := range symbols {
		id := SymbolID(sym.Name, string(sym.Kind), sym.File, sym.StartLine)
		if err := q.InsertSymbol(ctx, InsertSymbolParams{
			ID:        id,
			Name:      sym.Name,
			Kind:      string(sym.Kind),
			Language:  sym.Language,
			File:      sym.File,
			StartLine: int64(sym.StartLine),
			EndLine:   int64(sym.EndLine),
			Signature: sym.Signature,
			Content:   sym.Content,
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SymbolDB) ClearFile(filePath string) error {
	_ = s.queries.ClearFileOntologyTags(context.Background(), filePath)
	return s.queries.ClearFile(context.Background(), filePath)
}

func (s *SymbolDB) MarkFileIndexed(filePath, hash string) error {
	return s.queries.InsertSymbolFile(context.Background(), InsertSymbolFileParams{
		FilePath:    filePath,
		FileHash:    hash,
		LastIndexed: time.Now(),
	})
}

func (s *SymbolDB) GetFileHash(filePath string) string {
	hash, _ := s.queries.GetFileHash(context.Background(), filePath)
	return hash
}

func (s *SymbolDB) StoreReferences(refs []Reference) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	ctx := context.Background()

	for _, ref := range refs {
		refType := ref.RefType.String()
		if err := q.InsertReference(ctx, InsertReferenceParams{
			SymbolName: ref.SymbolName,
			File:       ref.File,
			Line:       int64(ref.Line),
			Col:        int64(ref.Column),
			RefType:    refType,
			Context:    ref.Context,
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SymbolDB) ClearFileReferences(filePath string) error {
	return s.queries.ClearFileReferences(context.Background(), filePath)
}

func mapReferences(rows []SearchReferencesRow) []Reference {
	var refs []Reference
	for _, row := range rows {
		ref := Reference{
			SymbolName: row.SymbolName,
			File:       row.File,
			Line:       int(row.Line),
			Column:     int(row.Col),
			Context:    row.Context,
		}
		switch row.RefType {
		case "definition":
			ref.RefType = RefDefinition
		case "reference":
			ref.RefType = RefReference
		case "import":
			ref.RefType = RefImport
		case "export":
			ref.RefType = RefExport
		}
		refs = append(refs, ref)
	}
	return refs
}

func (s *SymbolDB) SearchReferences(symbolName string) ([]Reference, error) {
	rows, err := s.queries.SearchReferences(context.Background(), symbolName)
	if err != nil {
		return nil, err
	}
	return mapReferences(rows), nil
}

func (s *SymbolDB) SearchReferencesLike(partial string) ([]Reference, error) {
	rows, err := s.queries.SearchReferencesLike(context.Background(), "%"+partial+"%")
	if err != nil {
		return nil, err
	}
	return mapReferences(rows), nil
}

func (s *SymbolDB) StoreCallEdges(edges []CallEdge) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	ctx := context.Background()

	for _, e := range edges {
		if err := q.InsertCallEdge(ctx, InsertCallEdgeParams{
			CallerName: e.CallerName,
			CallerFile: e.CallerFile,
			CalleeName: e.CalleeName,
			File:       e.File,
			Line:       int64(e.Line),
			Language:   e.Language,
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SymbolDB) ClearFileCallEdges(filePath string) error {
	return s.queries.ClearFileCallEdges(context.Background(), filePath)
}

func mapCallEdges(rows []SearchCallEdgesRow) []CallEdge {
	var edges []CallEdge
	for _, row := range rows {
		edges = append(edges, CallEdge{
			ID:         int(row.ID),
			CallerName: row.CallerName,
			CallerFile: row.CallerFile,
			CalleeName: row.CalleeName,
			File:       row.File,
			Line:       int(row.Line),
			Language:   row.Language,
		})
	}
	return edges
}

func (s *SymbolDB) SearchCallEdgesByCaller(callerName string) ([]CallEdge, error) {
	rows, err := s.queries.SearchCallEdgesByCaller(context.Background(), callerName)
	if err != nil {
		return nil, err
	}
	return mapCallEdges(rows), nil
}

func (s *SymbolDB) SearchCallEdges(calleeName string) ([]CallEdge, error) {
	rows, err := s.queries.SearchCallEdges(context.Background(), calleeName)
	if err != nil {
		return nil, err
	}
	return mapCallEdges(rows), nil
}

func (s *SymbolDB) SearchCallEdgesLike(partial string) ([]CallEdge, error) {
	rows, err := s.queries.SearchCallEdgesLike(context.Background(), "%"+partial+"%")
	if err != nil {
		return nil, err
	}
	return mapCallEdges(rows), nil
}

func (s *SymbolDB) GetCallEdgeCount() int {
	c, _ := s.queries.GetCallEdgeCount(context.Background())
	return int(c)
}

func (s *SymbolDB) GetReferenceCount() int {
	c, _ := s.queries.GetReferenceCount(context.Background())
	return int(c)
}

func (s *SymbolDB) Search(query string, limit int) ([]SymbolMatch, error) {
	if limit <= 0 {
		limit = 20
	}

	exact := query
	prefix := query + "%"
	partial := "%" + query + "%"

	rows, err := s.queries.Search(context.Background(), SearchParams{
		Name:      exact,
		Name_2:    prefix,
		Signature: partial,
		Name_3:    exact,
		Name_4:    prefix,
		Limit:     int64(limit),
	})
	if err != nil {
		return nil, err
	}

	var matches []SymbolMatch
	for _, row := range rows {
		sym := Symbol{
			Name:      row.Name,
			Kind:      SymbolKind(row.Kind),
			Language:  row.Language,
			File:      row.File,
			StartLine: int(row.StartLine),
			EndLine:   int(row.EndLine),
			Signature: row.Signature,
		}
		score := computeSymbolScore(query, sym)
		matches = append(matches, SymbolMatch{
			Symbol: sym,
			Score:  score,
		})
	}

	return matches, nil
}

func (s *SymbolDB) SearchByKind(kind SymbolKind, limit int) ([]SymbolMatch, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.queries.SearchByKind(context.Background(), SearchByKindParams{
		Kind:  string(kind),
		Limit: int64(limit),
	})
	if err != nil {
		return nil, err
	}

	var matches []SymbolMatch
	for _, row := range rows {
		matches = append(matches, SymbolMatch{
			Symbol: Symbol{
				Name:      row.Name,
				Kind:      SymbolKind(row.Kind),
				Language:  row.Language,
				File:      row.File,
				StartLine: int(row.StartLine),
				EndLine:   int(row.EndLine),
				Signature: row.Signature,
			},
			Score: 0.8,
		})
	}

	return matches, nil
}

func (s *SymbolDB) GetSymbolCount() int {
	c, _ := s.queries.GetSymbolCount(context.Background())
	return int(c)
}

func (s *SymbolDB) GetSymbolRange(name, file string) (int, int, error) {
	row, err := s.queries.GetSymbolRange(context.Background(), GetSymbolRangeParams{
		Name: name,
		File: file,
	})
	if err != nil {
		return 0, 0, err
	}
	return int(row.StartLine), int(row.EndLine), nil
}

func (s *SymbolDB) GetFileSymbols(filePath string) ([]Symbol, error) {
	rows, err := s.queries.GetFileSymbols(context.Background(), filePath)
	if err != nil {
		return nil, err
	}
	var symbols []Symbol
	for _, row := range rows {
		symbols = append(symbols, Symbol{
			Name:      row.Name,
			Kind:      SymbolKind(row.Kind),
			Language:  row.Language,
			File:      row.File,
			StartLine: int(row.StartLine),
			EndLine:   int(row.EndLine),
			Signature: row.Signature,
		})
	}
	return symbols, nil
}

func (s *SymbolDB) GetAllFiles() ([]string, error) {
	return s.queries.GetAllFiles(context.Background())
}

func (s *SymbolDB) GetFilesByHash() (map[string]string, error) {
	rows, err := s.queries.GetFilesByHash(context.Background())
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, row := range rows {
		result[row.FilePath] = row.FileHash
	}
	return result, nil
}

func computeSymbolScore(query string, sym Symbol) float64 {
	q := strings.ToLower(query)
	name := strings.ToLower(sym.Name)

	if name == q {
		return 1.0
	}

	parts := splitCamel(sym.Name)
	qParts := splitCamel(query)

	matchCount := 0
	for _, qp := range qParts {
		qpLower := strings.ToLower(qp)
		for _, np := range parts {
			if strings.ToLower(np) == qpLower {
				matchCount++
				break
			}
		}
	}

	if len(qParts) > 0 {
		ratio := float64(matchCount) / float64(len(qParts))
		if stringsContains(name, q) {
			return 0.8 + ratio*0.2
		}
		return ratio * 0.7
	}

	if stringsContains(name, q) {
		return 0.6
	}

	return 0.3
}

func splitCamel(s string) []string {
	var parts []string
	var cur []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' && len(cur) > 0 {
			parts = append(parts, string(cur))
			cur = []byte{c}
		} else {
			cur = append(cur, c)
		}
	}
	if len(cur) > 0 {
		parts = append(parts, string(cur))
	}
	return parts
}

func stringsContains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func (s *SymbolDB) StoreOntologyTags(tags []InsertOntologyTagParams) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	ctx := context.Background()

	for _, tag := range tags {
		if err := q.InsertOntologyTag(ctx, tag); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SymbolDB) SearchOntology(tag, domain string) ([]SymbolMatch, error) {
	rows, err := s.queries.SearchOntology(context.Background(), SearchOntologyParams{
		Tag:    tag,
		Domain: domain,
	})
	if err != nil {
		return nil, err
	}

	var matches []SymbolMatch
	for _, row := range rows {
		matches = append(matches, SymbolMatch{
			Symbol: Symbol{
				Name:      row.Name,
				Kind:      SymbolKind(row.Kind),
				Language:  row.Language,
				File:      row.File,
				StartLine: int(row.StartLine),
				EndLine:   int(row.EndLine),
				Signature: row.Signature,
			},
			Score: 1.0,
		})
	}

	return matches, nil
}
