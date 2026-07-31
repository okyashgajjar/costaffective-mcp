package cache

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractGitHubSlug(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://github.com/okyashgajjar/costwise-mcp", "okyashgajjar/costwise-mcp"},
		{"git@github.com:okyashgajjar/costwise-mcp.git", "okyashgajjar/costwise-mcp"},
		{"https://gitlab.com/test", ""},
	}

	for _, tt := range tests {
		got := extractGitHubSlug(tt.in)
		if got != tt.want {
			t.Errorf("extractGitHubSlug(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDownloadAndExtract(t *testing.T) {
	// Create a dummy zip file
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, _ := w.Create("test.txt")
	_, _ = f.Write([]byte("hello"))
	w.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(buf.Bytes())
	}))
	defer ts.Close()

	tmpDir, _ := os.MkdirTemp("", "cache-test")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err := downloadAndExtract(context.Background(), ts.URL, "", tmpDir)
	if err != nil {
		t.Fatalf("downloadAndExtract failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
	if err != nil || string(data) != "hello" {
		t.Errorf("Expected 'hello', got %s", data)
	}
}
