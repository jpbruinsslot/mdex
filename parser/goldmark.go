package parser

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type GoldmarkParser struct {
	mdParser goldmark.Markdown
}

func NewGoldmarkParser() *GoldmarkParser {
	mdParser := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			extension.Linkify,
			extension.Footnote,
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	return &GoldmarkParser{
		mdParser: mdParser,
	}
}

func (p *GoldmarkParser) Convert(markdown []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := p.mdParser.Convert(markdown, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *GoldmarkParser) ExtractTOC(markdown []byte) ([]TOCEntry, error) {
	// Parse the markdown with the same parser configuration
	reader := text.NewReader(markdown)
	doc := p.mdParser.Parser().Parse(reader)

	var toc []TOCEntry

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			var title bytes.Buffer

			// Extract text content from all child nodes
			for c := heading.FirstChild(); c != nil; c = c.NextSibling() {
				if textNode, ok := c.(*ast.Text); ok {
					title.Write(textNode.Segment.Value(markdown))
				}
			}

			// Get the heading ID - this will be auto-generated if WithAutoHeadingID() is used
			var headingID string
			if id, exists := heading.AttributeString("id"); exists {
				if idStr, ok := id.(string); ok {
					headingID = idStr
				} else if idBytes, ok := id.([]byte); ok {
					headingID = string(idBytes)
				} else {
					headingID = generateID(title.String())
				}
			} else {
				// Fallback: generate ID from title if no auto-generated ID exists
				headingID = generateID(title.String())
			}

			toc = append(toc, TOCEntry{
				Title: strings.TrimSpace(title.String()),
				Level: heading.Level,
				ID:    headingID,
			})
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, err
	}

	return toc, nil
}

func generateID(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "-")
}
