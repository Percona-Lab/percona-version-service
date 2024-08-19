package server

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Kunde21/markdownfmt/v3/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb/api"
)

type ReleaseNotes struct {
	// releaseNotes is a map of PMM versions to release notes.
	pmmReleaseNotes  map[string]*pbVersion.GetReleaseNotesResponse
	releaseNotesLock sync.Mutex
	fs               fs.FS
}

func NewReleaseNotes(fs fs.FS) *ReleaseNotes {
	return &ReleaseNotes{
		pmmReleaseNotes: make(map[string]*pbVersion.GetReleaseNotesResponse),
		fs:              fs,
	}
}

func (r *ReleaseNotes) getVersionsForProduct(product string) map[string]*pbVersion.GetReleaseNotesResponse {
	switch product {
	case "pmm":
		if r.pmmReleaseNotes == nil {
			r.pmmReleaseNotes = make(map[string]*pbVersion.GetReleaseNotesResponse)
		}
		return r.pmmReleaseNotes
	default:
		return nil
	}
}

func (r *ReleaseNotes) GetReleaseNote(product, version string) (*pbVersion.GetReleaseNotesResponse, error) {
	r.releaseNotesLock.Lock()
	defer r.releaseNotesLock.Unlock()

	availableVersions := r.getVersionsForProduct(product)
	if availableVersions == nil {
		return nil, errors.New(fmt.Sprintf("%s is not a valid product", product))
	}

	if notes, ok := availableVersions[version]; ok {
		return notes, nil
	}

	rn, err := r.refreshReleaseNotes(product, version)
	if err != nil {
		return nil, err
	}
	availableVersions[version] = rn
	if notes, ok := availableVersions[version]; ok {
		return notes, nil
	}

	return nil, ErrNotFound
}

func (r *ReleaseNotes) refreshReleaseNotes(product, version string) (*pbVersion.GetReleaseNotesResponse, error) {
	if !filepath.IsLocal(product) {
		return nil, errors.New("product name is invalid")
	}

	dir := filepath.Join(".", product)
	rnName := filepath.Join(dir, version+".md")
	rnFile, err := fs.ReadFile(r.fs, rnName)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("could not read file %s", rnName))
	}
	return &pbVersion.GetReleaseNotesResponse{
		Version:     version,
		Product:     product,
		ReleaseNote: string(rnFile),
	}, nil
}

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

// TransformReleaseNoteLinks rewrites relative links in markdown files to absolute links.
func TransformReleaseNoteLinks(sourceContent []byte) ([]byte, error) {
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
