package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Kunde21/markdownfmt/v3/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"regexp"
	"strings"
)

// createMarkdownRender creates a new goldmark.Markdown renderer that converts a markdown AST back to markdown.
func createMarkdownRenderer(opts ...markdown.Option) goldmark.Markdown {
	mr := markdown.NewRenderer()
	mr.AddMarkdownOptions(opts...)
	extensions := []goldmark.Extender{
		extension.GFM,
	}
	parserOptions := []parser.Option{
		parser.WithAttribute(), // We need this to enable # headers {#custom-ids}.
	}

	gm := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(parserOptions...),
		goldmark.WithRenderer(mr),
	)

	return gm
}

// TransformMarkdownVariables is a walker function that replaces icon variables in markdown files with the corresponding HTML code.
func TransformMarkdownVariables(sourceContent []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if t, ok := n.(*ast.String); ok {
			for search, replace := range iconsMap {
				t.Value = bytes.ReplaceAll(t.Value, []byte(search), []byte(replace))
			}
		}
	}
	return ast.WalkContinue, nil
}

func replaceAdmonitionText(sourceContent []byte) ([]byte, error) {
	var builder strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(sourceContent))
	pattern := regexp.MustCompile(`^.*"([^"]*)".*$`)
	for scanner.Scan() {
		content := scanner.Text()
		if strings.Contains(content, "!!! ") {
			// extract the admonition title and use it as a heading
			content = pattern.ReplaceAllString(content, "### $1")
		}
		builder.WriteString(content + "\n")
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}

// FormatReleaseNotes rewrites the markdown source of a release note to a GitHub-flavoured markdown with the following changes:
// - relative links are converted to absolute links pointing to percona docs.
// - custom icon variables are changed to their SVG/HTML equivalent as described in:
// - Admonitions are transformed to headings (see: )
func FormatReleaseNotes(sourceContent []byte) ([]byte, error) {
	for search, replace := range iconsMap {
		sourceContent = bytes.ReplaceAll(sourceContent, []byte(search), []byte(replace))
	}

	sourceContent, err := replaceAdmonitionText(sourceContent)
	if err != nil {
		return nil, err
	}
	md := createMarkdownRenderer()
	reader := text.NewReader(sourceContent)
	doc := md.Parser().Parse(reader)
	baseMarkdownURL := "https://github.com/percona/pmm-doc/tree/main/docs/" // use GitHub since these are still raw markdown files.
	baseImageURL := "https://docs.percona.com/percona-monitoring-and-management/"

	var buffer bytes.Buffer
	if err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if link, ok := node.(*ast.Link); ok && entering {
			target := string(link.Destination)
			if isRelativeLink(target) && strings.HasPrefix(target, "../") {
				newDestination := baseMarkdownURL + strings.Replace(target, "../", "", 1)
				link.Destination = []byte(newDestination)
			}
		} else if image, ok := node.(*ast.Image); ok && entering {
			target := string(image.Destination)
			if isRelativeLink(target) && strings.HasPrefix(target, "../") {
				newDestination := baseImageURL + strings.Replace(target, "../", "", 1)
				image.Destination = []byte(newDestination)
			}
		}
		return ast.WalkContinue, nil
	}); err != nil {
		return nil, fmt.Errorf("failed to rewrite relative link: %v", err)
	}

	if err := md.Renderer().Render(&buffer, sourceContent, doc); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// isRelativePath checks if a given image or link contains a relative URL
// isRelativeLink checks if a link is relative
func isRelativeLink(link string) bool {
	return strings.HasPrefix(link, "../") || strings.HasPrefix(link, "#")
}
