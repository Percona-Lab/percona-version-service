package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
)

type flags struct {
	Dir string `default:"sources/release-notes/pmm" help:"Directory where target markdown files are stored"`
}

func main() {
	var opts flags
	kong.Parse(
		&opts,
		kong.Name("format-release-notes"),
		kong.Description("Formats the markdown source of the available release notes."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	if err := formatReleaseNotes(opts.Dir); err != nil {
		log.Fatalf("failed to update relative paths %+v", err)
	}
}

// formatReleaseNotes formats the markdown source of the available release notes.
func formatReleaseNotes(dir string) error {
	root := os.DirFS(dir)
	matches, err := fs.Glob(root, "*.md")
	log.Println("Using path: ", dir)
	if err != nil {
		return err
	}

	for _, file := range matches {
		log.Printf("Found markdown file: %s in path: %s", file, dir)
		b, err := fs.ReadFile(root, file) //nolint:gosec
		if err != nil {
			return err
		}

		output, err := FormatReleaseNotes(b)
		if err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(dir, file), output, 0644); err != nil {
			return err
		}
		log.Printf("Processed markdown file: %s in path: %s", file, dir)
	}
	return nil
}
