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
	v1Data map[string]*pbVersion.MetadataResponse
	v2Data map[string]*pbVersion.MetadataV2Response
	fs     fs.FS
}

func NewMetadata(fs fs.FS) (*Metadata, error) {
	m := &Metadata{
		fs: fs,
	}
	err := m.readAll()
	return m, err
}

func (m *Metadata) Product(product string) (*pbVersion.MetadataResponse, error) {
	res, ok := m.v1Data[product]
	if !ok {
		return &pbVersion.MetadataResponse{}, nil
	}

	return res, nil
}

func (m *Metadata) ProductV2(product string) (*pbVersion.MetadataV2Response, error) {
	res, ok := m.v2Data[product]
	if !ok {
		return &pbVersion.MetadataV2Response{}, nil
	}

	return res, nil
}

func (m *Metadata) readAll() error {
	m.v1Data = make(map[string]*pbVersion.MetadataResponse)
	m.v2Data = make(map[string]*pbVersion.MetadataV2Response)

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
		v1Data, v2Data, err := m.getAllMetadataFromFiles(f.Name())
		if err != nil {
			return err
		}
		if len(v1Data) > 0 {
			m.v1Data[f.Name()] = &pbVersion.MetadataResponse{Versions: v1Data}
		} else if len(v2Data) > 0 {
			m.v2Data[f.Name()] = &pbVersion.MetadataV2Response{Versions: v2Data}
		}
	}
	return nil
}

func (m *Metadata) getAllMetadataFromFiles(product string) ([]*pbVersion.MetadataVersion, []*pbVersion.MetadataV2Version, error) {
	if !filepath.IsLocal(product) {
		return nil, nil, errors.New("product name is invalid")
	}

	dir := filepath.Join(".", product)
	files, err := fs.ReadDir(m.fs, dir)
	if err != nil {
		return nil, nil, errors.Join(err, errors.New("could not read metadata from directory"))
	}
	v1Data := make([]*pbVersion.MetadataVersion, 0, len(files))
	v2Data := make([]*pbVersion.MetadataV2Version, 0, len(files))
	for _, f := range files {
		p := filepath.Join(dir, f.Name())
		c, err := fs.ReadFile(m.fs, p)
		if err != nil {
			return nil, nil, errors.Join(err, fmt.Errorf("could not read file %s", p))
		}

		metaV, err := m.parseFile(c, filepath.Ext(f.Name()))
		if err != nil {
			return nil, nil, errors.Join(err, fmt.Errorf("could not parse file %s", f.Name()))
		}
		if metaV.ImageInfo == nil {
			v1Data = append(v1Data, &pbVersion.MetadataVersion{
				Version:     metaV.Version,
				Recommended: metaV.Recommended,
				Supported:   metaV.Supported,
			})
		} else {
			v2Data = append(v2Data, metaV)
		}
	}
	return v1Data, v2Data, nil
}

func (m *Metadata) parseFile(c []byte, fileExt string) (*pbVersion.MetadataV2Version, error) {
	meta := &pbVersion.MetadataV2Version{}
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
