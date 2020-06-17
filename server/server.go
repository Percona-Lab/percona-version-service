package server

import (
	"context"
	"fmt"

	pbExample "github.com/Percona-Lab/percona-version-service/proto"
)

// Backend implements the protobuf interface
type Backend struct {
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{}
}

func (b *Backend) Apply(ctx context.Context, req *pbExample.ApplyRequest) (*pbExample.VersionResponse, error) {
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
