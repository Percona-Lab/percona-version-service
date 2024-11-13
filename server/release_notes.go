package server

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

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
