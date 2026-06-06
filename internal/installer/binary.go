package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const FallbackBinaryDir = "/usr/local/bin"

type BinaryCheckResult struct {
	Exists      bool
	Executable  bool
	Version     string
	Path        string
	InPATH      bool
	IsBuildable bool
}

func CheckBinary() BinaryCheckResult {
	r := BinaryCheckResult{}

	// Check default path first
	defaultPath := DefaultBinaryPath()
	r.Path = defaultPath
	r.Exists = Exists(defaultPath)
	if r.Exists {
		fi, err := os.Stat(defaultPath)
		if err == nil {
			r.Executable = (fi.Mode()&0111 != 0)
		}
		r.Version = getBinaryVersion(defaultPath)
	}

	// Also check fallback
	fallbackPath := filepath.Join(FallbackBinaryDir, "costaffective")
	if !r.Exists && Exists(fallbackPath) {
		r.Path = fallbackPath
		r.Exists = true
		fi, err := os.Stat(fallbackPath)
		if err == nil {
			r.Executable = (fi.Mode()&0111 != 0)
		}
		r.Version = getBinaryVersion(fallbackPath)
	}

	// Try to find in PATH as last resort
	if !r.Exists {
		if path, err := exec.LookPath("costaffective"); err == nil {
			r.Path = path
			r.Exists = true
			r.Executable = true
			r.InPATH = true
			r.Version = getBinaryVersion(path)
		}
	}

	r.InPATH = isInPATH()
	r.IsBuildable = canBuild()
	return r
}

func InstallBinary() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get working directory: %w", err)
	}
	root := findGoModRoot(cwd)
	if root == "" {
		return "", fmt.Errorf("cannot find go.mod — not inside a Go module")
	}

	outputPath := DefaultBinaryPath()
	installDir := filepath.Dir(outputPath)

	if err := os.MkdirAll(installDir, 0755); err != nil {
		// Fallback to /usr/local/bin
		outputPath = filepath.Join(FallbackBinaryDir, "costaffective")
		installDir = filepath.Dir(outputPath)
		if err := os.MkdirAll(installDir, 0755); err != nil {
			return "", ActionableError{
				Message: fmt.Sprintf("CostAffective was not installed to %s.", DefaultBinaryPath()),
				Action:  fmt.Sprintf("mkdir -p %s\nor rerun:\n  costaffective install --repair", installDir),
			}
		}
	}

	cmd := exec.Command("go", "build", "-o", outputPath, "./cmd/mycli/")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build failed: %w", err)
	}

	if err := os.Chmod(outputPath, 0755); err != nil {
		return "", fmt.Errorf("cannot set executable permissions on %s: %w", outputPath, err)
	}

	if err := VerifyBinary(outputPath); err != nil {
		return "", err
	}

	SetBinaryPath(outputPath)
	return outputPath, nil
}

func VerifyBinary(binaryPath string) error {
	if !Exists(binaryPath) {
		return ActionableError{
			Message: fmt.Sprintf("CostAffective was not found at %s.", binaryPath),
			Action:  "Rerun: costaffective install --repair",
		}
	}

	fi, err := os.Stat(binaryPath)
	if err != nil {
		return ActionableError{
			Message: fmt.Sprintf("Cannot check %s: %s", binaryPath, err),
			Action:  "Check file permissions and ownership, then rerun: costaffective install --repair",
		}
	}
	if fi.Mode()&0111 == 0 {
		return ActionableError{
			Message: fmt.Sprintf("%s exists but is not executable.", binaryPath),
			Action:  fmt.Sprintf("Run: chmod +x %s\nor rerun: costaffective install --repair", binaryPath),
		}
	}

	version := getBinaryVersion(binaryPath)
	if version == "" {
		return ActionableError{
			Message: fmt.Sprintf("%s exists but did not respond to --version.", binaryPath),
			Action:  "The binary may be corrupted. Rerun: costaffective install --repair",
		}
	}

	return nil
}

type ActionableError struct {
	Message string
	Action  string
}

func (e ActionableError) Error() string {
	return fmt.Sprintf("%s\n\n%s", e.Message, e.Action)
}

func getBinaryVersion(binaryPath string) string {
	cmd := exec.Command(binaryPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func isInPATH() bool {
	path, err := exec.LookPath("costaffective")
	if err != nil {
		return false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return abs == DefaultBinaryPath() || strings.HasPrefix(abs, HomeDir())
}

func canBuild() bool {
	if _, err := exec.LookPath("go"); err != nil {
		return false
	}
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	return findGoModRoot(cwd) != ""
}

func findGoModRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func IsDirWritable(dir string) bool {
	tmpfile := filepath.Join(dir, ".costaffective_write_test")
	f, err := os.Create(tmpfile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(tmpfile)
	return true
}

func GetDefaultBinaryDir() string {
	homeBin := filepath.Join(HomeDir(), ".local", "bin")
	if Exists(homeBin) || os.MkdirAll(homeBin, 0755) == nil {
		return homeBin
	}
	return FallbackBinaryDir
}

func GetBinaryCandidates() []string {
	defaultPath := DefaultBinaryPath()
	fallbackPath := filepath.Join(FallbackBinaryDir, "costaffective")

	seen := make(map[string]bool)
	var candidates []string
	for _, p := range []string{defaultPath, fallbackPath} {
		if !seen[p] {
			seen[p] = true
			candidates = append(candidates, p)
		}
	}
	if runtime.GOOS == "linux" {
		if path, err := exec.LookPath("costaffective"); err == nil {
			if !seen[path] {
				candidates = append(candidates, path)
			}
		}
	}
	return candidates
}
