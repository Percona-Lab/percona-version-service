package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"
	gover "github.com/hashicorp/go-version"
	"google.golang.org/protobuf/encoding/protojson"

	"operator-tool/pkg/registry"
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

func readPatchFile(path string) (*vsAPI.VersionMatrix, error) {
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

// deleteOldVersionsWithMap removes versions from the matrix that are older than those specified in the file.
func deleteOldVersions(file string, matrix *vsAPI.VersionMatrix) error {
	minVersions, err := getOldestVersions(file)
	if err != nil {
		return fmt.Errorf("failed to get oldest versions from base file: %w", err)
	}
	deleteOldVersionsWithMap(matrix, minVersions)
	return nil
}

func iterateOverMatrixFields(matrix *vsAPI.VersionMatrix, f func(fieldName string, fieldValue reflect.Value) error) error {
	matrixType := reflect.TypeOf(matrix).Elem()
	matrixValue := reflect.ValueOf(matrix).Elem()

	for i := 0; i < matrixValue.NumField(); i++ {
		field := matrixType.Field(i)
		// check if value is exported
		if field.PkgPath != "" {
			continue
		}
		if err := f(field.Name, matrixValue.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

// deleteOldVersionsWithMap removes versions from the matrix that are older than those specified in oldestVersions.
func deleteOldVersionsWithMap(matrix *vsAPI.VersionMatrix, oldestVersions map[string]*gover.Version) {
	iterateOverMatrixFields(matrix, func(fieldName string, fieldValue reflect.Value) error {
		oldestVersion, ok := oldestVersions[fieldName]
		if !ok {
			return nil
		}

		m := fieldValue.Interface().(map[string]*vsAPI.Version)
		if len(m) == 0 {
			return nil
		}

		for k := range m {
			if goversion(k).Compare(oldestVersion) < 0 {
				fieldValue.SetMapIndex(reflect.ValueOf(k), reflect.Value{}) // delete old version from map
			}
		}
		return nil
	})
}

// getOldestVersions returns a map where each key is a struct field name from the VersionMatrix
// of the specified file, and each value is the corresponding oldest version for that field.
func getOldestVersions(filePath string) (map[string]*gover.Version, error) {
	prod, err := readBaseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base file: %w", err)
	}

	versions := make(map[string]*gover.Version)
	iterateOverMatrixFields(prod.Versions[0].Matrix, func(fieldName string, fieldValue reflect.Value) error {
		versionMap := fieldValue.Interface().(map[string]*vsAPI.Version)
		if len(versionMap) == 0 {
			return nil
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
		versions[fieldName] = goversion(oldestVersion)
		return nil
	})

	return versions, nil
}

func goversion(v string) *gover.Version {
	return gover.Must(gover.NewVersion(v))
}

func patchProductResponse(rc *registry.RegistryClient, baseFilepath, patchFilepath string) (*vsAPI.ProductResponse, error) {
	baseFile, err := readBaseFile(baseFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base file: %w", err)
	}
	patchFile, err := readPatchFile(patchFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read patch file: %w", err)
	}
	if err := updateMatrixHashes(rc, patchFile); err != nil {
		return nil, fmt.Errorf("failed to update patch matrix hashes: %w", err)
	}

	matrixToMap := func(matrix *vsAPI.VersionMatrix) (map[string]map[string]map[string]any, error) {
		data, err := protojson.Marshal(matrix)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal: %w", err)
		}

		m := make(map[string]map[string]map[string]any)
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal: %w", err)
		}
		return m, nil
	}

	baseMatrix, err := matrixToMap(baseFile.Versions[0].Matrix)
	if err != nil {
		return nil, fmt.Errorf("failed to convert base matrix to map: %w", err)
	}
	patchMatrix, err := matrixToMap(patchFile)
	if err != nil {
		return nil, fmt.Errorf("failed to convert patch matrix to map: %w", err)
	}

	for product, versions := range patchMatrix {
		for version, verInfo := range versions {
			if _, ok := baseMatrix[product]; !ok {
				baseMatrix[product] = make(map[string]map[string]any)
			}
			baseMatrix[product][version] = verInfo
		}
	}

	mapToMatrix := func(m map[string]map[string]map[string]any) (*vsAPI.VersionMatrix, error) {
		data, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal: %w", err)
		}

		matrix := new(vsAPI.VersionMatrix)
		if err := protojson.Unmarshal(data, matrix); err != nil {
			return nil, fmt.Errorf("failed to unmarshal: %w", err)
		}

		return matrix, nil
	}

	baseFile.Versions[0].Matrix, err = mapToMatrix(baseMatrix)
	if err != nil {
		return nil, fmt.Errorf("failed to convert patched map to matrix: %w", err)
	}
	return baseFile, nil
}

func updateMatrixHashes(rc *registry.RegistryClient, matrix *vsAPI.VersionMatrix) error {
	return iterateOverMatrixFields(matrix, func(fieldName string, fieldValue reflect.Value) error {
		versionMap := fieldValue.Interface().(map[string]*vsAPI.Version)
		if len(versionMap) == 0 {
			return nil
		}

		for k, v := range versionMap {
			imageSpl := strings.Split(v.ImagePath, ":")
			if len(imageSpl) == 1 {
				return fmt.Errorf("image %s doesn't have tag", v.ImagePath)
			}
			tag := imageSpl[len(imageSpl)-1]
			imageName := strings.TrimSuffix(v.ImagePath, ":"+tag)
			image, err := rc.GetTag(imageName, tag)
			if err != nil {
				return fmt.Errorf("failed to get tag %s for image %s: %w", tag, imageName, err)
			}
			versionMap[k].ImageHash = image.DigestAMD64
			versionMap[k].ImageHashArm64 = image.DigestARM64
		}
		return nil
	})
}

func limitMajorVersions(matrix *vsAPI.VersionMatrix, capacity int) error {
	if capacity <= 0 {
		return nil
	}
	return iterateOverMatrixFields(matrix, func(fieldName string, fieldValue reflect.Value) error {
		versionMap := fieldValue.Interface().(map[string]*vsAPI.Version)
		versionsByMajorVer := make(map[int][]string)
		for v := range versionMap {
			majorVer := goversion(v).Segments()[0]
			versionsByMajorVer[majorVer] = append(versionsByMajorVer[majorVer], v)
		}
		for _, versions := range versionsByMajorVer {
			if len(versions) <= capacity {
				return nil
			}
			slices.SortFunc(versions, func(a, b string) int {
				return goversion(b).Compare(goversion(a))
			})

			versionsToDelete := versions[capacity:]
			for _, v := range versionsToDelete {
				fieldValue.SetMapIndex(reflect.ValueOf(v), reflect.Value{})
			}
		}

		return nil
	})
}

func updateMatrixStatuses(matrix *vsAPI.VersionMatrix) error {
	return iterateOverMatrixFields(matrix, func(fieldName string, fieldValue reflect.Value) error {
		versionMap := fieldValue.Interface().(map[string]*vsAPI.Version)
		if len(versionMap) == 0 {
			return nil
		}
		setStatus(versionMap)
		return nil
	})
}

// setStatus updates the statuses of version map.
// For each major version, it sets the highest version as "recommended"
// and all other versions as "available".
func setStatus(vm map[string]*vsAPI.Version) {
	highestVersions := make(map[int]string)
	for version := range vm {
		vm[version].Status = vsAPI.Status_available

		majorVersion := goversion(version).Segments()[0]

		currentHighestVersion, ok := highestVersions[majorVersion]

		if !ok || goversion(version).Compare(goversion(currentHighestVersion)) > 0 {
			highestVersions[majorVersion] = version
		}
	}

	for _, version := range highestVersions {
		vm[version].Status = vsAPI.Status_recommended
	}
}