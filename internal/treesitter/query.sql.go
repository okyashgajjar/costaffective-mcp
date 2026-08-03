package treesitter

import (
	"context"
	"database/sql"
	"time"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

type Queries struct {
	db DBTX
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db: tx,
	}
}

// sqlc generated models (simplified to match existing usage)
type SearchParams struct {
	Name      string
	Name_2    string
	Signature string
	Name_3    string
	Name_4    string
	Limit     int64
}

type SearchRow struct {
	Name      string
	Kind      string
	Language  string
	File      string
	StartLine int64
	EndLine   int64
	Signature string
}

func (q *Queries) Search(ctx context.Context, arg SearchParams) ([]SearchRow, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT name, kind, language, file, start_line, end_line, signature
		FROM symbols WHERE (
			name LIKE ? OR
			name LIKE ? OR
			signature LIKE ?
		)
		ORDER BY
			CASE
				WHEN name LIKE ? THEN 0
				WHEN name LIKE ? THEN 1
				ELSE 2
			END,
			start_line ASC
		LIMIT ?
	`, arg.Name, arg.Name_2, arg.Signature, arg.Name_3, arg.Name_4, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchRow
	for rows.Next() {
		var i SearchRow
		if err := rows.Scan(
			&i.Name,
			&i.Kind,
			&i.Language,
			&i.File,
			&i.StartLine,
			&i.EndLine,
			&i.Signature,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type InsertSymbolParams struct {
	ID        string
	Name      string
	Kind      string
	Language  string
	File      string
	StartLine int64
	EndLine   int64
	Signature string
	Content   string
}

func (q *Queries) InsertSymbol(ctx context.Context, arg InsertSymbolParams) error {
	_, err := q.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO symbols (id, name, kind, language, file, start_line, end_line, signature, content, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, arg.ID, arg.Name, arg.Kind, arg.Language, arg.File, arg.StartLine, arg.EndLine, arg.Signature, arg.Content)
	return err
}

type SearchByKindParams struct {
	Kind  string
	Limit int64
}

func (q *Queries) SearchByKind(ctx context.Context, arg SearchByKindParams) ([]SearchRow, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT name, kind, language, file, start_line, end_line, signature
		FROM symbols WHERE kind = ?
		ORDER BY file, start_line ASC
		LIMIT ?
	`, arg.Kind, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchRow
	for rows.Next() {
		var i SearchRow
		if err := rows.Scan(
			&i.Name,
			&i.Kind,
			&i.Language,
			&i.File,
			&i.StartLine,
			&i.EndLine,
			&i.Signature,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) GetSymbolCount(ctx context.Context) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM symbols").Scan(&count)
	return count, err
}

type GetSymbolRangeParams struct {
	Name string
	File string
}

type GetSymbolRangeRow struct {
	StartLine int64
	EndLine   int64
}

func (q *Queries) GetSymbolRange(ctx context.Context, arg GetSymbolRangeParams) (GetSymbolRangeRow, error) {
	var i GetSymbolRangeRow
	err := q.db.QueryRowContext(ctx, `
		SELECT start_line, end_line FROM symbols
		WHERE name = ? AND file = ?
		ORDER BY start_line ASC LIMIT 1
	`, arg.Name, arg.File).Scan(&i.StartLine, &i.EndLine)
	return i, err
}

func (q *Queries) GetFileSymbols(ctx context.Context, file string) ([]SearchRow, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT name, kind, language, file, start_line, end_line, signature
		FROM symbols WHERE file = ?
		ORDER BY start_line ASC
	`, file)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchRow
	for rows.Next() {
		var i SearchRow
		if err := rows.Scan(
			&i.Name,
			&i.Kind,
			&i.Language,
			&i.File,
			&i.StartLine,
			&i.EndLine,
			&i.Signature,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) GetAllFiles(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT file_path FROM symbol_files ORDER BY file_path")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var file_path string
		if err := rows.Scan(&file_path); err != nil {
			return nil, err
		}
		items = append(items, file_path)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type GetFilesByHashRow struct {
	FilePath string
	FileHash string
}

func (q *Queries) GetFilesByHash(ctx context.Context) ([]GetFilesByHashRow, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT file_path, file_hash FROM symbol_files")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFilesByHashRow
	for rows.Next() {
		var i GetFilesByHashRow
		if err := rows.Scan(&i.FilePath, &i.FileHash); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type InsertSymbolFileParams struct {
	FilePath    string
	FileHash    string
	LastIndexed time.Time
}

func (q *Queries) InsertSymbolFile(ctx context.Context, arg InsertSymbolFileParams) error {
	_, err := q.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO symbol_files (file_path, file_hash, last_indexed)
		VALUES (?, ?, ?)
	`, arg.FilePath, arg.FileHash, arg.LastIndexed)
	return err
}

