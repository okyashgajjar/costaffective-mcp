package architecture

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	rePyDocstring = regexp.MustCompile(`"""([^"]*)"""`)
	rePySingleDQ  = regexp.MustCompile(`'''([^']*)'''`)
	reGoPackage   = regexp.MustCompile(`(?m)^//\s*Package\s+\S+\s+(.*)$`)
	reJSDocStart  = regexp.MustCompile(`(?ms)/\*\*([^*]+)\*/`)
)

var topicMap = map[string][]string{
	"exception":  {"error handling", "failures", "retry logic", "API failures"},
	"exceptions": {"error handling", "failures", "retry logic", "API failures"},
	"error":      {"error handling", "failures", "exceptions"},
	"err":        {"error handling", "failures"},
	"retry":      {"retry logic", "resilience", "backoff"},
	"cache":      {"caching", "performance", "memoization"},
	"cached":     {"caching", "performance"},
	"caching":    {"caching", "performance"},
	"prompt":     {"prompts", "instructions", "system messages"},
	"prompts":    {"prompts", "instructions"},
	"chat":       {"chat", "conversation", "messages"},
	"chunk":      {"batching", "message grouping", "chunks"},
	"chunks":     {"batching", "message grouping", "chunks"},
	"repo":       {"repository", "git", "source control"},
	"repository": {"repository", "git", "source control"},
	"repomap":    {"repository mapping", "code indexing", "graph"},
	"map":        {"mapping", "graph", "indexing"},
	"model":      {"LLM", "machine learning", "AI model"},
	"models":     {"LLM", "machine learning", "AI model"},
	"llm":        {"LLM", "language model", "AI"},
	"litellm":    {"LLM", "language model", "API client"},
	"send":       {"messaging", "communication", "API call"},
	"sendchat":   {"messaging", "communication", "API call"},
	"io":         {"input output", "terminal", "user interface"},
	"input":      {"input", "user input", "keyboard"},
	"output":     {"output", "display", "rendering"},
	"edit":       {"editing", "code modification", "diffs"},
	"editing":    {"editing", "code modification", "diffs"},
	"diff":       {"diffs", "comparison", "patches"},
	"diffs":      {"diffs", "comparison", "patches"},
	"format":     {"formatting", "code style", "output format"},
	"lint":       {"linting", "code quality", "validation"},
	"linter":     {"linting", "code quality", "validation"},
	"history":    {"history", "conversation log", "persistence"},
	"voice":      {"voice input", "audio", "speech recognition"},
	"analytics":  {"analytics", "telemetry", "usage tracking"},
	"config":     {"configuration", "settings", "options"},
	"args":       {"configuration", "CLI arguments", "settings"},
	"main":       {"entry point", "startup", "CLI"},
	"startup":    {"entry point", "initialization", "startup"},
	"init":       {"initialization", "constructor", "setup"},
	"test":       {"testing", "unit tests", "test framework"},
	"benchmark":  {"benchmarking", "performance measurement"},
	"token":      {"tokens", "tokenization", "context length"},
	"tokens":     {"tokens", "tokenization", "context length"},
	"context":    {"context", "context window", "prompt size"},
	"coder":      {"AI coding agent", "code generation", "autonomous agent"},
	"coders":     {"AI coding agents", "code generation", "autonomous agents"},
	"watch":      {"file watching", "monitoring", "events"},
	"watcher":    {"file watching", "monitoring", "events"},
	"copy":       {"clipboard", "paste", "clipboard watcher"},
	"paste":      {"clipboard", "paste", "clipboard watcher"},
	"version":    {"version check", "updates", "compatibility"},
	"report":     {"reporting", "exception reporting", "telemetry"},
	"dump":       {"serialization", "data export", "debugging"},
	"api":        {"API", "HTTP", "REST"},
	"client":     {"client", "consumer", "service consumer"},
	"settings":   {"settings", "configuration", "options"},
	"editor":     {"editor", "text editing", "user input"},
	"hash":       {"hashing", "checksums", "fingerprints"},
	"analytics_event": {"analytics", "telemetry", "event tracking"},
	"ask":        {"interactive", "questions", "conversational"},
	"architect":  {"architecture mode", "planning", "design"},
}

