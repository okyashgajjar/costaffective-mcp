package cache

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func AttemptRemoteFetch(ctx context.Context, repoRoot string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// First check if a cache already exists
	cacheDir := filepath.Join(repoRoot, ".mycli-fts")
	if _, err := os.Stat(cacheDir); err == nil {
		return nil // Cache already exists locally, skip
	}

	// 1. Get git remote URL
	cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get git remote: %w", err)
	}

	repoSlug := extractGitHubSlug(string(out))
	if repoSlug == "" {
		return fmt.Errorf("not a GitHub repository or remote not origin")
	}

	// 2. Query GitHub API for the latest run of "costwise-cache" artifact
	// Note: For private repositories, this requires a GITHUB_TOKEN.
	token := os.Getenv("GITHUB_TOKEN")
	artifactURL, err := fetchLatestArtifactURL(ctx, repoSlug, token)
	if err != nil {
		return fmt.Errorf("failed to fetch artifact url: %w", err)
	}

	// 3. Download and extract
	if err := downloadAndExtract(ctx, artifactURL, token, cacheDir); err != nil {
		return fmt.Errorf("failed to download and extract cache: %w", err)
	}

	return nil
}

func extractGitHubSlug(remoteURL string) string {
	remoteURL = strings.TrimSpace(remoteURL)
	if strings.HasPrefix(remoteURL, "https://github.com/") {
		return strings.TrimSuffix(strings.TrimPrefix(remoteURL, "https://github.com/"), ".git")
	}
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		return strings.TrimSuffix(strings.TrimPrefix(remoteURL, "git@github.com:"), ".git")
	}
	return ""
}

func fetchLatestArtifactURL(ctx context.Context, repoSlug, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/artifacts", repoSlug)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var result struct {
		Artifacts []struct {
			Name               string `json:"name"`
			ArchiveDownloadUrl string `json:"archive_download_url"`
		} `json:"artifacts"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	for _, artifact := range result.Artifacts {
		if artifact.Name == "costwise-cache" {
			return artifact.ArchiveDownloadUrl, nil
		}
	}

	return "", fmt.Errorf("no 'costwise-cache' artifact found")
}

func downloadAndExtract(ctx context.Context, url, token, destDir string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("artifact download failed with status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(bodyBytes), int64(len(bodyBytes)))
	if err != nil {
		return err
	}

	_ = os.MkdirAll(destDir, 0755)

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue // Prevent ZipSlip
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
