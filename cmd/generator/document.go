package generator

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type section struct {
	subSections []*section
	node        *ast.Heading
	name        string
}

func createSection(root *ast.Document, source []byte) *section {
	lastSection := []*section{{
		subSections: []*section{},
	}, nil, nil, nil, nil, nil, nil}
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		if n.Kind() == ast.KindHeading {
			cased := n.(*ast.Heading)
			level := cased.Level
			if lastSection[level-1] != nil {
				current := &section{
					subSections: []*section{},
					node:        cased,
					name:        string(cased.Text(source)),
				}
				lastSection[level-1].subSections = append(lastSection[level-1].subSections, current)
				lastSection[level] = current
				for i := level + 1; i <= 6; i++ {
					lastSection[i] = nil
				}
			}
		}
	}
	return lastSection[0]
}

func ParseDocument(raw []byte) ([]Endpoint, map[string]string) {
	docParser := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
	)
	ctx := parser.NewContext()
	docAST := docParser.Parser().Parse(text.NewReader(raw), parser.WithContext(ctx))
	sectionTree := createSection(docAST.(*ast.Document), raw)
	endpoints := []Endpoint{}
	rest := extractRestApi(sectionTree, raw)
	endpoints = append(endpoints, rest...)
	types := extractTypes(sectionTree, raw)
	return endpoints, types
}
