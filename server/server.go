package server

import (
	"context"
	"strings"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Backend implements the protobuf interface.
type Backend struct {
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{}
}

func (b *Backend) Product(ctx context.Context, req *pbVersion.ProductRequest) (*pbVersion.ProductResponse, error) {
	return operatorData(req.Product)
}

func (b *Backend) Operator(ctx context.Context, req *pbVersion.OperatorRequest) (*pbVersion.OperatorResponse, error) {
	vs, err := operatorProductData(req.Product, req.OperatorVersion)
	if err != nil {
		return nil, err
	}

	return &pbVersion.OperatorResponse{
		Versions: vs.Versions,
	}, nil
}

func (b *Backend) Apply(_ context.Context, req *pbVersion.ApplyRequest) (*pbVersion.VersionResponse, error) {
	err := validate(req)
	if err != nil {
		return nil, err
	}

	vs, err := operatorProductData(req.Product, req.OperatorVersion)
	if err != nil {
		return nil, err
	}

	deps, err := getDep(req.Product, req.OperatorVersion)
	if err != nil {
		return nil, err
	}

	switch req.Product {
	case "pxc-operator":
		err := pxc(vs, deps, req)
		if err != nil {
			return nil, err
		}
	case "psmdb-operator":
		err := psmdb(vs, deps, req)
		if err != nil {
			return nil, err
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid product: %s", req.Product)
	}

	return vs, nil
}

func pxc(vs *pbVersion.VersionResponse, deps Deps, req *pbVersion.ApplyRequest) error {
	err := pxcFilter(vs.Versions[0].Matrix.Pxc, req.Apply, req.DatabaseVersion)
	if err != nil {
		return err
	}

	productVersion := ""
	for k := range vs.Versions[0].Matrix.Pxc {
		productVersion = k
		break
	}

	backupVersion, err := depFilter(deps.Backup, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Backup, backupVersion)
	if err != nil {
		return err
	}

	pmmVersion, err := depFilter(deps.PMM, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Pmm, pmmVersion)
	if err != nil {
		return err
	}

	proxySQL, err := depFilter(deps.ProxySQL, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Proxysql, proxySQL)
	if err != nil {
		return err
	}

	logCollectorVersion, err := depFilter(deps.LogCollector, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.LogCollector, logCollectorVersion)
	if err != nil {
		return err
	}

	return nil
}

func psmdb(vs *pbVersion.VersionResponse, deps Deps, req *pbVersion.ApplyRequest) error {
	err := psmdbFilter(vs.Versions[0].Matrix.Mongod, req.Apply, req.DatabaseVersion)
	if err != nil {
		return err
	}

	productVersion := ""
	for k := range vs.Versions[0].Matrix.Mongod {
		productVersion = k
		break
	}

	backupVersion, err := depFilter(deps.Backup, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Backup, backupVersion)
	if err != nil {
		return err
	}

	pmmVersion, err := depFilter(deps.PMM, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Pmm, pmmVersion)
	if err != nil {
		return err
	}

	return nil
}

func validate(req *pbVersion.ApplyRequest) error {
	sep := "-"
	if strings.HasSuffix(req.Apply, sep+recommended) || strings.HasSuffix(req.Apply, sep+latest) {
		sp := strings.Split(req.Apply, sep)
		if len(sp) != 2 {
			return status.Errorf(codes.InvalidArgument, "invalid apply option: %s", req.Apply)
		}

		req.DatabaseVersion = sp[0]
		req.Apply = sp[1]
	}

	return nil
}
