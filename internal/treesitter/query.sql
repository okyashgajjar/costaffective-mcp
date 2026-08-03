-- name: GetSymbolCount :one
SELECT COUNT(*) FROM symbols;

-- name: GetCallEdgeCount :one
SELECT COUNT(*) FROM call_edges;

-- name: GetReferenceCount :one
SELECT COUNT(*) FROM references_t;

-- name: GetFileHash :one
SELECT file_hash FROM symbol_files WHERE file_path = ?;

-- name: GetAllFiles :many
SELECT file_path FROM symbol_files ORDER BY file_path;

-- name: GetFilesByHash :many
SELECT file_path, file_hash FROM symbol_files;

-- name: GetSymbolRange :one
SELECT start_line, end_line FROM symbols
WHERE name = ? AND file = ?
ORDER BY start_line ASC LIMIT 1;

-- name: GetFileSymbols :many
SELECT name, kind, language, file, start_line, end_line, signature
FROM symbols WHERE file = ?
ORDER BY start_line ASC;

-- name: ClearFile :exec
DELETE FROM symbols WHERE file = ?;

-- name: ClearFileReferences :exec
DELETE FROM references_t WHERE file = ?;

-- name: ClearFileCallEdges :exec
DELETE FROM call_edges WHERE file = ?;

-- name: SearchReferences :many
SELECT symbol_name, file, line, col, ref_type, context
FROM references_t
WHERE symbol_name = ?
ORDER BY ref_type, file, line;

-- name: SearchReferencesLike :many
SELECT symbol_name, file, line, col, ref_type, context
FROM references_t
WHERE symbol_name LIKE ?
ORDER BY ref_type, file, line;

-- name: SearchCallEdgesByCaller :many
SELECT id, caller_name, caller_file, callee_name, file, line, language
FROM call_edges
WHERE caller_name = ?
ORDER BY callee_name, file, line;

-- name: SearchCallEdges :many
SELECT id, caller_name, caller_file, callee_name, file, line, language
FROM call_edges
WHERE callee_name = ?
ORDER BY caller_name, file, line;

-- name: SearchCallEdgesLike :many
SELECT id, caller_name, caller_file, callee_name, file, line, language
FROM call_edges
WHERE callee_name LIKE ?
ORDER BY caller_name, file, line;

-- name: SearchByKind :many
SELECT name, kind, language, file, start_line, end_line, signature
FROM symbols WHERE kind = ?
ORDER BY file, start_line ASC
LIMIT ?;

-- name: Search :many
SELECT name, kind, language, file, start_line, end_line, signature
FROM symbols WHERE (
	name LIKE ?1 OR
	name LIKE ?2 OR
	signature LIKE ?3
)
ORDER BY
	CASE
		WHEN name LIKE ?4 THEN 0
		WHEN name LIKE ?5 THEN 1
		ELSE 2
	END,
	start_line ASC
LIMIT ?6;

-- name: InsertSymbol :exec
INSERT OR REPLACE INTO symbols (id, name, kind, language, file, start_line, end_line, signature, content, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP);

-- name: InsertSymbolFile :exec
INSERT OR REPLACE INTO symbol_files (file_path, file_hash, last_indexed)
VALUES (?, ?, ?);

-- name: InsertReference :exec
INSERT INTO references_t (symbol_name, file, line, col, ref_type, context)
VALUES (?, ?, ?, ?, ?, ?);

-- name: InsertCallEdge :exec
INSERT INTO call_edges (caller_name, caller_file, callee_name, file, line, language)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetSchemaVersion :one
SELECT version FROM schema_version;

-- name: DeleteSchemaVersion :exec
DELETE FROM schema_version;

-- name: InsertSchemaVersion :exec
INSERT INTO schema_version (version) VALUES (?);

-- name: DeleteSymbols :exec
DELETE FROM symbols;

-- name: DeleteSymbolFiles :exec
DELETE FROM symbol_files;

-- name: DeleteReferences :exec
DELETE FROM references_t;

-- name: DeleteCallEdges :exec
DELETE FROM call_edges;
