package main

import (
	"fmt"
	"reflect"
	"slices"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"
	gover "github.com/hashicorp/go-version"

	"operator-tool/internal/matrix"
	"operator-tool/internal/util"
)

// deleteOldVersionsWithMap removes versions from the matrix that are older than those specified in the file.
func deleteOldVersions(file string, m *vsAPI.VersionMatrix) error {
	oldestVersions, err := getOldestVersions(file)
	if err != nil {
		return fmt.Errorf("failed to get oldest versions from base file: %w", err)
	}

	return matrix.Iterate(m, func(fieldName string, fieldValue reflect.Value) error {
		oldestVersion, ok := oldestVersions[fieldName]
		if !ok {
			return nil
		}

		m := fieldValue.Interface().(map[string]*vsAPI.Version)
		if len(m) == 0 {
			return nil
		}

		for k := range m {
			if util.Goversion(k).Compare(oldestVersion) < 0 {
				fieldValue.SetMapIndex(reflect.ValueOf(k), reflect.Value{}) // delete old version from map
			}
		}
		return nil
	})
}

// getOldestVersions returns a map where each key is a struct field name from the VersionMatrix
// of the specified file, and each value is the corresponding oldest version for that field.
func getOldestVersions(filePath string) (map[string]*gover.Version, error) {
	prod, err := util.ReadBaseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base file: %w", err)
	}

	versions := make(map[string]*gover.Version)
	if err := matrix.Iterate(prod.Versions[0].Matrix, func(fieldName string, fieldValue reflect.Value) error {
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
			if util.Goversion(oldestVersion).Compare(util.Goversion(k)) > 0 {
				oldestVersion = k
			}
		}
		versions[fieldName] = util.Goversion(oldestVersion)
		return nil
	}); err != nil {
		return nil, err
	}

	return versions, nil
}

func limitMajorVersions(m *vsAPI.VersionMatrix, capacity int) error {
	if capacity <= 0 {
		return nil
	}
	return matrix.Iterate(m, func(fieldName string, fieldValue reflect.Value) error {
		versionMap := fieldValue.Interface().(map[string]*vsAPI.Version)
		versionsByMajorVer := make(map[int][]string)
		for v := range versionMap {
			majorVer := util.Goversion(v).Segments()[0]
			versionsByMajorVer[majorVer] = append(versionsByMajorVer[majorVer], v)
		}
		for _, versions := range versionsByMajorVer {
			if len(versions) <= capacity {
				return nil
			}
			slices.SortFunc(versions, func(a, b string) int {
				return util.Goversion(b).Compare(util.Goversion(a))
			})

			versionsToDelete := versions[capacity:]
			for _, v := range versionsToDelete {
				fieldValue.SetMapIndex(reflect.ValueOf(v), reflect.Value{})
			}
		}

		return nil
	})
}

// updateMatrixStatuses updates the statuses of version maps.
// For each major version, it sets the highest version as "recommended"
// and all other versions as "available".
func updateMatrixStatuses(m *vsAPI.VersionMatrix) error {
	setStatus := func(vm map[string]*vsAPI.Version) {
		highestVersions := make(map[int]string)
		for version := range vm {
			vm[version].Status = vsAPI.Status_available

			majorVersion := util.Goversion(version).Segments()[0]

			currentHighestVersion, ok := highestVersions[majorVersion]

			if !ok || util.Goversion(version).Compare(util.Goversion(currentHighestVersion)) > 0 {
				highestVersions[majorVersion] = version
			}
		}

		for _, version := range highestVersions {
			vm[version].Status = vsAPI.Status_recommended
		}
	}

	return matrix.Iterate(m, func(fieldName string, fieldValue reflect.Value) error {
		versionMap := fieldValue.Interface().(map[string]*vsAPI.Version)
		if len(versionMap) == 0 {
			return nil
		}
		setStatus(versionMap)
		return nil
	})
}
