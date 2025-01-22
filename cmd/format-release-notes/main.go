package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin/v2"
)

func main() {
	app := kingpin.New("format-release-notes", "Modifies release notes by replacing relative links with absolute ones")
	markdownDir := app.Flag("dir", "Directory where target markdown files are stored").Default("sources/release-notes/pmm").String()

	if _, err := app.Parse(os.Args[1:]); err != nil {
		log.Fatalf("failed to parse command %+v", err)
	}

	if err := formatReleaseNotes(*markdownDir); err != nil {
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
