package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	pbExample "github.com/Percona-Lab/percona-version-service/proto"
)

func parse(product string, operatorVersion string) (*pbExample.VersionResponse, error) {
	vs := &pbExample.VersionResponse{}
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

func filter(versions map[string]*pbExample.Version, apply string, current string) error {
	sorted, err := sortedVersions(versions)
	if err != nil {
		return fmt.Errorf("failed to sort versions: %v", err)
	}

	if (strings.ToLower(apply) == "recommended" && current == "") || strings.ToLower(apply) == "latest" {
		desired := sorted[0]
		deleteOtherBut(desired, versions)
		return nil
	}
	switch strings.ToLower(apply) {
	case "recommended":
	default:
		//assume version number
		deleteOtherBut(apply, versions)
		if len(versions) == 0 {
			return fmt.Errorf("version %s does not exist", apply)
		}

		return nil
	}

	return nil
}

func deleteOtherBut(v string, versions map[string]*pbExample.Version) {
	for k := range versions {
		if k != v {
			delete(versions, k)
		}
	}
}

func sortedVersions(versions map[string]*pbExample.Version) ([]string, error) {
	v := make([]*semver.Version, 0, len(versions))
	res := make([]string, 0, len(versions))

	for k := range versions {
		sv, err := semver.NewVersion(k)
		if err != nil {
			return nil, err
		}
		v = append(v, sv)
	}

	sort.Sort(sort.Reverse(semver.Collection(v)))
	for _, sv := range v {
		res = append(res, sv.String())
	}

	return res, nil
}
