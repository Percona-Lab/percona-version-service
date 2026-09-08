package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"
	"google.golang.org/protobuf/encoding/protojson"

	"operator-tool/internal/matrix"
	"operator-tool/internal/util"
	"operator-tool/pkg/registry"
)

func patchProductResponse(rc *registry.RegistryClient, baseFilepath string, patchMatrix *vsAPI.VersionMatrix, operatorVersion string) (*vsAPI.ProductResponse, error) {
	baseFile, err := util.ReadBaseFile(baseFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base file: %w", err)
	}
	if err := updateMatrixData(rc, patchMatrix); err != nil {
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
	patchMatrixMap, err := matrixToMap(patchMatrix)
	if err != nil {
		return nil, fmt.Errorf("failed to convert patch matrix to map: %w", err)
	}

	for product, versions := range patchMatrixMap {
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
	if operatorVersion == "" {
		baseFile.Versions[0].Operator = operatorVersion
	}
	return baseFile, nil
}

func updateMatrixData(rc *registry.RegistryClient, m *vsAPI.VersionMatrix) error {
	return matrix.Iterate(m, func(fieldName string, fieldValue reflect.Value) error {
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
