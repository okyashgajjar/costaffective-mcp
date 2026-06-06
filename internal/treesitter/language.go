package treesitter

import (
	"path/filepath"
	"strings"
)

type Language string

const (
	LangGo         Language = "go"
	LangPython     Language = "python"
	LangJavaScript Language = "javascript"
	LangTypeScript Language = "typescript"
)

func DetectLanguage(filePath string) Language {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return LangGo
	case ".py":
		return LangPython
	case ".js", ".jsx", ".mjs":
		return LangJavaScript
	case ".ts", ".tsx":
		return LangTypeScript
	default:
		return ""
	}
}

func IsSupported(path string) bool {
	return DetectLanguage(path) != ""
}

var SupportedLanguages = []Language{LangGo, LangPython, LangJavaScript, LangTypeScript}
