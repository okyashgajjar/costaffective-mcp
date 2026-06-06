package classifier

import (
	"testing"
)

func TestClassifySymbolQuery(t *testing.T) {
	tests := []struct {
		query string
		class QueryClass
	}{
		{"Find RepositoryManager", SymbolQuery},
		{"where is ParseLevel implemented", SymbolQuery},
		{"Show the NewBuilder function", SymbolQuery},
		{"RepositoryManager struct", SymbolQuery},
		{"BuildContext method", SymbolQuery},
		{"SymbolRetriever", SymbolQuery},
		{"implemented methods in Retriever interface", SymbolQuery},
	}

	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query)
			if result.Class != tc.class {
				t.Errorf("Classify(%q) = %v (confidence=%.2f, reason=%s), want %v",
					tc.query, result.Label, result.Confidence, result.Reason, tc.class)
			}
		})
	}
}

func TestClassifyTextQuery(t *testing.T) {
	tests := []struct {
		query string
		class QueryClass
	}{
		{"how does authentication work", TextQuery},
		{"explain the login flow", TextQuery},
		{"find how config is loaded", TextQuery},
		{"how to implement a new provider", TextQuery},
		{"how error handling works", TextQuery},
		{"find where database connections are configured", TextQuery},
	}

	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query)
			if result.Class != tc.class {
				t.Errorf("Classify(%q) = %v (confidence=%.2f, reason=%s), want %v",
					tc.query, result.Label, result.Confidence, result.Reason, tc.class)
			}
		})
	}
}

func TestClassifyRepoQuery(t *testing.T) {
	tests := []struct {
		query string
		class QueryClass
	}{
		{"about", RepositoryQuery},
		{"purpose", RepositoryQuery},
		{"overview", RepositoryQuery},
		{"what is this project about", RepositoryQuery},
		{"project overview", RepositoryQuery},
		{"architecture overview", RepositoryQuery},
		{"project description", RepositoryQuery},
	}

	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query)
			if result.Class != tc.class {
				t.Errorf("Classify(%q) = %v (confidence=%.2f, reason=%s), want %v",
					tc.query, result.Label, result.Confidence, result.Reason, tc.class)
			}
		})
	}
}

func TestClassifyConfidenceRange(t *testing.T) {
	result := Classify("Find RepositoryManager struct")
	if result.Confidence < 0.3 || result.Confidence > 1.0 {
		t.Errorf("confidence out of range [0,1]: %.3f", result.Confidence)
	}

	result = Classify("how does authentication work")
	if result.Confidence < 0.3 || result.Confidence > 1.0 {
		t.Errorf("confidence out of range [0,1]: %.3f", result.Confidence)
	}

	result = Classify("")
	if result.Confidence < 0 || result.Confidence > 1.0 {
		t.Errorf("confidence out of range [0,1]: %.3f", result.Confidence)
	}
}

func TestClassifyCallQuery(t *testing.T) {
	tests := []struct {
		query string
		class QueryClass
	}{
		{"Who calls RepoMap", CallQuery},
		{"who calls send_chat", CallQuery},
		{"find callers of ModelSettings", CallQuery},
		{"show callers of OpenAIChatCompletion", CallQuery},
		{"show call sites of build_prompt", CallQuery},
	}

	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query)
			if result.Class != tc.class {
				t.Errorf("Classify(%q) = %v (confidence=%.2f, reason=%s), want %v",
					tc.query, result.Label, result.Confidence, result.Reason, tc.class)
			}
		})
	}
}

func TestClassifyLabelMatch(t *testing.T) {
	result := Classify("SymbolRetriever")
	if result.Label != "symbol" {
		t.Errorf("expected label 'symbol', got %q", result.Label)
	}

	result = Classify("how does it work")
	if result.Label != "text" {
		t.Errorf("expected label 'text', got %q", result.Label)
	}

	result = Classify("about")
	if result.Label != "repository" {
		t.Errorf("expected label 'repository', got %q", result.Label)
	}
}
