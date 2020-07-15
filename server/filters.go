package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb"
	"github.com/diegoholiveira/jsonlogic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	recommended = "recommended"
	latest      = "latest"
)

func psmdbFilter(versions map[string]*pbVersion.Version, apply string, current string) error {
	// at that moment logic is identical
	return pxcFilter(versions, apply, current)
}

func pxcFilter(versions map[string]*pbVersion.Version, apply string, current string) error {
	if len(versions) == 0 {
		return status.Error(codes.Internal, "no versions to filter")
	}

	keys := make([]string, 0, len(versions))
	for k, v := range versions {
		if strings.ToLower(apply) == recommended && v.Status != pbVersion.Status_recommended {
			continue
		}

		keys = append(keys, k)
	}

	sorted, err := sortedVersionsDesc(keys)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to sort versions: %v", err)
	}

	if (strings.ToLower(apply) == recommended || strings.ToLower(apply) == latest) && current == "" {
		return deleteOtherBut(sorted[0].String(), versions)
	}

	desired := apply //assume version number
	if strings.ToLower(apply) == recommended || strings.ToLower(apply) == latest {
		desired = current

		c, err := semver.NewVersion(current)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid current version: %s", current)
		}

		for _, s := range sorted {
			if s.Equal(c) || s.LessThan(c) {
				break
			}

			if versions[s.String()].Status != pbVersion.Status_disabled && c.Major() == s.Major() && c.Minor() == s.Minor() {
				desired = s.String()
				if strings.ToLower(apply) == latest {
					break
				}
			}
		}
	}

	err = deleteOtherBut(desired, versions)
	if err != nil {
		return err
	}
	if len(versions) == 0 {
		return status.Errorf(codes.NotFound, "version %s does not exist", desired)
	}

	return nil
}

func defaultFilter(versions map[string]*pbVersion.Version, apply string) error {
	if len(versions) == 0 {
		return status.Error(codes.Internal, "no versions to filter")
	}

	keys := make([]string, 0, len(versions))
	for k := range versions {
		keys = append(keys, k)
	}

	sorted, err := sortedVersionsDesc(keys)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to sort versions: %v", err)
	}

	if apply == "" {
		apply = sorted[0].String()
	}

	return deleteOtherBut(apply, versions)
}

func depFilter(versions map[string]interface{}, productVersion string) (string, error) {
	if len(versions) == 0 {
		return "", nil
	}

	keys := make([]string, 0, len(versions))
	for k := range versions {
		keys = append(keys, k)
	}

	sorted, err := sortedVersionsDesc(keys)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to sort versions: %v", err)
	}

	desired := sorted[0].String()
	for _, s := range sorted {
		b, err := json.Marshal(versions[s.String()])
		if err != nil {
			return "", status.Errorf(codes.Internal, "failed to marshal deps logic: %v", err)
		}
		logic := bytes.NewReader(b)
		data := strings.NewReader(fmt.Sprintf(`{  "productVersion" : "%s" }`, productVersion))

		var result bytes.Buffer
		err = jsonlogic.Apply(logic, data, &result)
		if err != nil {
			return "", status.Errorf(codes.Internal, "failed to apply logic: %v", err)
		}

		if strings.TrimSuffix(result.String(), "\n") == "true" {
			desired = s.String()
			break
		}
	}

	return desired, nil
}

func deleteOtherBut(v string, versions map[string]*pbVersion.Version) error {
	sv, err := semver.NewVersion(v)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to parse version: %s", v)
	}

	for k := range versions {
		sk, err := semver.NewVersion(k)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to parse version: %s", k)
		}

		// ignore prerelease/buildmetadata suffix
		if sk.Major() != sv.Major() && sk.Minor() != sv.Minor() && sk.Patch() != sv.Patch() {
			delete(versions, k)
		}
	}

	// check situation when there are more than 1 version with same major.minor.patch in source file
	// in such case do not ignore prerelease/buildmetadata
	if len(versions) > 1 {
		for k := range versions {
			if k != v {
				delete(versions, k)
			}
		}
	}

	return nil
}

func sortedVersionsDesc(versions []string) ([]*semver.Version, error) {
	v := make([]*semver.Version, 0, len(versions))

	for _, k := range versions {
		sv, err := semver.NewVersion(k)
		if err != nil {
			return nil, err
		}
		v = append(v, sv)
	}

	sort.Sort(sort.Reverse(semver.Collection(v)))

	return v, nil
}
