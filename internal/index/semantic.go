package index

import (
	"context"
	"path/filepath"

	"github.com/blugelabs/bluge"
)

type SemanticResult struct {
	Path  string
	Score float64
}

type SemanticIndexer struct {
	writer *bluge.Writer
}

func NewSemanticIndexer(dbPath string) (*SemanticIndexer, error) {
	config := bluge.DefaultConfig(filepath.Join(dbPath, "semantic.bluge"))
	writer, err := bluge.OpenWriter(config)
	if err != nil {
		return nil, err
	}
	return &SemanticIndexer{writer: writer}, nil
}

func (s *SemanticIndexer) Close() error {
	if s.writer != nil {
		return s.writer.Close()
	}
	return nil
}

func (s *SemanticIndexer) IndexFile(path string, content string, docstrings string) error {
	doc := bluge.NewDocument(path)
	doc.AddField(bluge.NewKeywordField("path", path).StoreValue())
	doc.AddField(bluge.NewTextField("content", content))
	doc.AddField(bluge.NewTextField("docstrings", docstrings))

	return s.writer.Update(doc.ID(), doc)
}

func (s *SemanticIndexer) RemoveFile(path string) error {
	return s.writer.Delete(bluge.Identifier(path))
}

func (s *SemanticIndexer) Search(query string, limit int) ([]SemanticResult, error) {
	reader, err := s.writer.Reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	q1 := bluge.NewMatchQuery(query).SetField("content")
	q2 := bluge.NewMatchQuery(query).SetField("docstrings")
	q3 := bluge.NewMatchQuery(query).SetField("path")
	q := bluge.NewBooleanQuery().AddShould(q1, q2, q3)
	
	req := bluge.NewTopNSearch(limit, q).WithStandardAggregations()
	dmi, err := reader.Search(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var results []SemanticResult
	next, err := dmi.Next()
	for err == nil && next != nil {
		var path string
		err = next.VisitStoredFields(func(field string, value []byte) bool {
			if field == "path" {
				path = string(value)
			}
			return true
		})
		if err != nil {
			return nil, err
		}
		if path != "" {
			results = append(results, SemanticResult{Path: path, Score: next.Score})
		}
		next, err = dmi.Next()
	}
	return results, err
}
