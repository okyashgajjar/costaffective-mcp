package answertype

import (
	"testing"
)

func TestClassifyYesNo(t *testing.T) {
	tests := []struct {
		query string
		atype AnswerType
	}{
		{"Does this repo have tests?", YesNo},
		{"Is there a config file?", YesNo},
		{"Does it support streaming?", YesNo},
		{"Can this run on Windows?", YesNo},
		{"Are there any tests?", YesNo},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query, "chat")
			if result.Type != tc.atype {
				t.Errorf("Classify(%q) = %s (conf=%.2f), want %s", tc.query, result.Label, result.Confidence, tc.atype)
			}
		})
	}
}

func TestClassifyLocation(t *testing.T) {
	tests := []struct {
		query string
		atype AnswerType
	}{
		{"Where is RepoMap implemented?", Location},
		{"Where are the tests?", Location},
		{"Which file defines SymbolRetriever?", Location},
		{"Find the location of ParseLevel", Location},
		{"File path of KnowledgeStore", Location},
		{"RepoMap", Location},
		{"SymbolRetriever", Location},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query, "chat")
			if result.Type != tc.atype {
				t.Errorf("Classify(%q) = %s (conf=%.2f), want %s", tc.query, result.Label, result.Confidence, tc.atype)
			}
		})
	}
}

func TestClassifyCaller(t *testing.T) {
	tests := []struct {
		query string
		atype AnswerType
	}{
		{"Who calls RepoMap?", Caller},
		{"Who uses KnowledgeStore?", Caller},
		{"Find callers of ParseLevel", Caller},
		{"Show callers of sendMessage", Caller},
		{"What calls ExecutePlan?", Caller},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query, "chat")
			if result.Type != tc.atype {
				t.Errorf("Classify(%q) = %s (conf=%.2f), want %s", tc.query, result.Label, result.Confidence, tc.atype)
			}
		})
	}
}

func TestClassifyReference(t *testing.T) {
	tests := []struct {
		query string
		atype AnswerType
	}{
		{"What references RepoMap?", Reference},
		{"Find references to KnowledgeStore", Reference},
		{"Show references to ParseLevel", Reference},
		{"What tests cover SymbolRetriever?", Reference},
		{"Which tests cover KnowledgeStore?", Reference},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query, "chat")
			if result.Type != tc.atype {
				t.Errorf("Classify(%q) = %s (conf=%.2f), want %s", tc.query, result.Label, result.Confidence, tc.atype)
			}
		})
	}
}

func TestClassifyOverview(t *testing.T) {
	tests := []struct {
		query string
		atype AnswerType
	}{
		{"What does this repo do?", Overview},
		{"Describe this project", Overview},
		{"Overview of the codebase", Overview},
		{"Tell me about this project", Overview},
		{"Give me an overview", Overview},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query, "chat")
			if result.Type != tc.atype {
				t.Errorf("Classify(%q) = %s (conf=%.2f), want %s", tc.query, result.Label, result.Confidence, tc.atype)
			}
		})
	}
}

func TestClassifyImprovement(t *testing.T) {
	tests := []struct {
		query string
		atype AnswerType
	}{
		{"Suggest improvements for this repo", Improvement},
		{"How can this be improved?", Improvement},
		{"What should I improve?", Improvement},
		{"Areas for improvement", Improvement},
		{"Potential improvements", Improvement},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query, "chat")
			if result.Type != tc.atype {
				t.Errorf("Classify(%q) = %s (conf=%.2f), want %s", tc.query, result.Label, result.Confidence, tc.atype)
			}
		})
	}
}

func TestClassifyRepositoryAnalysis(t *testing.T) {
	tests := []struct {
		query string
		atype AnswerType
	}{
		{"Analyze repository", RepositoryAnalysis},
		{"Analyze this codebase", RepositoryAnalysis},
		{"Repository analysis", RepositoryAnalysis},
		{"Codebase analysis", RepositoryAnalysis},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query, "chat")
			if result.Type != tc.atype {
				t.Errorf("Classify(%q) = %s (conf=%.2f), want %s", tc.query, result.Label, result.Confidence, tc.atype)
			}
		})
	}
}

func TestClassifyArchitectureReview(t *testing.T) {
	tests := []struct {
		query string
		atype AnswerType
	}{
		{"Review architecture", ArchitectureReview},
		{"Architecture review", ArchitectureReview},
		{"System architecture analysis", ArchitectureReview},
		{"Architectural analysis", ArchitectureReview},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query, "chat")
			if result.Type != tc.atype {
				t.Errorf("Classify(%q) = %s (conf=%.2f), want %s", tc.query, result.Label, result.Confidence, tc.atype)
			}
		})
	}
}

func TestClassifyFeatureSuggestion(t *testing.T) {
	tests := []struct {
		query string
		atype AnswerType
	}{
		{"Suggest features for this project", FeatureSuggestion},
		{"What features could I add?", FeatureSuggestion},
		{"Feature ideas", FeatureSuggestion},
		{"What should I add next?", FeatureSuggestion},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query, "chat")
			if result.Type != tc.atype {
				t.Errorf("Classify(%q) = %s (conf=%.2f), want %s", tc.query, result.Label, result.Confidence, tc.atype)
			}
		})
	}
}

