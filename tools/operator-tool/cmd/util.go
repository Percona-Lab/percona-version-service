package main

import (
	"fmt"
	"os"
	"reflect"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"
	gover "github.com/hashicorp/go-version"
	"google.golang.org/protobuf/encoding/protojson"
)

func readBaseFile(path string) (*vsAPI.ProductResponse, error) {
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

// deleteOldVersionsWithMap removes versions from the matrix that are older than those specified in the file.
func deleteOldVersions(file string, matrix *vsAPI.VersionMatrix) error {
	minVersions, err := getOldestVersions(file)
	if err != nil {
		return fmt.Errorf("failed to get oldest versions from base file: %w", err)
	}
	deleteOldVersionsWithMap(matrix, minVersions)
	return nil
}

// deleteOldVersionsWithMap removes versions from the matrix that are older than those specified in oldestVersions.
func deleteOldVersionsWithMap(matrix *vsAPI.VersionMatrix, oldestVersions map[string]*gover.Version) {
	matrixType := reflect.TypeOf(matrix).Elem()
	matrixValue := reflect.ValueOf(matrix).Elem()

	for i := 0; i < matrixValue.NumField(); i++ {
		field := matrixType.Field(i)
		// check if value is exported
		if field.PkgPath != "" {
			continue
		}
		oldestVersion, ok := oldestVersions[field.Name]
		if !ok {
			continue
		}

		value := matrixValue.Field(i)

		m := value.Interface().(map[string]*vsAPI.Version)
		if len(m) == 0 {
			continue
		}

		for k := range m {
			if goversion(k).Compare(oldestVersion) < 0 {
				value.SetMapIndex(reflect.ValueOf(k), reflect.Value{}) // delete old version from map
			}
		}
	}
}

// getOldestVersions returns a map where each key is a struct field name from the VersionMatrix
// of the specified file, and each value is the corresponding oldest version for that field.
func getOldestVersions(filePath string) (map[string]*gover.Version, error) {
	prod, err := readBaseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base file: %w", err)
	}

	matrixType := reflect.TypeOf(prod.Versions[0].Matrix).Elem()
	matrixValue := reflect.ValueOf(prod.Versions[0].Matrix).Elem()

	versions := make(map[string]*gover.Version)
	for i := 0; i < matrixValue.NumField(); i++ {
		field := matrixType.Field(i)
		// ignore if value is not exported
		if field.PkgPath != "" {
			continue
		}
		versionMapValue := matrixValue.Field(i)

		versionMap := versionMapValue.Interface().(map[string]*vsAPI.Version)
		if len(versionMap) == 0 {
			continue
		}
		oldestVersion := ""
		for k := range versionMap {
			if oldestVersion == "" {
				oldestVersion = k
				continue
			}
			if goversion(oldestVersion).Compare(goversion(k)) > 0 {
				oldestVersion = k
			}
		}
		versions[field.Name] = goversion(oldestVersion)
	}

	return versions, nil
}

func goversion(v string) *gover.Version {
	return gover.Must(gover.NewVersion(v))
}
