package architecture

type ModuleSummary struct {
	FilePath    string   `json:"file_path"`
	Language    string   `json:"language"`
	Classes     []string `json:"classes"`
	Functions   []string `json:"functions"`
	Imports     []string `json:"imports"`
	Exports     []string `json:"exports"`
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
}

type ArchitectureIndex struct {
	Modules map[string]*ModuleSummary
}

func NewArchitectureIndex() *ArchitectureIndex {
	return &ArchitectureIndex{
		Modules: make(map[string]*ModuleSummary),
	}
}

func (a *ArchitectureIndex) Add(summary *ModuleSummary) {
	if a.Modules == nil {
		a.Modules = make(map[string]*ModuleSummary)
	}
	a.Modules[summary.FilePath] = summary
}

func (a *ArchitectureIndex) Get(filePath string) (*ModuleSummary, bool) {
	s, ok := a.Modules[filePath]
	return s, ok
}

func (a *ArchitectureIndex) All() []*ModuleSummary {
	var out []*ModuleSummary
	for _, s := range a.Modules {
		out = append(out, s)
	}
	return out
}