func TestClassifyPlan(t *testing.T) {
	tests := []struct {
		query string
		atype AnswerType
	}{
		{"Plan to implement caching", Plan},
		{"Steps to add authentication", Plan},
		{"How would you implement tests?", Plan},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			result := Classify(tc.query, "chat")
			if result.Type != tc.atype {
				t.Errorf("Classify(%q) = %s (conf=%.2f), want %s", tc.query, result.Label, result.Confidence, tc.atype)
			}
		})
	}
}

func TestClassifyModePlan(t *testing.T) {
	result := Classify("anything here", "plan")
	if result.Type != Plan {
		t.Errorf("expected Plan for mode=plan, got %s", result.Label)
	}
}

func TestClassifyModeAgent(t *testing.T) {
	result := Classify("anything here", "agent")
	if result.Type != Agent {
		t.Errorf("expected Agent for mode=agent, got %s", result.Label)
	}
}

func TestMaxTokens(t *testing.T) {
	tests := []struct {
		atype      AnswerType
		wantTokens int
	}{
		{YesNo, 10},
		{Location, 25},
		{Caller, 50},
		{Reference, 50},
		{Overview, 150},
		{Improvement, 200},
		{FeatureSuggestion, 200},
		{ArchitectureReview, 250},
		{RepositoryAnalysis, 300},
		{Explanation, 400},
		{Plan, 500},
		{Agent, 0},
	}
	for _, tc := range tests {
		t.Run(tc.atype.String(), func(t *testing.T) {
			got := tc.atype.MaxTokens()
			if got != tc.wantTokens {
				t.Errorf("%s.MaxTokens() = %d, want %d", tc.atype.String(), got, tc.wantTokens)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		atype AnswerType
		want  string
	}{
		{YesNo, "yes_no"},
		{Location, "location"},
		{Reference, "reference"},
		{Caller, "caller"},
		{Overview, "overview"},
		{Explanation, "explanation"},
		{Plan, "plan"},
		{Agent, "agent"},
		{Improvement, "improvement"},
		{RepositoryAnalysis, "repository_analysis"},
		{ArchitectureReview, "architecture_review"},
		{FeatureSuggestion, "feature_suggestion"},
		{AnswerType(99), "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := tc.atype.String()
			if got != tc.want {
				t.Errorf("(%d).String() = %q, want %q", tc.atype, got, tc.want)
			}
		})
	}
}

func TestClassifyDefaultFallback(t *testing.T) {
	result := Classify("i want to know more about the system", "chat")
	if result.Type != Explanation {
		t.Errorf("expected Explanation fallback, got %s (conf=%.2f, reason=%s)", result.Label, result.Confidence, result.Reason)
	}
	if result.Confidence < 0.4 || result.Confidence > 0.6 {
		t.Errorf("expected confidence ~0.5, got %.2f", result.Confidence)
	}
}

func TestClassifyQuestionFallback(t *testing.T) {
	result := Classify("What is the meaning of life?", "chat")
	if result.Type != Explanation {
		t.Errorf("expected Explanation for question, got %s", result.Label)
	}
}

func TestClassifyWhatDoesThisRepoDo(t *testing.T) {
	result := Classify("What does this repo do?", "chat")
	if result.Type != Overview {
		t.Errorf("expected Overview, got %s (conf=%.2f)", result.Label, result.Confidence)
	}
}

func TestClassifySummaryIn3Bullets(t *testing.T) {
	result := Classify("Summarize this repository in 3 bullets", "chat")
	if result.Type != Overview {
		t.Errorf("expected Overview, got %s (conf=%.2f)", result.Label, result.Confidence)
	}
}

func TestClassifyHowCanThisBeImproved(t *testing.T) {
	result := Classify("How can this repo be improved?", "chat")
	if result.Type != Improvement {
		t.Errorf("expected Improvement, got %s (conf=%.2f)", result.Label, result.Confidence)
	}
}

func TestClassifyDeepAnalyzeRepo(t *testing.T) {
	result := Classify("Deep analyze repository", "chat")
	if result.Type != RepositoryAnalysis {
		t.Errorf("expected RepositoryAnalysis, got %s (conf=%.2f)", result.Label, result.Confidence)
	}
}

func TestClassifyDoesRepoContainAuth(t *testing.T) {
	result := Classify("Does this repo contain authentication?", "chat")
	if result.Type != YesNo {
		t.Errorf("expected YesNo, got %s (conf=%.2f)", result.Label, result.Confidence)
	}
}

func TestClassifyDoesLoginExist(t *testing.T) {
	result := Classify("Does login exist?", "chat")
	if result.Type != YesNo {
		t.Errorf("expected YesNo, got %s (conf=%.2f)", result.Label, result.Confidence)
	}
}

func TestClassifyHighestImpactImprovement(t *testing.T) {
	result := Classify("Give me only the highest impact improvement", "chat")
	if result.Type != Improvement {
		t.Errorf("expected Improvement, got %s (conf=%.2f)", result.Label, result.Confidence)
	}
}
