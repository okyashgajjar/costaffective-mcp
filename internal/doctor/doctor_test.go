package doctor

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"costaffective/internal/installer"
	_ "costaffective/internal/installer/targets"
)

func TestDoctorBinaryCheck_PASS(t *testing.T) {
	binPath := buildTempBinary(t)
	installer.SetBinaryPath(binPath)
	defer installer.SetBinaryPath("")

	results := CheckBinary()
	passCount := 0
	failCount := 0
	for _, r := range results {
		if r.Status == PASS {
			passCount++
		}
		if r.Status == FAIL {
			failCount++
		}
	}

	if failCount > 0 {
		t.Fatalf("expected 0 FAIL, got %d. Results: %+v", failCount, results)
	}
	if passCount == 0 {
		t.Fatal("expected at least 1 PASS")
	}
}

func TestDoctorBinaryCheck_verifyNonexistent(t *testing.T) {
	// Test that VerifyBinary produces proper error for nonexistent path
	err := installer.VerifyBinary("/nonexistent/costaffective")
	if err == nil {
		t.Fatal("VerifyBinary should fail for nonexistent path")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error should mention 'not found': %v", err)
	}

	// Test that an installer.ActionableError is returned
	var actionable installer.ActionableError
	if !asActionable(err, &actionable) {
		t.Fatalf("error should be ActionableError: %T %v", err, err)
	}
	if actionable.Action == "" {
		t.Fatal("ActionableError should have non-empty Action")
	}
}

func asActionable(err error, target *installer.ActionableError) bool {
	a, ok := err.(installer.ActionableError)
	if ok {
		*target = a
		return true
	}
	// Check through fmt.Errorf wrapping
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return asActionable(wrapped.Unwrap(), target)
	}
	return false
}

func TestDoctorPATH(t *testing.T) {
	results := CheckPATH()
	if len(results) == 0 {
		t.Fatal("expected at least 1 result from CheckPATH")
	}
}

func TestDoctorMCPConfigs(t *testing.T) {
	binPath := buildTempBinary(t)
	installer.SetBinaryPath(binPath)
	defer installer.SetBinaryPath("")

	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	// Install binary to the default path so config validation passes
	defaultBin := installer.DefaultBinaryPath()
	os.MkdirAll(filepath.Dir(defaultBin), 0755)
	data, _ := os.ReadFile(binPath)
	os.WriteFile(defaultBin, data, 0755)

	cursorDir := filepath.Join(dir, ".cursor")
	os.MkdirAll(cursorDir, 0755)
	cursorConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"costaffective": map[string]interface{}{
				"command": defaultBin,
				"args":    []string{"serve"},
				"type":    "stdio",
			},
		},
	}
	cursorFile := filepath.Join(cursorDir, "mcp.json")
	installer.WriteJSONFile(cursorFile, cursorConfig)

	results := CheckMCPConfigs()
	foundCursor := false
	for _, r := range results {
		if strings.Contains(r.Name, "Cursor") {
			foundCursor = true
			if r.Status != PASS {
				t.Fatalf("Cursor config should PASS, got %s: %s", r.Status, r.Detail)
			}
			break
		}
	}
	if !foundCursor {
		for _, r := range results {
			t.Logf("result: %s %s", r.Status, r.Name)
		}
		t.Fatal("expected Cursor config check result — targets may not be registered")
	}
}

func TestDoctorRepository(t *testing.T) {
	results := CheckRepository()
	hasPass := false
	for _, r := range results {
		if r.Status == PASS {
			hasPass = true
		}
	}
	if !hasPass {
		t.Fatal("expected at least 1 PASS from Repository check")
	}
}

func TestDoctorRunAll(t *testing.T) {
	binPath := buildTempBinary(t)
	installer.SetBinaryPath(binPath)
	defer installer.SetBinaryPath("")

	results := RunAll()
	if len(results) == 0 {
		t.Fatal("RunAll should return at least 1 check")
	}
}

func TestDoctorFinalStatus(t *testing.T) {
	allPass := []CheckResult{
		{Name: "Test1", Status: PASS},
		{Name: "Test2", Status: PASS},
	}
	status, pass, fail := FinalStatus(allPass)
	if status != PASS || pass != 2 || fail != 0 {
		t.Fatalf("all pass: status=%s pass=%d fail=%d", status, pass, fail)
	}

	mixed := []CheckResult{
		{Name: "Test1", Status: PASS},
		{Name: "Test2", Status: FAIL},
	}
	status, pass, fail = FinalStatus(mixed)
	if status != FAIL || pass != 1 || fail != 1 {
		t.Fatalf("mixed: status=%s pass=%d fail=%d", status, pass, fail)
	}
}

func buildTempBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	out := filepath.Join(dir, "costaffective")

	cmd := exec.Command("go", "build", "-o", out, "../../cmd/mycli/")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build temp binary: %v", err)
	}
	return out
}
