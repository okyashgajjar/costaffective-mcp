package retrieval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/okyashgajjar/costwise-mcp/internal/lsif"
	"github.com/okyashgajjar/costwise-mcp/internal/repository"
)

type LSIFRetriever struct {
	repo  *repository.RepositoryInfo
	index *lsif.Index
}

func NewLSIFRetriever() *LSIFRetriever {
	return &LSIFRetriever{}
}

func (r *LSIFRetriever) Name() string {
	return "lsif"
}

func (r *LSIFRetriever) Initialize(ctx context.Context, repo *repository.RepositoryInfo) error {
	r.repo = repo
	dumpPath := filepath.Join(repo.Root, "dump.lsif")
	if _, err := os.Stat(dumpPath); err != nil {
		return fmt.Errorf("no dump.lsif found in repo root")
	}

	idx, err := lsif.Parse(dumpPath)
	if err != nil {
		return fmt.Errorf("failed to parse dump.lsif: %w", err)
	}
	r.index = idx
	return nil
}

func (r *LSIFRetriever) Retrieve(ctx context.Context, query string) ([]RetrievalResult, error) {
	return nil, fmt.Errorf("LSIF retriever requires specific symbol resolution, not raw string query")
}

func (r *LSIFRetriever) FindReferences(symbol string, filepath string, line int, character int) ([]RetrievalResult, error) {
	if r.index == nil {
		return nil, fmt.Errorf("LSIF index not loaded")
	}

	// Find the range that matches the symbol at the given location
	var targetResultSetID int = -1

	for rangeID, loc := range r.index.Ranges {
		if loc.URI == filepath && loc.Range.Start.Line == line {
			// For simplicity, we just match line. In reality we'd check character.
			if rsID, ok := r.index.NextMap[rangeID]; ok {
				targetResultSetID = rsID
				break
			}
		}
	}

	if targetResultSetID == -1 {
		return nil, fmt.Errorf("symbol not found in LSIF dump")
	}

	var results []RetrievalResult
	if refRangeIDs, ok := r.index.References[targetResultSetID]; ok {
		for _, refID := range refRangeIDs {
			if loc, ok := r.index.Ranges[refID]; ok {
				results = append(results, RetrievalResult{
					File:      loc.URI,
					Score:     1.0,
					MatchHits: 1,
				})
			}
		}
	}

	return results, nil
}

func (r *LSIFRetriever) Metrics() RetrievalMetrics {
	return RetrievalMetrics{}
}

func (r *LSIFRetriever) Shutdown() error {
	r.index = nil // Free memory
	return nil
}
