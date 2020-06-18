package server

import (
	"context"
	"fmt"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb"
)

// Backend implements the protobuf interface
type Backend struct {
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{}
}

func (b *Backend) Apply(ctx context.Context, req *pbVersion.ApplyRequest) (*pbVersion.VersionResponse, error) {
	vs, err := parse(req.Product, req.OperatorVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %v", err)
	}

	err = pxcFilter(vs.Versions[0].Matrix.Pxc, req.Apply, req.DatabaseVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to filter versions: %v", err)
	}

	//TODO: pxc filter used here
	//the reason is I have no info how to filter deps versions
	//so it returns latest version
	err = pxcFilter(vs.Versions[0].Matrix.Proxysql, "latest", "")
	if err != nil {
		return nil, fmt.Errorf("failed to filter versions: %v", err)
	}
	err = pxcFilter(vs.Versions[0].Matrix.Pmm, "latest", "")
	if err != nil {
		return nil, fmt.Errorf("failed to filter versions: %v", err)
	}
	err = pxcFilter(vs.Versions[0].Matrix.Backup, "latest", "")
	if err != nil {
		return nil, fmt.Errorf("failed to filter versions: %v", err)
	}

	return vs, nil
}
