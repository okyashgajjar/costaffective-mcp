package policy

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type Violation struct {
	Rule    *ValidationRule
	Message string
	Line    int // 0 if file-level
}

type EvaluationResult struct {
	Score      int
	Category   string
	Violations []Violation
}

type Engine struct {
	policy     *Policy
	regexCache map[string]*regexp.Regexp
	globCache  map[string]*regexp.Regexp
	cacheMu    sync.RWMutex
}

func NewEngine(p *Policy) *Engine {
	return &Engine{
		policy:     p,
		regexCache: make(map[string]*regexp.Regexp),
		globCache:  make(map[string]*regexp.Regexp),
	}
}

// compileRegex caches and returns a compiled regex.
func (e *Engine) compileRegex(pattern string) (*regexp.Regexp, error) {
	e.cacheMu.RLock()
	re, ok := e.regexCache[pattern]
	e.cacheMu.RUnlock()
	if ok {
		return re, nil
	}

	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	// Double-check
	if re, ok := e.regexCache[pattern]; ok {
		return re, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	e.regexCache[pattern] = re
	return re, nil
}

// globToRegex translates a glob pattern to a regex pattern.
func globToRegex(glob string) string {
	var sb strings.Builder
	sb.WriteString("^")

	// Fast translation: * -> .*, ? -> ., . -> \.
	for i := 0; i < len(glob); i++ {
		c := glob[i]
		switch c {
		case '*':
			// Handle ** for directory crossing
			if i+1 < len(glob) && glob[i+1] == '*' {
				sb.WriteString(".*")
				i++ // skip next *
				// if followed by /, make it optional
				if i+1 < len(glob) && glob[i+1] == '/' {
					sb.WriteString("/?")
					i++
				}
			} else {
				sb.WriteString("[^/]*")
			}
		case '?':
			sb.WriteString("[^/]")
		case '.', '+', '(', ')', '|', '^', '$', '[', ']', '{', '}':
			sb.WriteString(`\`)
			sb.WriteByte(c)
		default:
			sb.WriteByte(c)
		}
	}
	sb.WriteString("$")
	return sb.String()
}

// matchGlob checks if a file path matches a glob pattern using translated regex.
func (e *Engine) matchGlob(pattern, path string) bool {
	e.cacheMu.RLock()
	re, ok := e.globCache[pattern]
	e.cacheMu.RUnlock()

	if !ok {
		e.cacheMu.Lock()
		if re, ok = e.globCache[pattern]; !ok {
			r := globToRegex(pattern)
			compiled, err := regexp.Compile(r)
			if err == nil {
				re = compiled
				e.globCache[pattern] = compiled
			}
		}
		e.cacheMu.Unlock()
	}

	if re != nil {
		// Normalize path separators to /
		path = filepath.ToSlash(path)
		return re.MatchString(path)
	}

	// Fallback to standard filepath.Match if translation failed
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	return matched
}

// matchAnyGlob returns true if the path matches any of the given globs.
func (e *Engine) matchAnyGlob(globs []string, path string) bool {
	for _, g := range globs {
		if e.matchGlob(g, path) {
			return true
		}
	}
	return false
}

// EvaluateFile validates a file's content against the loaded policy.
func (e *Engine) EvaluateFile(relPath, content string) EvaluationResult {
	if e.policy == nil {
		return EvaluationResult{Score: 100, Category: "Unknown"}
	}

	// 1. Identify category
	categoryName := "misc" // default
	var matchedCat *Category

	for _, cat := range e.policy.Categories {
		if e.matchAnyGlob(cat.Patterns, relPath) {
			categoryName = cat.Name
			matchedCat = &cat
			break
		}
	}

	res := EvaluationResult{
		Score:      100,
		Category:   categoryName,
		Violations: []Violation{},
	}

	if matchedCat == nil {
		return res // No rules to enforce
	}

	// 2. Evaluate rules
	for i := range matchedCat.Rules {
		rule := &matchedCat.Rules[i]

		// Does this rule apply to this file?
		if len(rule.Match) > 0 && !e.matchAnyGlob(rule.Match, relPath) {
			continue
		}

		// Check Require patterns
		for _, req := range rule.Require {
			re, err := e.compileRegex(req)
			if err != nil {
				continue
			}
			if !re.MatchString(content) {
				res.Violations = append(res.Violations, Violation{
					Rule:    rule,
					Message: fmt.Sprintf("Missing required pattern: %s. %s", req, rule.Message),
				})
				res.Score--
			}
		}

		// Check Deny patterns
		for _, deny := range rule.Deny {
			re, err := e.compileRegex(deny)
			if err != nil {
				continue
			}
			if loc := re.FindStringIndex(content); loc != nil {
				// Naive line number calculation for the report
				line := strings.Count(content[:loc[0]], "\n") + 1
				res.Violations = append(res.Violations, Violation{
					Rule:    rule,
					Message: fmt.Sprintf("Found denied pattern: %s. %s", deny, rule.Message),
					Line:    line,
				})
				res.Score--
			}
		}
	}

	return res
}
