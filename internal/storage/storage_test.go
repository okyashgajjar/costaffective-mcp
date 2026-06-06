package storage

import (
	"context"
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

func TestNewStore(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()
}

func TestSaveAndGetBenchmarkRecord(t *testing.T) {
	store := setupTestStore(t)

	record := &BenchmarkRecord{
		Task:      "test-task",
		Model:     "gpt-4",
		TokensIn:  100,
		TokensOut: 50,
		LatencyMs: 1500,
		Cost:      0.002,
		Provider:  "openai",
		Timestamp: time.Now(),
	}

	err := store.SaveBenchmarkRecord(context.Background(), record)
	if err != nil {
		t.Fatalf("failed to save benchmark record: %v", err)
	}

	records, err := store.GetBenchmarkRecords(context.Background(), 10)
	if err != nil {
		t.Fatalf("failed to get benchmark records: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	got := records[0]
	if got.Task != record.Task {
		t.Errorf("expected task %s, got %s", record.Task, got.Task)
	}
	if got.Model != record.Model {
		t.Errorf("expected model %s, got %s", record.Model, got.Model)
	}
	if got.TokensIn != record.TokensIn {
		t.Errorf("expected tokens_in %d, got %d", record.TokensIn, got.TokensIn)
	}
	if got.TokensOut != record.TokensOut {
		t.Errorf("expected tokens_out %d, got %d", record.TokensOut, got.TokensOut)
	}
	if got.LatencyMs != record.LatencyMs {
		t.Errorf("expected latency_ms %d, got %d", record.LatencyMs, got.LatencyMs)
	}
	if got.Cost != record.Cost {
		t.Errorf("expected cost %f, got %f", record.Cost, got.Cost)
	}
	if got.Provider != record.Provider {
		t.Errorf("expected provider %s, got %s", record.Provider, got.Provider)
	}
}

func TestSaveMultipleRecords(t *testing.T) {
	store := setupTestStore(t)

	for i := 0; i < 5; i++ {
		record := &BenchmarkRecord{
			Task:      "task",
			Model:     "gpt-4",
			TokensIn:  10 * i,
			TokensOut: 5 * i,
			LatencyMs: int64(100 * i),
			Cost:      0.001 * float64(i),
			Provider:  "openai",
			Timestamp: time.Now(),
		}
		if err := store.SaveBenchmarkRecord(context.Background(), record); err != nil {
			t.Fatalf("failed to save record %d: %v", i, err)
		}
	}

	records, err := store.GetBenchmarkRecords(context.Background(), 10)
	if err != nil {
		t.Fatalf("failed to get benchmark records: %v", err)
	}

	if len(records) != 5 {
		t.Errorf("expected 5 records, got %d", len(records))
	}
}

func TestGetLimit(t *testing.T) {
	store := setupTestStore(t)

	for i := 0; i < 10; i++ {
		record := &BenchmarkRecord{
			Task:     "task",
			Model:    "gpt-4",
			TokensIn: 10, TokensOut: 5,
			LatencyMs: 100, Cost: 0.001,
			Provider:  "openai",
			Timestamp: time.Now(),
		}
		if err := store.SaveBenchmarkRecord(context.Background(), record); err != nil {
			t.Fatalf("failed to save record %d: %v", i, err)
		}
	}

	records, err := store.GetBenchmarkRecords(context.Background(), 3)
	if err != nil {
		t.Fatalf("failed to get benchmark records: %v", err)
	}

	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}

func TestGetEmptyStore(t *testing.T) {
	store := setupTestStore(t)

	records, err := store.GetBenchmarkRecords(context.Background(), 10)
	if err != nil {
		t.Fatalf("failed to get benchmark records: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}