var topicToTopics = map[string][]string{
	"error handling":   {"failures", "exceptions", "resilience"},
	"failures":         {"error handling", "reliability"},
	"exceptions":       {"error handling", "failures"},
	"retry logic":      {"resilience", "error recovery"},
	"caching":          {"performance", "memoization"},
	"prompts":          {"LLM", "instructions", "system messages"},
	"chat":             {"conversation", "messages"},
	"repository":       {"git", "source control", "codebase"},
	"LLM":              {"language model", "AI", "OpenAI", "Anthropic"},
	"messaging":        {"API call", "communication"},
	"input":            {"user input", "interactive"},
	"editing":          {"code modification", "diffs"},
	"diffs":            {"comparison", "patches"},
	"formatting":       {"code style", "presentation"},
	"linting":          {"code quality", "validation"},
	"history":          {"conversation log", "persistence"},
	"voice input":      {"audio", "speech recognition"},
	"analytics":        {"telemetry", "usage tracking"},
	"help system":      {"documentation", "user guidance"},
	"configuration":    {"settings", "options"},
	"CLI":              {"command line", "terminal", "shell"},
	"tokens":           {"tokenization", "context length"},
	"context":          {"context window", "prompt size"},
	"clipboard":        {"copy paste", "clipboard watcher"},
	"version check":    {"updates", "compatibility"},
	"exception reporting": {"telemetry", "error tracking"},
	"serialization":    {"data export", "debugging"},
	"API":              {"HTTP", "REST", "client-server"},
	"editor":           {"text editing", "user input"},
	"hashing":          {"checksums", "fingerprints"},
	"telemetry":        {"analytics", "usage tracking"},
	"startup":          {"initialization", "entry point"},
	"initialization":   {"setup", "constructor"},
	"code generation":  {"AI coding", "code completion"},
	"interactive":      {"user interaction", "questions"},
	"architecture mode": {"planning", "design", "high-level design"},
	"performance":      {"optimization", "efficiency"},
}

func ExtractTopics(classes, functions, imports []string, filePath string) []string {
	seen := make(map[string]bool)
	var topics []string

	addTopic := func(t string) {
		lower := strings.ToLower(t)
		if !seen[lower] {
			seen[lower] = true
			topics = append(topics, lower)
		}
	}

	words := combineKeywords(classes, functions, imports)
	words = append(words, pathKeywords(filePath)...)

	for _, word := range words {
		wordLower := strings.ToLower(word)
		if mapped, ok := topicMap[wordLower]; ok {
			for _, t := range mapped {
				addTopic(t)
			}
		}
		if strings.Contains(wordLower, "cache") {
			addTopic("caching")
		}
		if strings.Contains(wordLower, "error") || strings.Contains(wordLower, "exception") {
			addTopic("error handling")
		}
		if strings.Contains(wordLower, "test") {
			addTopic("testing")
		}
		if strings.Contains(wordLower, "repo") {
			addTopic("repository")
			addTopic("git")
			addTopic("source control")
		}
		if strings.Contains(wordLower, "map") {
			addTopic("mapping")
			addTopic("graph")
			addTopic("indexing")
		}
		if strings.Contains(wordLower, "send") {
			addTopic("messaging")
			addTopic("communication")
			addTopic("API call")
		}
	}

	for _, t := range topics {
		if expansions, ok := topicToTopics[t]; ok {
			for _, e := range expansions {
				addTopic(e)
			}
		}
	}

	return topics
}

func combineKeywords(classes, functions, imports []string) []string {
	var words []string
	words = append(words, classes...)
	words = append(words, functions...)
	words = append(words, imports...)
	return words
}

func pathKeywords(relPath string) []string {
	parts := strings.Split(relPath, "/")
	var words []string
	for _, p := range parts {
		if p == "" || p == "." {
			continue
		}
		base := strings.TrimSuffix(p, filepath.Ext(p))
		if base != "" {
			words = append(words, base)
		}
		for _, seg := range splitCamel(base) {
			words = append(words, strings.ToLower(seg))
		}
	}
	return words
}

func splitCamel(s string) []string {
	var parts []string
	var cur []rune
	for _, r := range s {
		if unicode.IsUpper(r) && len(cur) > 0 {
			parts = append(parts, string(cur))
			cur = nil
		}
		cur = append(cur, r)
	}
	if len(cur) > 0 {
		parts = append(parts, string(cur))
	}
	return parts
}

