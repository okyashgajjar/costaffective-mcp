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
