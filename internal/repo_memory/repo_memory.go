// Package repo_memory provides persistent repository intelligence storage.
package repo_memory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Symbol represents a repository symbol definition.
type Symbol struct {
	Name       string
	File       string
	Definition string
	LastSeen   time.Time
}

// Reference represents a usage of a symbol.
type Reference struct {
	Symbol  string
	File    string
	Line    int
	Context string
}

// Call represents a call relation.
type Call struct {
	Caller string
	Callee string
	File   string
	Line   int
}

// RepoMemory wraps the SQLite DB.
type RepoMemory struct {
	db *sql.DB
}

// Init opens (or creates) the repository memory DB at the given path.
func Init(dbPath string) (*RepoMemory, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("creating directory for repo memory: %w", err)
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening repo memory DB: %w", err)
	}
	rm := &RepoMemory{db: db}
	if err := rm.migrate(); err != nil {
		return nil, err
	}
	return rm, nil
}

func (rm *RepoMemory) migrate() error {
	// Create tables if they do not exist.
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS symbols (
            name TEXT PRIMARY KEY,
            file TEXT NOT NULL,
            definition TEXT,
            last_seen TIMESTAMP NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS symbol_references (
            symbol TEXT,
            file TEXT,
            line INTEGER,
            context TEXT,
            PRIMARY KEY(symbol, file, line)
        ) WITHOUT ROWID;`,
		`CREATE TABLE IF NOT EXISTS calls (
            caller TEXT,
            callee TEXT,
            file TEXT,
            line INTEGER,
            PRIMARY KEY(caller, callee, file, line)
        ) WITHOUT ROWID;`,
		`CREATE TABLE IF NOT EXISTS modules (
            name TEXT PRIMARY KEY,
            path TEXT NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS architecture (
            key TEXT PRIMARY KEY,
            value TEXT NOT NULL
        )`,
	}
	for _, s := range stmts {
		if _, err := rm.db.Exec(s); err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
	}
	return nil
}

// LearnSymbol upserts a symbol definition.
func (rm *RepoMemory) LearnSymbol(sym Symbol) error {
	_, err := rm.db.Exec(`INSERT INTO symbols(name, file, definition, last_seen) VALUES(?, ?, ?, ?)
        ON CONFLICT(name) DO UPDATE SET file=excluded.file, definition=excluded.definition, last_seen=excluded.last_seen`,
		sym.Name, sym.File, sym.Definition, time.Now())
	return err
}

// GetSymbol retrieves a symbol by name.
func (rm *RepoMemory) GetSymbol(name string) (*Symbol, bool, error) {
	row := rm.db.QueryRow(`SELECT name, file, definition, last_seen FROM symbols WHERE name = ?`, name)
	var s Symbol
	var ts string
	err := row.Scan(&s.Name, &s.File, &s.Definition, &ts)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		s.LastSeen = t
	}
	return &s, true, nil
}

// LearnReference stores a reference.
func (rm *RepoMemory) LearnReference(ref Reference) error {
	_, err := rm.db.Exec(`INSERT OR IGNORE INTO references(symbol, file, line, context) VALUES(?, ?, ?, ?)`,
		ref.Symbol, ref.File, ref.Line, ref.Context)
	return err
}

// GetReferences returns all references for a symbol.
func (rm *RepoMemory) GetReferences(symbol string) ([]Reference, error) {
	rows, err := rm.db.Query(`SELECT file, line, context FROM references WHERE symbol = ?`, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refs []Reference
	for rows.Next() {
		var r Reference
		r.Symbol = symbol
		if err := rows.Scan(&r.File, &r.Line, &r.Context); err != nil {
			return nil, err
		}
		refs = append(refs, r)
	}
	return refs, nil
}

// LearnCall stores a call relation.
func (rm *RepoMemory) LearnCall(c Call) error {
	_, err := rm.db.Exec(`INSERT OR IGNORE INTO calls(caller, callee, file, line) VALUES(?, ?, ?, ?)`,
		c.Caller, c.Callee, c.File, c.Line)
	return err
}

// GetCallers returns callers of a callee.
func (rm *RepoMemory) GetCallers(callee string) ([]Call, error) {
	rows, err := rm.db.Query(`SELECT caller, file, line FROM calls WHERE callee = ?`, callee)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cs []Call
	for rows.Next() {
		var c Call
		c.Callee = callee
		if err := rows.Scan(&c.Caller, &c.File, &c.Line); err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, nil
}

// Close releases resources.
func (rm *RepoMemory) Close() error {
	return rm.db.Close()
}
