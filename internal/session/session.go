package session

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Session represents a chat session.
type Session struct {
	ID           string
	Messages     []Message
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastResolved string
}

// Message represents a chat message in a session.
type Message struct {
	Role       string // e.g., "system", "user", "assistant"
	Content    string
	Timestamp  time.Time
	TokenUsage *TokenUsage
}

// TokenUsage contains token usage information.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Store handles persistence of sessions.
type Store struct {
	db *sql.DB
}

// NewStore creates a new session store with SQLite.
func NewStore(dataDir string) (*Store, error) {
	// Ensure directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "sessions.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// createTables creates the necessary tables if they don't exist.
func createTables(db *sql.DB) error {
	// Sessions table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			last_resolved TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Try to alter existing table in case it was created without last_resolved
	_, _ = db.Exec("ALTER TABLE sessions ADD COLUMN last_resolved TEXT")

	// Messages table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			prompt_tokens INTEGER,
			completion_tokens INTEGER,
			total_tokens INTEGER,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// SaveSession saves or updates a session.
func (s *Store) SaveSession(ctx context.Context, session *Session) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Insert or update session
	res, err := tx.ExecContext(ctx, `
		INSERT OR REPLACE INTO sessions (id, created_at, updated_at, last_resolved)
		VALUES (?, ?, ?, ?)
	`, session.ID, session.CreatedAt, session.UpdatedAt, session.LastResolved)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Check if session was actually affected
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("no rows affected when saving session")
	}

	// Delete existing messages for this session
	if _, err := tx.ExecContext(ctx, "DELETE FROM messages WHERE session_id = ?", session.ID); err != nil {
		return fmt.Errorf("failed to delete existing messages: %w", err)
	}

	// Insert messages
	for _, msg := range session.Messages {
		var promptTokens, completionTokens, totalTokens *int
		if msg.TokenUsage != nil {
			promptTokens = &msg.TokenUsage.PromptTokens
			completionTokens = &msg.TokenUsage.CompletionTokens
			totalTokens = &msg.TokenUsage.TotalTokens
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO messages (session_id, role, content, timestamp, prompt_tokens, completion_tokens, total_tokens)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, session.ID, msg.Role, msg.Content, msg.Timestamp,
			promptTokens, completionTokens, totalTokens)
		if err != nil {
			return fmt.Errorf("failed to insert message: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetSession retrieves a session by ID.
func (s *Store) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	var session Session
	var createdAt, updatedAt time.Time
	var lastResolved sql.NullString

	// Get session
	err := s.db.QueryRowContext(ctx, `
		SELECT id, created_at, updated_at, last_resolved
		FROM sessions
		WHERE id = ?
	`, sessionID).Scan(&session.ID, &createdAt, &updatedAt, &lastResolved)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	session.CreatedAt = createdAt
	session.UpdatedAt = updatedAt
	if lastResolved.Valid {
		session.LastResolved = lastResolved.String
	}

	// Get messages
	rows, err := s.db.QueryContext(ctx, `
		SELECT role, content, timestamp, prompt_tokens, completion_tokens, total_tokens
		FROM messages
		WHERE session_id = ?
		ORDER BY timestamp ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg Message
		var timestamp time.Time
		var promptTokens, completionTokens, totalTokens sql.NullInt64

		if err := rows.Scan(&msg.Role, &msg.Content, &timestamp, &promptTokens, &completionTokens, &totalTokens); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Timestamp = timestamp
		if promptTokens.Valid {
			msg.TokenUsage = &TokenUsage{
				PromptTokens:     int(promptTokens.Int64),
				CompletionTokens: int(completionTokens.Int64),
				TotalTokens:      int(totalTokens.Int64),
			}
		}

		session.Messages = append(session.Messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return &session, nil
}

// ListSessions returns a list of session IDs ordered by most recently updated.
func (s *Store) ListSessions(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id
		FROM sessions
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessionIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan session ID: %w", err)
		}
		sessionIDs = append(sessionIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessionIDs, nil
}

// DeleteSession deletes a session and its messages.
func (s *Store) DeleteSession(ctx context.Context, sessionID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete messages first (due to foreign key constraint)
	if _, err := tx.ExecContext(ctx, "DELETE FROM messages WHERE session_id = ?", sessionID); err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	// Delete session
	res, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Check if session was actually deleted
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
