package treesitter

type RefType int

const (
	RefDefinition RefType = iota
	RefReference
	RefImport
	RefExport
)

func (r RefType) String() string {
	switch r {
	case RefDefinition:
		return "definition"
	case RefReference:
		return "reference"
	case RefImport:
		return "import"
	case RefExport:
		return "export"
	default:
		return "unknown"
	}
}

type Reference struct {
	SymbolID   string  `json:"symbol_id"`
	SymbolName string  `json:"symbol_name"`
	File       string  `json:"file"`
	Line       int     `json:"line"`
	Column     int     `json:"column"`
	RefType    RefType `json:"ref_type"`
	Context    string  `json:"context"`
}

type RefQueryResult struct {
	Symbol     Symbol      `json:"symbol"`
	Definition *Reference  `json:"definition,omitempty"`
	References []Reference `json:"references"`
	Imports    []Reference `json:"imports"`
	Exports    []Reference `json:"exports"`
	Score      float64     `json:"score"`
}

type defNodeType struct {
	parentType string
	childType  string
}

var definitionNodeTypes = map[string][]defNodeType{
	"python": {
		{"function_definition", "identifier"},
		{"class_definition", "identifier"},
	},
	"go": {
		{"function_declaration", "identifier"},
		{"method_declaration", "identifier"},
		{"type_spec", "identifier"},
		{"field_declaration", "identifier"},
	},
	"javascript": {
		{"function_declaration", "identifier"},
		{"method_definition", "identifier"},
		{"class_declaration", "identifier"},
		{"variable_declarator", "identifier"},
	},
	"typescript": {
		{"function_declaration", "identifier"},
		{"method_definition", "identifier"},
		{"class_declaration", "identifier"},
		{"interface_declaration", "identifier"},
		{"variable_declarator", "identifier"},
	},
}

var importNodeTypes = map[string][]string{
	"python":     {"import_statement", "import_from_statement"},
	"go":         {"import_declaration", "import_spec"},
	"javascript": {"import_statement", "import_spec"},
	"typescript": {"import_statement", "import_spec"},
}

var exportNodeTypes = map[string][]string{
	"javascript": {"export_statement"},
	"typescript": {"export_statement"},
}

var callNodeTypes = map[string][]string{
	"python":     {"call"},
	"go":         {"call_expression"},
	"javascript": {"call_expression"},
	"typescript": {"call_expression"},
}
