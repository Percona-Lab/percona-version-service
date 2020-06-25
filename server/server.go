package server

import (
	"context"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb"
)

// Backend implements the protobuf interface.
type Backend struct {
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{}
}

func (b *Backend) Apply(ctx context.Context, req *pbVersion.ApplyRequest) (*pbVersion.VersionResponse, error) {
	vs, err := getData(req.Product, req.OperatorVersion)
	if err != nil {
		return nil, err
	}

	deps, err := getDep(req.Product, req.OperatorVersion)
	if err != nil {
		return nil, err
	}

	err = pxcFilter(vs.Versions[0].Matrix.Pxc, req.Apply, req.DatabaseVersion)
	if err != nil {
		return nil, err
	}

	productVersion := ""
	for k := range vs.Versions[0].Matrix.Pxc {
		productVersion = k
		break
	}

	backupVersion, err := depFilter(deps.Backup, productVersion)
	if err != nil {
		return nil, err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Backup, backupVersion)
	if err != nil {
		return nil, err
	}

	pmmVersion, err := depFilter(deps.PMM, productVersion)
	if err != nil {
		return nil, err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Pmm, pmmVersion)
	if err != nil {
		return nil, err
	}

	proxySQL, err := depFilter(deps.ProxySQL, productVersion)
	if err != nil {
		return nil, err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Proxysql, proxySQL)
	if err != nil {
		return nil, err
	}

	return vs, nil
}
