package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func parse(product string, operatorVersion string) (*pbVersion.VersionResponse, error) {
	vs := &pbVersion.VersionResponse{}
	source := fmt.Sprintf("operator.%s.%s.json", operatorVersion, product)

	content, err := ioutil.ReadFile("./sources/" + source)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read versions source file: %v", err)
	}

	err = json.Unmarshal(content, vs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal source file content: %v", err)
	}

	return vs, nil
}

func parseDep(product string, operatorVersion string) (Deps, error) {
	deps := Deps{}
	source := fmt.Sprintf("operator.%s.%s.dep.json", operatorVersion, product)

	content, err := ioutil.ReadFile("./sources/" + source)
	if err != nil {
		return Deps{}, status.Errorf(codes.Internal, "failed to read versions source file: %v", err)
	}

	err = json.Unmarshal(content, &deps)
	if err != nil {
		return Deps{}, status.Errorf(codes.Internal, "failed to unmarshal source file content: %v", err)
	}

	return deps, nil
}
