package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()

	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() {
		store.Close()
	})

	return store
}

func createTestSession() *Session {
	return &Session{
		ID:        "test-session-1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages: []Message{
			{
				Role:      "user",
				Content:   "Hello",
				Timestamp: time.Now(),
			},
			{
				Role:      "assistant",
				Content:   "Hi there!",
				Timestamp: time.Now(),
				TokenUsage: &TokenUsage{
					PromptTokens:     10,
					CompletionTokens: 20,
					TotalTokens:      30,
				},
			},
		},
	}
}

func TestNewStore(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Verify database file was created
	dbPath := filepath.Join(tmpDir, "sessions.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("database file was not created at %s", dbPath)
	}
}

func TestSaveAndGetSession(t *testing.T) {
	store := setupTestStore(t)
	session := createTestSession()

	err := store.SaveSession(context.Background(), session)
	if err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	retrieved, err := store.GetSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}

	if retrieved.ID != session.ID {
		t.Errorf("expected session ID %s, got %s", session.ID, retrieved.ID)
	}

	if len(retrieved.Messages) != len(session.Messages) {
		t.Fatalf("expected %d messages, got %d", len(session.Messages), len(retrieved.Messages))
	}

	for i, msg := range retrieved.Messages {
		if msg.Role != session.Messages[i].Role {
			t.Errorf("message %d: expected role %s, got %s", i, session.Messages[i].Role, msg.Role)
		}
		if msg.Content != session.Messages[i].Content {
			t.Errorf("message %d: expected content %s, got %s", i, session.Messages[i].Content, msg.Content)
		}

		if msg.TokenUsage == nil && session.Messages[i].TokenUsage != nil {
			t.Errorf("message %d: expected token usage", i)
		} else if msg.TokenUsage != nil && session.Messages[i].TokenUsage == nil {
			t.Errorf("message %d: unexpected token usage", i)
		} else if msg.TokenUsage != nil {
			if msg.TokenUsage.TotalTokens != session.Messages[i].TokenUsage.TotalTokens {
				t.Errorf("message %d: expected %d total tokens, got %d", i, session.Messages[i].TokenUsage.TotalTokens, msg.TokenUsage.TotalTokens)
			}
		}
	}
}

func TestGetNonExistentSession(t *testing.T) {
	store := setupTestStore(t)

	_, err := store.GetSession(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent session")
	}
}

func TestListSessions(t *testing.T) {
	store := setupTestStore(t)

	// Create two sessions
	session1 := createTestSession()
	session1.ID = "session-1"
	if err := store.SaveSession(context.Background(), session1); err != nil {
		t.Fatalf("failed to save session 1: %v", err)
	}

	session2 := createTestSession()
	session2.ID = "session-2"
	if err := store.SaveSession(context.Background(), session2); err != nil {
		t.Fatalf("failed to save session 2: %v", err)
	}

	sessions, err := store.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
}

func TestDeleteSession(t *testing.T) {
	store := setupTestStore(t)
	session := createTestSession()

	if err := store.SaveSession(context.Background(), session); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	if err := store.DeleteSession(context.Background(), session.ID); err != nil {
		t.Fatalf("failed to delete session: %v", err)
	}

	// Verify session was deleted
	_, err := store.GetSession(context.Background(), session.ID)
	if err == nil {
		t.Error("expected error for deleted session")
	}
}

func TestDeleteNonExistentSession(t *testing.T) {
	store := setupTestStore(t)

	err := store.DeleteSession(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for deleting non-existent session")
	}
}

func TestConcurrentAccess(t *testing.T) {
	store := setupTestStore(t)

	// Test concurrent saves
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(id int) {
			session := createTestSession()
			session.ID = "concurrent-" + string(rune('0'+id))
			if err := store.SaveSession(context.Background(), session); err != nil {
				t.Errorf("failed to save session concurrently: %v", err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	sessions, err := store.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}
	if len(sessions) != 5 {
		t.Errorf("expected 5 sessions, got %d", len(sessions))
	}
}