func (q *Queries) GetFileHash(ctx context.Context, filePath string) (string, error) {
	var file_hash string
	err := q.db.QueryRowContext(ctx, "SELECT file_hash FROM symbol_files WHERE file_path = ?", filePath).Scan(&file_hash)
	return file_hash, err
}

func (q *Queries) ClearFile(ctx context.Context, file string) error {
	_, err := q.db.ExecContext(ctx, "DELETE FROM symbols WHERE file = ?", file)
	return err
}

func (q *Queries) ClearFileReferences(ctx context.Context, file string) error {
	_, err := q.db.ExecContext(ctx, "DELETE FROM references_t WHERE file = ?", file)
	return err
}

func (q *Queries) ClearFileCallEdges(ctx context.Context, file string) error {
	_, err := q.db.ExecContext(ctx, "DELETE FROM call_edges WHERE file = ?", file)
	return err
}

type InsertReferenceParams struct {
	SymbolName string
	File       string
	Line       int64
	Col        int64
	RefType    string
	Context    string
}

func (q *Queries) InsertReference(ctx context.Context, arg InsertReferenceParams) error {
	_, err := q.db.ExecContext(ctx, `
		INSERT INTO references_t (symbol_name, file, line, col, ref_type, context)
		VALUES (?, ?, ?, ?, ?, ?)
	`, arg.SymbolName, arg.File, arg.Line, arg.Col, arg.RefType, arg.Context)
	return err
}

type SearchReferencesRow struct {
	SymbolName string
	File       string
	Line       int64
	Col        int64
	RefType    string
	Context    string
}

