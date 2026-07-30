package retrieval

import (
	"context"
)

type TreeSitterRetriever interface {
	Retriever
	ParseAST(ctx context.Context, filePath string) ([]Symbol, error)
	FindDefinitions(ctx context.Context, symbol string) ([]RetrievalResult, error)
	FindReferences(ctx context.Context, symbol string) ([]RetrievalResult, error)
}

type CallGraphRetrieverInterface interface {
	Retriever
	BuildCallGraph(ctx context.Context) error
	GetCallers(ctx context.Context, function string) ([]RetrievalResult, error)
	GetCallees(ctx context.Context, function string) ([]RetrievalResult, error)
}

type ReachabilityRetriever interface {
	Retriever
	BuildReachabilityIndex(ctx context.Context) error
	FindReachable(ctx context.Context, from, to string) ([]RetrievalResult, error)
	AffectedBy(ctx context.Context, symbol string) ([]RetrievalResult, error)
}

type LSIFRetriever interface {
	Retriever
	LoadLSIFIndex(ctx context.Context, indexPath string) error
	FindDefinitions(ctx context.Context, symbol string) ([]RetrievalResult, error)
	FindReferences(ctx context.Context, symbol string) ([]RetrievalResult, error)
	Hover(ctx context.Context, symbol string) (string, error)
}

type SCIPRetriever interface {
	Retriever
	LoadSCIPIndex(ctx context.Context, indexPath string) error
	FindDefinitions(ctx context.Context, symbol string) ([]RetrievalResult, error)
	FindReferences(ctx context.Context, symbol string) ([]RetrievalResult, error)
	DocumentSymbols(ctx context.Context, filePath string) ([]Symbol, error)
}

type DependencyGraphRetriever interface {
	Retriever
	BuildDependencyGraph(ctx context.Context) error
	GetDependencies(ctx context.Context, module string) ([]RetrievalResult, error)
	GetDependents(ctx context.Context, module string) ([]RetrievalResult, error)
}

type SymbolTableRetriever interface {
	Retriever
	BuildSymbolTable(ctx context.Context) error
	LookupSymbol(ctx context.Context, name string) ([]RetrievalResult, error)
	ListSymbols(ctx context.Context, prefix string) ([]Symbol, error)
}

type Symbol struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	FilePath   string `json:"file_path"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	ParentName string `json:"parent_name,omitempty"`
	Signature  string `json:"signature,omitempty"`
}
