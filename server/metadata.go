package server

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb/api"
	"github.com/bufbuild/protoyaml-go"
	"google.golang.org/protobuf/encoding/protojson"
)

type Metadata struct {
	data map[string]*pbVersion.MetadataResponse
	fs   fs.FS
}

func NewMetadata(fs fs.FS) (*Metadata, error) {
	m := &Metadata{
		fs: fs,
	}
	err := m.readAll()
	return m, err
}

func (m *Metadata) Product(product string) (*pbVersion.MetadataResponse, error) {
	res, ok := m.data[product]
	if !ok {
		return &pbVersion.MetadataResponse{}, nil
	}

	return res, nil
}

func (m *Metadata) readAll() error {
	m.data = make(map[string]*pbVersion.MetadataResponse)

	files, err := fs.ReadDir(m.fs, ".")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		d, err := m.getAllMetadataFromFiles(f.Name())
		if err != nil {
			return err
		}
		m.data[f.Name()] = &pbVersion.MetadataResponse{Versions: d}
	}

	return nil
}

func (m *Metadata) getAllMetadataFromFiles(product string) ([]*pbVersion.MetadataVersion, error) {
	if !filepath.IsLocal(product) {
		return nil, errors.New("product name is invalid")
	}

	dir := filepath.Join(".", product)
	files, err := fs.ReadDir(m.fs, dir)
	if err != nil {
		return nil, errors.Join(err, errors.New("could not read metadata from directory"))
	}

	ret := make([]*pbVersion.MetadataVersion, 0, len(files))
	for _, f := range files {
		p := filepath.Join(dir, f.Name())
		c, err := fs.ReadFile(m.fs, p)
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("could not read file %s", p))
		}

		metaV, err := m.parseFile(c, filepath.Ext(f.Name()))
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("could not parse file %s", f.Name()))
		}
		ret = append(ret, metaV)
	}

	return ret, nil
}

func (m *Metadata) parseFile(c []byte, fileExt string) (*pbVersion.MetadataVersion, error) {
	meta := &pbVersion.MetadataVersion{}
	switch fileExt {
	case ".yaml", ".yml":
		if err := protoyaml.Unmarshal(c, meta); err != nil {
			return nil, errors.Join(err, errors.New("could not unmarshal yaml"))
		}

	case ".json":
		if err := protojson.Unmarshal(c, meta); err != nil {
			return nil, errors.Join(err, errors.New("could not unmarshal json"))
		}
	default:
		return nil, fmt.Errorf("extension %s not supported", fileExt)
	}

	return meta, nil
}