func (q *Queries) SearchReferences(ctx context.Context, symbolName string) ([]SearchReferencesRow, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT symbol_name, file, line, col, ref_type, context
		FROM references_t
		WHERE symbol_name = ?
		ORDER BY ref_type, file, line
	`, symbolName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchReferencesRow
	for rows.Next() {
		var i SearchReferencesRow
		if err := rows.Scan(
			&i.SymbolName,
			&i.File,
			&i.Line,
			&i.Col,
			&i.RefType,
			&i.Context,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) SearchReferencesLike(ctx context.Context, symbolName string) ([]SearchReferencesRow, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT symbol_name, file, line, col, ref_type, context
		FROM references_t
		WHERE symbol_name LIKE ?
		ORDER BY ref_type, file, line
	`, symbolName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchReferencesRow
	for rows.Next() {
		var i SearchReferencesRow
		if err := rows.Scan(
			&i.SymbolName,
			&i.File,
			&i.Line,
			&i.Col,
			&i.RefType,
			&i.Context,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type InsertCallEdgeParams struct {
	CallerName string
	CallerFile string
	CalleeName string
	File       string
	Line       int64
	Language   string
}

func (q *Queries) InsertCallEdge(ctx context.Context, arg InsertCallEdgeParams) error {
	_, err := q.db.ExecContext(ctx, `
		INSERT INTO call_edges (caller_name, caller_file, callee_name, file, line, language)
		VALUES (?, ?, ?, ?, ?, ?)
	`, arg.CallerName, arg.CallerFile, arg.CalleeName, arg.File, arg.Line, arg.Language)
	return err
}

type SearchCallEdgesRow struct {
	ID         int64
	CallerName string
	CallerFile string
	CalleeName string
	File       string
	Line       int64
	Language   string
}

func (q *Queries) SearchCallEdgesByCaller(ctx context.Context, callerName string) ([]SearchCallEdgesRow, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, caller_name, caller_file, callee_name, file, line, language
		FROM call_edges
		WHERE caller_name = ?
		ORDER BY callee_name, file, line
	`, callerName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchCallEdgesRow
	for rows.Next() {
		var i SearchCallEdgesRow
		if err := rows.Scan(
			&i.ID,
			&i.CallerName,
			&i.CallerFile,
			&i.CalleeName,
			&i.File,
			&i.Line,
			&i.Language,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) SearchCallEdges(ctx context.Context, calleeName string) ([]SearchCallEdgesRow, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, caller_name, caller_file, callee_name, file, line, language
		FROM call_edges
		WHERE callee_name = ?
		ORDER BY caller_name, file, line
	`, calleeName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchCallEdgesRow
	for rows.Next() {
		var i SearchCallEdgesRow
		if err := rows.Scan(
			&i.ID,
			&i.CallerName,
			&i.CallerFile,
			&i.CalleeName,
			&i.File,
			&i.Line,
			&i.Language,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) SearchCallEdgesLike(ctx context.Context, calleeName string) ([]SearchCallEdgesRow, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, caller_name, caller_file, callee_name, file, line, language
		FROM call_edges
		WHERE callee_name LIKE ?
		ORDER BY caller_name, file, line
	`, calleeName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchCallEdgesRow
	for rows.Next() {
		var i SearchCallEdgesRow
		if err := rows.Scan(
			&i.ID,
			&i.CallerName,
			&i.CallerFile,
			&i.CalleeName,
			&i.File,
			&i.Line,
			&i.Language,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) GetCallEdgeCount(ctx context.Context) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM call_edges").Scan(&count)
	return count, err
}

func (q *Queries) GetReferenceCount(ctx context.Context) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM references_t").Scan(&count)
	return count, err
}

type InsertOntologyTagParams struct {
	SymbolID string
	Tag      string
	Domain   string
	File     string
}

func (q *Queries) InsertOntologyTag(ctx context.Context, arg InsertOntologyTagParams) error {
	_, err := q.db.ExecContext(ctx, `
		INSERT INTO ontology_tags (symbol_id, tag, domain, file)
		VALUES (?, ?, ?, ?)
	`, arg.SymbolID, arg.Tag, arg.Domain, arg.File)
	return err
}

type SearchOntologyParams struct {
	Tag    string
	Domain string
}

func (q *Queries) SearchOntology(ctx context.Context, arg SearchOntologyParams) ([]SearchRow, error) {
	query := `
		SELECT s.name, s.kind, s.language, s.file, s.start_line, s.end_line, s.signature
		FROM symbols s
		JOIN ontology_tags t ON s.id = t.symbol_id
		WHERE t.tag = ?
	`
	args := []interface{}{arg.Tag}
	
	if arg.Domain != "" {
		query += " AND t.domain = ?"
		args = append(args, arg.Domain)
	}
	
	query += " ORDER BY s.file, s.start_line"
	
	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchRow
	for rows.Next() {
		var i SearchRow
		if err := rows.Scan(
			&i.Name,
			&i.Kind,
			&i.Language,
			&i.File,
			&i.StartLine,
			&i.EndLine,
			&i.Signature,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) ClearFileOntologyTags(ctx context.Context, file string) error {
	_, err := q.db.ExecContext(ctx, "DELETE FROM ontology_tags WHERE file = ?", file)
	return err
}
