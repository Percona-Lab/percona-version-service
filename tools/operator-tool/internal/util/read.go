package util

import (
	"fmt"
	"os"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"
	"google.golang.org/protobuf/encoding/protojson"
)

func ReadBaseFile(path string) (*vsAPI.ProductResponse, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	product := new(vsAPI.ProductResponse)
	err = protojson.Unmarshal(content, product)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return product, nil
}

func ReadPatchFile(path string) (*vsAPI.VersionMatrix, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	matrix := new(vsAPI.VersionMatrix)
	err = protojson.Unmarshal(content, matrix)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return matrix, nil
}
