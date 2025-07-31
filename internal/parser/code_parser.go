package parser

import (
	"context"
	"log"

	"github.com/google/uuid"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/python"
)

type CodeParser struct {
	language   *sitter.Language
	nodeTypes  []string
	langString Language
}

func NewCodeParser(language Language) *CodeParser {
	switch language {
	case LanguagePython:
		return &CodeParser{
			language:   python.GetLanguage(),
			nodeTypes:  []string{"class_definition", "function_definition"},
			langString: language,
		}
	case LanguageGo:
		return &CodeParser{
			language:   golang.GetLanguage(),
			nodeTypes:  []string{"type_declaration", "function_declaration", "method_declaration"},
			langString: language,
		}
	}
	return nil
}

func (p *CodeParser) isTargetType(nodeType string) bool {
	for _, t := range p.nodeTypes {
		if t == nodeType {
			return true
		}
	}
	return false
}

func (p *CodeParser) ParseFile(ctx context.Context, content []byte, filename string) []*Snippet {
	parser := sitter.NewParser()
	parser.SetLanguage(p.language)

	tree, err := parser.ParseCtx(ctx, nil, content)
	if err != nil {
		log.Printf("Error parsing file %s: %v", filename, err)
		return nil
	}

	cursor := sitter.NewTreeCursor(tree.RootNode())
	snippets := []*Snippet{
		{
			ID:       uuid.New().String(),
			Content:  cursor.CurrentNode().Content(content),
			Filename: filename,
			Language: string(p.langString),
		},
	}

	if cursor.GoToFirstChild() {
		for {
			node := cursor.CurrentNode()
			if p.isTargetType(node.Type()) {
				snippets = append(snippets, &Snippet{
					ID:       uuid.New().String(),
					Content:  node.Content(content),
					Filename: filename,
					Language: string(p.langString),
				})
			}
			if !cursor.GoToNextSibling() {
				break
			}
		}
	}

	return snippets
}

type ProjectParser struct {
	parsers map[string]*CodeParser
}
