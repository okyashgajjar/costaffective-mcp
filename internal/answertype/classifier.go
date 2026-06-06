package answertype

import (
	"strings"
)

type AnswerType int

const (
	YesNo AnswerType = iota
	Location
	Reference
	Caller
	Overview
	Explanation
	Plan
	Agent
	Improvement
	RepositoryAnalysis
	ArchitectureReview
	FeatureSuggestion
)

func (at AnswerType) String() string {
	switch at {
	case YesNo:
		return "yes_no"
	case Location:
		return "location"
	case Reference:
		return "reference"
	case Caller:
		return "caller"
	case Overview:
		return "overview"
	case Explanation:
		return "explanation"
	case Plan:
		return "plan"
	case Agent:
		return "agent"
	case Improvement:
		return "improvement"
	case RepositoryAnalysis:
		return "repository_analysis"
	case ArchitectureReview:
		return "architecture_review"
	case FeatureSuggestion:
		return "feature_suggestion"
	default:
		return "unknown"
	}
}

func (at AnswerType) MaxTokens() int {
	switch at {
	case YesNo:
		return 10
	case Location:
		return 25
	case Caller:
		return 50
	case Reference:
		return 50
	case Overview:
		return 150
	case Explanation:
		return 400
	case Improvement:
		return 200
	case FeatureSuggestion:
		return 200
	case ArchitectureReview:
		return 250
	case RepositoryAnalysis:
		return 300
	case Plan:
		return 500
	case Agent:
		return 0
	default:
		return 250
	}
}

type Classification struct {
	Type       AnswerType
	Label      string
	Confidence float64
	Reason     string
}

var yesNoPatterns = []string{
	"does this repo have ", "does the codebase have ",
	"does this project have ", "is there a ", "is there an ",
	"are there ", "does it support ", "does it have ",
	"does this support ", "is this ", "is it ",
	"can this ", "can it ", "should this ",
	"has this ", "has it ",
	"do you have ", "do we have ",
}

var locationPatterns = []string{
	"where is ", "where are ", "where can i find ",
	"where does ", "find the location", "which file ",
	"what file ", "in which file ", "file path of ",
	"location of ", "path of ", "find file ",
}

var callerPatterns = []string{
	"who calls ", "find callers of ", "show callers of ",
	"show call sites of ", "callers of ", "called by ",
	"what calls ", "which functions call ",
	"who uses ", "who imports ",
}

var referencePatterns = []string{
	"what references ", "find references to ", "find usages of ",
	"find all references to ", "show references to ",
	"references to ", "usages of ", "who references ",
	"what tests cover ", "test coverage for ",
	"which tests ", "what files test ",
	"dependents of ", "dependants of ",
}

var overviewPatterns = []string{
	"what does this repo do", "what does this project do",
	"what is this repo", "what is this project",
	"describe this repo", "describe this project",
	"summarize this repo", "summarize this repository",
	"summarize this project",
	"overview of ", "summary of ", "about this ",
	"what is the purpose", "what problem does ",
	"how does this work",
	"how is this organized", "what are the main ",
	"tell me about this ", "explain this repo",
	"explain this project", "give me an overview",
}

var improvementPatterns = []string{
	"suggest improvements", "suggest improvement",
	"how can this be improved", "how can this be better",
	"how can this repo be improved", "how can this repo be better",
	"what should i improve", "what could be improved",
	"improve this repo", "improve this project",
	"improve the codebase", "improve the repository",
	"make this better", "potential improvements",
	"areas for improvement", "improvement suggestions",
	"highest impact improvement",
	"improvements for",
}

var repositoryAnalysisPatterns = []string{
	"analyze repository", "analyze codebase",
	"analyze this repo", "analyze this project",
	"analyze this codebase",
	"deep analyze",
	"repository analysis", "codebase analysis",
	"deep analysis", "code analysis",
}

var architectureReviewPatterns = []string{
	"architecture review", "architecture analysis",
	"review architecture", "architectural review",
	"architectural analysis", "system architecture",
}

var featureSuggestionPatterns = []string{
	"suggest features", "feature suggestion",
	"what features could", "new features for",
	"what should i add", "feature ideas",
	"potential features", "could add",
}

var explanationPatterns = []string{
	"explain how ", "how does ", "how do ",
	"how is ", "how are ", "how can ",
	"what is ", "what are ", "what does ",
	"why does ", "why is ", "why are ",
	"explain the ", "explain this ", "how it works",
}

var planPatterns = []string{
	"plan to ", "implementation plan",
	"how to implement ", "steps to ",
	"design a ", "create a plan",
	"what would it take to ",
	"how would you implement ",
}

var agentPatterns = []string{
	"implement ", "write code to ",
	"create a ", "add a ",
	"modify ", "refactor ",
	"fix ", "update ",
	"change ", "delete ",
}

