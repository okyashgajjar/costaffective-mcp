package ledger

import (
	"testing"
)

func TestTypedLedgerAppend(t *testing.T) {
	defer CloseAll()
	evt := StashCreateEvent{Handle: "abc", Tokens: 100, Summary: "test"}
	err := AppendTyped(t.TempDir(), evt)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