func ExtractDescription(content, filePath, language string) string {
	switch language {
	case "python":
		if m := rePyDocstring.FindStringSubmatch(content); len(m) > 1 {
			return cleanDocstring(m[1])
		}
		if m := rePySingleDQ.FindStringSubmatch(content); len(m) > 1 {
			return cleanDocstring(m[1])
		}
		return extractLeadingComments(content, "#")
	case "go":
		if m := reGoPackage.FindStringSubmatch(content); len(m) > 2 {
			return strings.TrimSpace(m[2])
		}
		return extractLeadingComments(content, "//")
	case "javascript", "typescript":
		if m := reJSDocStart.FindStringSubmatch(content); len(m) > 1 {
			return cleanDocstring(m[1])
		}
		return extractLeadingComments(content, "//")
	}
	return ""
}

func cleanDocstring(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}

func extractLeadingComments(content, prefix string) string {
	lines := strings.Split(content, "\n")
	var comments []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			text := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			comments = append(comments, text)
		} else if len(comments) > 0 {
			break
		}
	}
	return strings.Join(comments, " ")
}

func ExtractImports(content, language string) []string {
	switch language {
	case "python":
		return extractPyImports(content)
	case "go":
		return extractGoImports(content)
	case "javascript", "typescript":
		return extractJSImports(content)
	}
	return nil
}

var rePyImport1 = regexp.MustCompile(`(?m)^import\s+([\w.]+)`)
var rePyImport2 = regexp.MustCompile(`(?m)^from\s+([\w.]+)\s+import`)

func extractPyImports(content string) []string {
	var imports []string
	seen := make(map[string]bool)
	for _, m := range rePyImport1.FindAllStringSubmatch(content, -1) {
		mod := m[1]
		mod = strings.Split(mod, ".")[0]
		if !seen[mod] {
			seen[mod] = true
			imports = append(imports, mod)
		}
	}
	for _, m := range rePyImport2.FindAllStringSubmatch(content, -1) {
		mod := m[1]
		mod = strings.Split(mod, ".")[0]
		if !seen[mod] {
			seen[mod] = true
			imports = append(imports, mod)
		}
	}
	return imports
}

var reGoImport = regexp.MustCompile(`(?m)^\s*"([^"]+)"`)

func extractGoImports(content string) []string {
	lines := strings.Split(content, "\n")
	inImport := false
	var imports []string
	seen := make(map[string]bool)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import") {
			inImport = true
			if strings.Contains(trimmed, "(") {
				continue
			}
			if m := reGoImport.FindStringSubmatch(line); len(m) > 1 {
				if !seen[m[1]] {
					seen[m[1]] = true
					imports = append(imports, m[1])
				}
			}
			continue
		}
		if inImport {
			if strings.HasPrefix(trimmed, ")") {
				inImport = false
				continue
			}
			if m := reGoImport.FindStringSubmatch(line); len(m) > 1 {
				pkg := m[1]
				pkg = strings.Split(pkg, "/")[len(strings.Split(pkg, "/"))-1]
				if !seen[pkg] {
					seen[pkg] = true
					imports = append(imports, pkg)
				}
			}
		}
	}
	return imports
}

var reJSImport = regexp.MustCompile(`(?m)(?:import\s+.*?from\s+|require\(\s*)["']([^"']+)["']`)

func extractJSImports(content string) []string {
	var imports []string
	seen := make(map[string]bool)
	for _, m := range reJSImport.FindAllStringSubmatch(content, -1) {
		mod := m[1]
		if strings.HasPrefix(mod, ".") {
			mod = strings.TrimPrefix(mod, "./")
			mod = strings.Split(mod, "/")[0]
		} else {
			mod = strings.Split(mod, "/")[0]
			if strings.HasPrefix(mod, "@") {
				parts := strings.SplitN(strings.TrimPrefix(mod, "@"), "/", 2)
				if len(parts) > 0 {
					mod = "@" + parts[0]
				}
			}
		}
		if !seen[mod] {
			seen[mod] = true
			imports = append(imports, mod)
		}
	}
	return imports
}