func isYesNoQuestion(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	for _, p := range yesNoPatterns {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}

	firstWord := ""
	if fields := strings.Fields(lower); len(fields) > 0 {
		firstWord = fields[0]
	}
	switch firstWord {
	case "does", "is", "are", "can", "do", "has", "have", "should", "will", "did", "was", "were":
		auxPrefixes := []string{"does the", "what does", "how does"}
		for _, p := range auxPrefixes {
			if strings.HasPrefix(lower, p) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func containsAnyWord(lower string, words []string) bool {
	for _, w := range words {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

func matchAny(lower string, patterns []string) bool {
	for _, p := range patterns {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

func Classify(query string, mode string) Classification {
	lower := strings.ToLower(strings.TrimSpace(query))
	fields := strings.Fields(lower)

	mode = strings.ToLower(mode)
	if mode == "plan" {
		return Classification{
			Type:       Plan,
			Label:      "plan",
			Confidence: 1.0,
			Reason:     "mode:plan",
		}
	}
	if mode == "agent" {
		return Classification{
			Type:       Agent,
			Label:      "agent",
			Confidence: 1.0,
			Reason:     "mode:agent",
		}
	}

	numWords := len(fields)

	if numWords <= 6 && isYesNoQuestion(query) {
		return Classification{
			Type:       YesNo,
			Label:      "yes_no",
			Confidence: 0.85,
			Reason:     "yes_no_pattern",
		}
	}

	for _, p := range locationPatterns {
		if strings.HasPrefix(lower, p) {
			return Classification{
				Type:       Location,
				Label:      "location",
				Confidence: 0.9,
				Reason:     "location_pattern",
			}
		}
	}

	for _, p := range callerPatterns {
		if strings.HasPrefix(lower, p) {
			return Classification{
				Type:       Caller,
				Label:      "caller",
				Confidence: 0.9,
				Reason:     "caller_pattern",
			}
		}
	}

	for _, p := range referencePatterns {
		if strings.HasPrefix(lower, p) {
			return Classification{
				Type:       Reference,
				Label:      "reference",
				Confidence: 0.9,
				Reason:     "reference_pattern",
			}
		}
	}

	if matchAny(lower, improvementPatterns) {
		return Classification{
			Type:       Improvement,
			Label:      "improvement",
			Confidence: 0.85,
			Reason:     "improvement_pattern",
		}
	}

	if matchAny(lower, repositoryAnalysisPatterns) {
		return Classification{
			Type:       RepositoryAnalysis,
			Label:      "repository_analysis",
			Confidence: 0.85,
			Reason:     "repo_analysis_pattern",
		}
	}

	if matchAny(lower, architectureReviewPatterns) {
		return Classification{
			Type:       ArchitectureReview,
			Label:      "architecture_review",
			Confidence: 0.85,
			Reason:     "architecture_pattern",
		}
	}

	if matchAny(lower, featureSuggestionPatterns) {
		return Classification{
			Type:       FeatureSuggestion,
			Label:      "feature_suggestion",
			Confidence: 0.8,
			Reason:     "feature_suggestion_pattern",
		}
	}

	for _, p := range overviewPatterns {
		if strings.HasPrefix(lower, p) {
			return Classification{
				Type:       Overview,
				Label:      "overview",
				Confidence: 0.9,
				Reason:     "overview_pattern",
			}
		}
	}

	for _, p := range planPatterns {
		if strings.HasPrefix(lower, p) {
			return Classification{
				Type:       Plan,
				Label:      "plan",
				Confidence: 0.85,
				Reason:     "plan_pattern",
			}
		}
	}

	for _, p := range explanationPatterns {
		if strings.HasPrefix(lower, p) {
			return Classification{
				Type:       Explanation,
				Label:      "explanation",
				Confidence: 0.7,
				Reason:     "explanation_pattern",
			}
		}
	}

	for _, p := range agentPatterns {
		if strings.HasPrefix(lower, p) {
			return Classification{
				Type:       Plan,
				Label:      "plan",
				Confidence: 0.7,
				Reason:     "implementation_pattern",
			}
		}
	}

	if containsAnyWord(lower, []string{"improvement", "improvements", "improve"}) {
		return Classification{
			Type:       Improvement,
			Label:      "improvement",
			Confidence: 0.65,
			Reason:     "improvement_keyword",
		}
	}

	if strings.Contains(lower, "?") {
		return Classification{
			Type:       Explanation,
			Label:      "explanation",
			Confidence: 0.6,
			Reason:     "question_fallback",
		}
	}

	if numWords <= 4 {
		hasUpper := false
		for _, w := range fields {
			if len(w) >= 2 && w[0] >= 'A' && w[0] <= 'Z' {
				hasUpper = true
				break
			}
		}
		if hasUpper || numWords <= 2 {
			return Classification{
				Type:       Location,
				Label:      "location",
				Confidence: 0.65,
				Reason:     "short_symbol_query",
			}
		}
	}

	return Classification{
		Type:       Explanation,
		Label:      "explanation",
		Confidence: 0.5,
		Reason:     "default",
	}
}
