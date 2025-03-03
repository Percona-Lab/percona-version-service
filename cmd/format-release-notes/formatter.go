package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/Kunde21/markdownfmt/v3/markdown"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// createMarkdownRender creates a new goldmark.Markdown renderer that allows us re-format markdown files.
func createMarkdownRenderer(opts ...markdown.Option) goldmark.Markdown {
	mr := markdown.NewRenderer()
	mr.AddMarkdownOptions(opts...)
	extensions := []goldmark.Extender{
		extension.GFM,
		meta.Meta,
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
// - custom icon variables are changed to their SVG/HTML equivalent based on the iconsMap specified in variables.go.
// - Admonitions are transformed to headings.
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
	// add an extra slash to the URL, else the last path will get removed by url.ResolveReference()
	docsURLPrefix, err := url.Parse("https://docs.percona.com/percona-monitoring-and-management/3//")
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	if err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if link, ok := node.(*ast.Link); ok && entering {
			dest := string(link.Destination)
			if target, isRelativeLink := extractRelativeURL(dest); isRelativeLink {
				newDestination := docsURLPrefix.ResolveReference(target).String()
				if strings.HasSuffix(newDestination, ".md") {
					newDestination = strings.TrimSuffix(newDestination, ".md") + ".html"
				}
				link.Destination = []byte(newDestination)
			}
		} else if image, ok := node.(*ast.Image); ok && entering {
			dest := string(image.Destination)
			if target, isRelativeLink := extractRelativeURL(dest); isRelativeLink {
				newDestination := docsURLPrefix.ResolveReference(target)
				image.Destination = []byte(newDestination.String())
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

func extractRelativeURL(link string) (*url.URL, bool) {
	target, err := url.Parse(link)
	if err != nil {
		log.Println(err)
		return nil, false
	}
	return target, !target.IsAbs()
}
