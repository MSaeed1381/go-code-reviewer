package parser

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func NewProjectParser() *ProjectParser {
	return &ProjectParser{
		parsers: map[string]*CodeParser{
			".py": NewCodeParser(LanguagePython),
			".go": NewCodeParser(LanguageGo),
		},
	}
}

func (pp *ProjectParser) ParseProject(ctx context.Context, rootPath string) ([]*Snippet, error) {
	var allSnippets []*Snippet

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if parser, supported := pp.parsers[ext]; supported {
			log.Printf("Processing file: %s", path)
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				log.Printf("Failed to read file %s: %v", path, readErr)
				return nil
			}
			fileSnippets := parser.ParseFile(ctx, content, path)
			allSnippets = append(allSnippets, fileSnippets...)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Printf("Total snippets parsed: %d", len(allSnippets))
	return allSnippets, nil
}
