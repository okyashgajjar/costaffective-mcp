package retrieval

import (
	"path/filepath"
	"sort"
	"strings"
)

// FilterResults filters out retrieval-poisoning paths/filenames and enforces score thresholds.
func FilterResults(results []RetrievalResult, minScore float64, maxResults int) []RetrievalResult {
	var filtered []RetrievalResult

	// Excluded directory paths (checks if path contains these segments)
	excludedDirs := []string{
		"tests/fixtures/", "/tests/fixtures/",
		"fixtures/", "/fixtures/",
		"testdata/", "/testdata/",
		"examples/", "/examples/",
		"samples/", "/samples/",
		"benchmarks/", "/benchmarks/",
		"reports/", "/reports/",
		"coverage/", "/coverage/",
		".cache/", "/.cache/",
		".git/", "/.git/",
		"node_modules/", "/node_modules/",
		"vendor/", "/vendor/",
		"dist/", "/dist/",
		"build/", "/build/",
		"tmp/", "/tmp/",
		"logs/", "/logs/",
	}

	for _, res := range results {
		// Enforce minimum score/confidence threshold
		if res.Score < minScore {
			continue
		}

		pathLower := strings.ToLower(res.File)
		base := filepath.Base(pathLower)

		// Exclude retrieval-poisoning directory paths
		isExcludedDir := false
		for _, dir := range excludedDirs {
			if strings.Contains(pathLower, strings.ToLower(dir)) {
				isExcludedDir = true
				break
			}
		}
		if isExcludedDir {
			continue
		}

		// Exclude filename patterns
		// - *chat-history*
		// - *.gold.*
		// - *.snapshot.*
		// - *.log
		// - *benchmark-report*
		if strings.Contains(base, "chat-history") ||
			strings.Contains(base, ".gold.") ||
			strings.Contains(base, ".snapshot.") ||
			strings.HasSuffix(base, ".log") ||
			strings.Contains(base, "benchmark-report") {
			continue
		}

		filtered = append(filtered, res)
	}

	// Sort by Score descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	// Cap to maxResults if maxResults > 0
	if maxResults > 0 && len(filtered) > maxResults {
		filtered = filtered[:maxResults]
	}

	return filtered
}
