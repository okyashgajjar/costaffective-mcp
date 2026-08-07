package impact

import (
	"github.com/okyashgajjar/costwise-mcp/internal/treesitter"
)

type ImpactAnalysis struct {
	SymbolName       string                  `json:"symbol_name"`
	DirectReferences []treesitter.Reference  `json:"direct_references"`
	Callers          []treesitter.CallEdge   `json:"callers"`
	Callees          []treesitter.CallEdge   `json:"callees"`
	Complexity       string                  `json:"complexity"`
}

func Analyze(db *treesitter.SymbolDB, symbolName string) (*ImpactAnalysis, error) {
	refs, err := db.SearchReferences(symbolName)
	if err != nil {
		return nil, err
	}

	// Find places that call this symbol
	callers, err := db.SearchCallEdges(symbolName)
	if err != nil {
		return nil, err
	}

	// Find things this symbol calls
	callees, err := db.SearchCallEdgesByCaller(symbolName)
	if err != nil {
		return nil, err
	}

	complexity := "Low"
	totalImpact := len(refs) + len(callers)
	if totalImpact > 10 {
		complexity = "High"
	} else if totalImpact > 3 {
		complexity = "Medium"
	}

	return &ImpactAnalysis{
		SymbolName:       symbolName,
		DirectReferences: refs,
		Callers:          callers,
		Callees:          callees,
		Complexity:       complexity,
	}, nil
}
