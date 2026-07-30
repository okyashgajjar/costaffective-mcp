package lsif

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type RangeData struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string
	Range RangeData
}

type Index struct {
	Ranges      map[int]Location
	Definitions map[int][]int
	References  map[int][]int
	NextMap     map[int]int // RangeID -> ResultSetID
}

// Element represents a generic LSIF vertex or edge
type Element struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`

	// Vertex fields
	URI   string    `json:"uri,omitempty"`
	Start *Position `json:"start,omitempty"`
	End   *Position `json:"end,omitempty"`

	// Edge fields
	OutV int   `json:"outV,omitempty"`
	InV  int   `json:"inV,omitempty"`
	InVs []int `json:"inVs,omitempty"`
}

func NewIndex() *Index {
	return &Index{
		Ranges:      make(map[int]Location),
		Definitions: make(map[int][]int),
		References:  make(map[int][]int),
		NextMap:     make(map[int]int),
	}
}

// Parse parses a standard .lsif JSON-lines dump and builds an in-memory graph.
func Parse(filepath string) (*Index, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	idx := NewIndex()

	documents := make(map[int]string) // DocumentID -> URI
	ranges := make(map[int]RangeData) // RangeID -> RangeData

	// First pass: collect vertices
	scanner := bufio.NewScanner(file)
	// Some LSIF files have huge lines if they contain base64 content
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var edges []Element

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var el Element
		if err := json.Unmarshal(line, &el); err != nil {
			continue // Skip malformed lines
		}

		if el.Type == "vertex" {
			switch el.Label {
			case "document":
				uri := strings.TrimPrefix(el.URI, "file://")
				documents[el.ID] = uri
			case "range":
				if el.Start != nil && el.End != nil {
					ranges[el.ID] = RangeData{Start: *el.Start, End: *el.End}
				}
			}
		} else if el.Type == "edge" {
			// We delay processing edges until all vertices are loaded
			// To save memory, we only keep relevant edges
			if el.Label == "contains" || el.Label == "next" || el.Label == "item" || el.Label == "textDocument/definition" || el.Label == "textDocument/references" {
				edges = append(edges, el)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	// DefinitionResultID -> ResultSetID
	defResults := make(map[int]int)
	// ReferenceResultID -> ResultSetID
	refResults := make(map[int]int)

	// Second pass: process edges
	for _, el := range edges {
		switch el.Label {
		case "contains":
			if uri, ok := documents[el.OutV]; ok {
				for _, inV := range el.InVs {
					if rData, ok := ranges[inV]; ok {
						idx.Ranges[inV] = Location{URI: uri, Range: rData}
					}
				}
			}
		case "next":
			// range -> resultSet
			idx.NextMap[el.OutV] = el.InV
		case "textDocument/definition":
			// resultSet -> definitionResult
			defResults[el.InV] = el.OutV
		case "textDocument/references":
			// resultSet -> referenceResult
			refResults[el.InV] = el.OutV
		}
	}

	// Third pass: process items (they rely on definitionResult / referenceResult)
	for _, el := range edges {
		if el.Label == "item" {
			// definitionResult -> range
			if resultSetID, ok := defResults[el.OutV]; ok {
				idx.Definitions[resultSetID] = append(idx.Definitions[resultSetID], el.InVs...)
			}
			// referenceResult -> range
			if resultSetID, ok := refResults[el.OutV]; ok {
				idx.References[resultSetID] = append(idx.References[resultSetID], el.InVs...)
			}
		}
	}

	return idx, nil
}
