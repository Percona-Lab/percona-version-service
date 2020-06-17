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

	err = filter(vs.Versions[0].Matrix.Pxc, req.Apply, req.DatabaseVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to filter versions: %v", err)
	}
	err = filter(vs.Versions[0].Matrix.Proxysql, "latest", req.DatabaseVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to filter versions: %v", err)
	}
	err = filter(vs.Versions[0].Matrix.Pmm, "latest", req.DatabaseVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to filter versions: %v", err)
	}
	err = filter(vs.Versions[0].Matrix.Backup, "latest", req.DatabaseVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to filter versions: %v", err)
	}

	return vs, nil
}
