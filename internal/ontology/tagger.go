package ontology

import (
	"strings"

	"github.com/okyashgajjar/costwise-mcp/internal/treesitter"
)

type Tag struct {
	Concept string
	Domain  string
}

func AnalyzeSymbol(sym treesitter.Symbol) []Tag {
	var tags []Tag
	name := sym.Name
	path := sym.File

	if sym.Kind == "class" || sym.Kind == "struct" || sym.Kind == "type" {
		tags = append(tags, Tag{Concept: "TYPE_DECL"})
	} else if sym.Kind == "function" || sym.Kind == "method" {
		tags = append(tags, Tag{Concept: "METHOD"})
	}

	if strings.HasSuffix(name, "Handler") || strings.HasSuffix(name, "Controller") {
		tags = append(tags, Tag{Concept: "Handler"})
	} else if strings.HasSuffix(name, "Repository") || strings.HasSuffix(name, "Store") || strings.Contains(path, "/db/") {
		tags = append(tags, Tag{Concept: "Storage"})
	} else if strings.HasSuffix(name, "Service") || strings.HasSuffix(name, "Manager") {
		tags = append(tags, Tag{Concept: "Service"})
	} else if strings.HasSuffix(name, "Event") || strings.HasSuffix(name, "Message") {
		tags = append(tags, Tag{Concept: "Event"})
	}

	parts := strings.Split(path, "/")
	domain := ""
	if len(parts) > 2 && parts[0] == "internal" {
		domain = parts[1]
	} else if len(parts) > 1 && parts[0] != "cmd" && parts[0] != "pkg" && parts[0] != "internal" {
		domain = parts[0]
	}

	if domain != "" {
		for i := range tags {
			tags[i].Domain = domain
		}
	}

	return tags
}
