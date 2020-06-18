package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb"
)

func parse(product string, operatorVersion string) (*pbVersion.VersionResponse, error) {
	vs := &pbVersion.VersionResponse{}
	source := fmt.Sprintf("operator.%s.%s.json", operatorVersion, product)

	content, err := ioutil.ReadFile("./sources/" + source)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions source file: %v", err)
	}

	err = json.Unmarshal(content, vs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal content: %v", err)
	}

	return vs, nil
}

func pxcFilter(versions map[string]*pbVersion.Version, apply string, current string) error {
	if len(versions) == 0 {
		return fmt.Errorf("no versions to filter")
	}

	sorted, err := sortedVersionsDesc(versions)
	if err != nil {
		return fmt.Errorf("failed to sort versions: %v", err)
	}

	if (strings.ToLower(apply) == "recommended" || strings.ToLower(apply) == "latest") && current == "" {
		deleteOtherBut(sorted[0].String(), versions)
		return nil
	}

	desired := apply //assume version number
	if strings.ToLower(apply) == "recommended" || strings.ToLower(apply) == "latest" {
		desired = current

		c, err := semver.NewVersion(current)
		if err != nil {
			return fmt.Errorf("invalid current version: %s", current)
		}

		for _, s := range sorted {
			if s.Equal(c) || s.LessThan(c) {
				break
			}
			if versions[s.String()].Status != "disabled" && c.Major() == s.Major() {
				desired = s.String()
			}
		}
	}

	deleteOtherBut(desired, versions)
	if len(versions) == 0 {
		return fmt.Errorf("version %s does not exist", desired)
	}

	return nil
}

func deleteOtherBut(v string, versions map[string]*pbVersion.Version) {
	for k := range versions {
		if k != v {
			delete(versions, k)
		}
	}
}

func sortedVersionsDesc(versions map[string]*pbVersion.Version) ([]*semver.Version, error) {
	v := make([]*semver.Version, 0, len(versions))

	for k := range versions {
		sv, err := semver.NewVersion(k)
		if err != nil {
			return nil, err
		}
		v = append(v, sv)
	}

	sort.Sort(sort.Reverse(semver.Collection(v)))

	return v, nil
}
