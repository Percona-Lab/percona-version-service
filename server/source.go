package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func getData(product string, operatorVersion string) (*pbVersion.VersionResponse, error) {
	source := fmt.Sprintf("operator.%s.%s.json", operatorVersion, product)
	v, ok := data[source]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such source file: %s", source)
	}

	data := &pbVersion.VersionResponse{}
	err := jsonpb.Unmarshal(bytes.NewBuffer(v), data)
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
