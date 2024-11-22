package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb/api"
)

const pmmServerProduct = "pmm-server"

var ErrNotFound = errors.New("requested resource was not found")

type jsonpbObjectMarshaler struct {
	pb proto.Message
}

func (j *jsonpbObjectMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	// ZAP jsonEncoder deals with AddReflect by using json.MarshalObject. The same thing applies for consoleEncoder.
	return e.AddReflected("msg", j)
}

func (j *jsonpbObjectMarshaler) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := JsonPbMarshaller.Marshal(b, j.pb); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b.Bytes(), nil
}

var (
	// JsonPbMarshaller is the marshaller used for serializing protobuf messages.
	// If needed, this variable can be reassigned with a different marshaller with the same Marshal() signature.
	JsonPbMarshaller grpc_logging.JsonPbMarshaler = &jsonpb.Marshaler{}
)

// Backend implements the protobuf interface.
type Backend struct {
	metadata     *Metadata
	releaseNotes *ReleaseNotes
	pbVersion.UnimplementedVersionServiceServer
}

// New initializes a new Backend struct.
func New(metadata fs.FS, releaseNotes fs.FS) (*Backend, error) {
	m, err := NewMetadata(metadata)
	if err != nil {
		return nil, err
	}

	rn := NewReleaseNotes(releaseNotes)
	return &Backend{metadata: m, releaseNotes: rn}, nil
}

func (b *Backend) Product(ctx context.Context, req *pbVersion.ProductRequest) (*pbVersion.ProductResponse, error) {
	return operatorData(req.Product)
}

func (b *Backend) Operator(ctx context.Context, req *pbVersion.OperatorRequest) (*pbVersion.OperatorResponse, error) {
	productFamily := "operator"
	if req.Product == pmmServerProduct {
		productFamily = "pmm"
	}
	vs, err := operatorProductData(productFamily, req.Product, req.OperatorVersion)
	if err != nil {
		return nil, err
	}

	return &pbVersion.OperatorResponse{
		Versions: vs.Versions,
	}, nil
}

func (b *Backend) Apply(ctx context.Context, req *pbVersion.ApplyRequest) (*pbVersion.VersionResponse, error) {
	logger := ctxzap.Extract(ctx)

	logger.Info(
		"server request payload logged as grpc.request.content field",
		zap.String("grcp.start_time", time.Now().Format("2006-01-02T15:04:05Z07:00")),
		zap.String("system", "grpc"),
		zap.String("span.kind", "server"),
		zap.String("grpc.service", "version.VersionService"),
		zap.String("grpc.method", "Apply"),
		zap.Object("grpc.request.content", &jsonpbObjectMarshaler{req}),
	)

	if req.Product == pmmServerProduct {
		return nil, status.Error(codes.Unimplemented, "not implemented for pmm-server")
	}

	if req.Apply == disabled || req.Apply == never {
		return &pbVersion.VersionResponse{}, nil
	}

	err := transformRequest(req)
	if err != nil {
		return nil, err
	}

	vs, err := operatorProductData("operator", req.Product, req.OperatorVersion)
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
	case "pg-operator":
		err := pg(vs, deps, req)
		if err != nil {
			return nil, err
		}
	case "ps-operator":
		err := ps(vs, deps, req)
		if err != nil {
			return nil, err
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid product: %s", req.Product)
	}

	return vs, nil
}

func (b *Backend) Metadata(ctx context.Context, req *pbVersion.MetadataRequest) (*pbVersion.MetadataResponse, error) {
	return b.metadata.Product(req.Product)
}

func (b *Backend) MetadataV2(ctx context.Context, req *pbVersion.MetadataRequest) (*pbVersion.MetadataV2Response, error) {
	return b.metadata.ProductV2(req.Product)
}

func (b *Backend) GetReleaseNotes(ctx context.Context, req *pbVersion.GetReleaseNotesRequest) (*pbVersion.GetReleaseNotesResponse, error) {
	return b.releaseNotes.GetReleaseNote(req.Product, req.Version)
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
	err = defaultFilter(vs.Versions[0].Matrix.Backup, backupVersion, true)
	if err != nil {
		return err
	}

	pmmVersion, err := depFilter(deps.PMM, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Pmm, pmmVersion, true)
	if err != nil {
		return err
	}

	proxySQL, err := depFilter(deps.ProxySQL, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Proxysql, proxySQL, false)
	if err != nil {
		return err
	}

	haproxy, err := depFilter(deps.Haproxy, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Haproxy, haproxy, true)
	if err != nil {
		return err
	}

	logCollectorVersion, err := depFilter(deps.LogCollector, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.LogCollector, logCollectorVersion, false)
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
	err = defaultFilter(vs.Versions[0].Matrix.Backup, backupVersion, true)
	if err != nil {
		return err
	}

	pmmVersion, err := depFilter(deps.PMM, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Pmm, pmmVersion, true)
	if err != nil {
		return err
	}

	return nil
}

func pg(vs *pbVersion.VersionResponse, deps Deps, req *pbVersion.ApplyRequest) error {
	err := pgFilter(vs.Versions[0].Matrix.Postgresql, req.Apply, req.DatabaseVersion)
	if err != nil {
		return err
	}

	productVersion := ""
	for k := range vs.Versions[0].Matrix.Postgresql {
		productVersion = k
		break
	}

	depVer, err := pgDepFilter(deps.PgBackrest, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Pgbackrest, depVer, true)
	if err != nil {
		return err
	}

	depVer, err = pgDepFilter(deps.PgBackrestRepo, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.PgbackrestRepo, depVer, true)
	if err != nil {
		return err
	}

	depVer, err = pgDepFilter(deps.Pgbadger, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Pgbadger, depVer, true)
	if err != nil {
		return err
	}

	depVer, err = pgDepFilter(deps.Pgbouncer, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Pgbouncer, depVer, true)
	if err != nil {
		return err
	}

	depVer, err = pgDepFilter(deps.Postgis, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Postgis, depVer, true)
	if err != nil {
		return err
	}

	return nil
}

func ps(vs *pbVersion.VersionResponse, deps Deps, req *pbVersion.ApplyRequest) error {
	err := psFilter(vs.Versions[0].Matrix.Mysql, req.Apply, req.DatabaseVersion)
	if err != nil {
		return err
	}

	productVersion := ""
	for k := range vs.Versions[0].Matrix.Mysql {
		productVersion = k
		break
	}

	backupVersion, err := depFilter(deps.Backup, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Backup, backupVersion, true)
	if err != nil {
		return err
	}

	pmmVersion, err := depFilter(deps.PMM, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Pmm, pmmVersion, true)
	if err != nil {
		return err
	}
	orchestratorVersion, err := depFilter(deps.Orchestrator, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Orchestrator, orchestratorVersion, true)
	if err != nil {
		return err
	}
	routerVersion, err := depFilter(deps.Router, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Router, routerVersion, true)
	if err != nil {
		return err
	}
	haproxyVersion, err := depFilter(deps.Haproxy, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Haproxy, haproxyVersion, true)
	if err != nil {
		return err
	}
	toolkitVersion, err := depFilter(deps.Toolkit, productVersion)
	if err != nil {
		return err
	}
	err = defaultFilter(vs.Versions[0].Matrix.Toolkit, toolkitVersion, true)
	if err != nil {
		return err
	}

	return nil
}

func transformRequest(req *pbVersion.ApplyRequest) error {
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
