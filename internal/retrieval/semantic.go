package retrieval

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/okyashgajjar/costwise-mcp/internal/index"
	"github.com/okyashgajjar/costwise-mcp/internal/repository"
)

type SemanticRetriever struct {
	repo     *repository.RepositoryInfo
	semantic *index.SemanticIndexer
}

func NewSemanticRetriever() *SemanticRetriever {
	return &SemanticRetriever{}
}

func (r *SemanticRetriever) Name() string {
	return "semantic"
}

func (r *SemanticRetriever) Initialize(ctx context.Context, repo *repository.RepositoryInfo) error {
	r.repo = repo
	semIdx, err := index.NewSemanticIndexer(filepath.Join(repo.Root, ".mycli-fts"))
	if err != nil {
		return fmt.Errorf("failed to open semantic index: %w", err)
	}
	r.semantic = semIdx
	return nil
}

func (r *SemanticRetriever) Retrieve(ctx context.Context, query string) ([]RetrievalResult, error) {
	if r.semantic == nil {
		return nil, fmt.Errorf("semantic index not initialized")
	}
	// Limit to top 10 results
	semResults, err := r.semantic.Search(query, 10)
	if err != nil {
		return nil, err
	}

	var results []RetrievalResult
	for _, sr := range semResults {
		res := RetrievalResult{
			File:      sr.Path,
			Score:     sr.Score,
			MatchHits: 1,
		}
		// Try to extract some snippets or context if needed, but for now we rely on the summary logic.
		results = append(results, res)
	}
	return results, nil
}

func (r *SemanticRetriever) Metrics() RetrievalMetrics {
	return RetrievalMetrics{} // Stub for now
}

func (r *SemanticRetriever) Shutdown() error {
	if r.semantic != nil {
		return r.semantic.Close()
	}
	return nil
}
