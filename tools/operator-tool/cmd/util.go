package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"
	gover "github.com/hashicorp/go-version"
	"google.golang.org/protobuf/encoding/protojson"

	"operator-tool/registry"
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
	matrixType := reflect.TypeOf(matrix).Elem()
	matrixValue := reflect.ValueOf(matrix).Elem()

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
	}
	return nil
}

func updateMatrixStatuses(matrix *vsAPI.VersionMatrix) error {
	matrixType := reflect.TypeOf(matrix).Elem()
	matrixValue := reflect.ValueOf(matrix).Elem()

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
		setStatus(versionMap)
	}
	return nil
}

// setStatus sets a recommended status to the latest version and available status to other versions.
func setStatus(vm map[string]*vsAPI.Version) {
	maxVer := ""
	for k := range vm {
		vm[k].Status = vsAPI.Status_available
		if maxVer == "" {
			maxVer = k
			continue
		}

		if goversion(k).Compare(goversion(maxVer)) > 0 {
			maxVer = k
		}
	}

	if _, ok := vm[maxVer]; ok {
		vm[maxVer].Status = vsAPI.Status_recommended
	}
}
