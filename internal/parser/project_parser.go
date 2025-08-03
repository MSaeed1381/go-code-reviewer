package parser

import (
	"context"
	"go_code_reviewer/internal/models"
	"go_code_reviewer/pkg/log"
	"os"
	"path/filepath"
	"strings"
)

type ProjectParser struct {
	parsers map[string]*CodeParser
}

func NewProjectParser(parsers map[string]*CodeParser) *ProjectParser {
	return &ProjectParser{
		parsers: parsers,
	}
}

func (pp *ProjectParser) ParseProject(ctx context.Context, rootPath string) ([]*models.Snippet, error) {
	logger := log.GetLogger()
	var allSnippets []*models.Snippet
	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if parser, supported := pp.parsers[ext]; supported {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			fileSnippets := parser.ParseFile(ctx, content, path)
			allSnippets = append(allSnippets, fileSnippets...)
		}
		return nil
	})
	if err != nil {
		logger.WithError(err).Error("failed to walk project")
		return nil, err
	}

	return allSnippets, nil
}
