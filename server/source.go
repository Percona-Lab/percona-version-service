package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

var data = map[string][]byte{}
var deps = map[string][]byte{}

func init() {
	files, err := ioutil.ReadDir("./sources")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fname := file.Name()
		content, err := ioutil.ReadFile(path.Join("./sources", fname))
		if err != nil {
			log.Fatalf("failed to read source file: %v", err)
		}

		if strings.HasSuffix(fname, ".dep.json") {
			deps[fname] = content
			continue
		}

		data[fname] = content
	}
}

const pmmServerSuffix = ".pmm-server.json"

func operatorData(product string) (*pbVersion.ProductResponse, error) {
	suffix := fmt.Sprintf(".%s.json", product)
	r := pbVersion.ProductResponse{}

	for k, v := range data {
		if strings.HasSuffix(k, suffix) {
			if r.Versions == nil {
				err := protojson.Unmarshal(v, &r)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "failed to unmarshal source file: %v", err)
				}
			} else {
				pr := pbVersion.ProductResponse{}

				err := protojson.Unmarshal(v, &pr)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "failed to unmarshal source file: %v", err)
				}

				r.Versions = append(r.Versions, pr.Versions...)
			}
		}
	}

	return &r, nil
}

func operatorProductData(team string, product string, version string) (*pbVersion.VersionResponse, error) {
	source := fmt.Sprintf("%s.%s.%s.json", team, version, product)
	v, ok := data[source]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such source file: %s", source)
	}

	data := &pbVersion.VersionResponse{}
	err := protojson.Unmarshal(v, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal source file: %v", err)
	}

	return data, nil
}

func getDep(product string, operatorVersion string) (Deps, error) {
	source := fmt.Sprintf("operator.%s.%s.dep.json", operatorVersion, product)
	v, ok := deps[source]
	if !ok {
		return Deps{}, status.Errorf(codes.NotFound, "no such source file: %s", source)
	}

	dep := Deps{}
	err := json.Unmarshal(v, &dep)
	if err != nil {
		return Deps{}, status.Errorf(codes.Internal, "failed to unmarshal source file: %v", err)
	}
	return dep, nil
}
