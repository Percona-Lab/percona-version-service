package server

import (
	"context"

	pbExample "github.com/Percona-Lab/percona-version-service/proto"
)

// Backend implements the protobuf interface
type Backend struct {
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{}
}

func (b *Backend) Apply(ctx context.Context, req *pbExample.EmptyRequest) (*pbExample.VersionResponse, error) {
	return &pbExample.VersionResponse{
		Versions: []*pbExample.OperatorVersion{
			{
				Operator: "1.5.0",
				Database: "pxc",
				Matrix: &pbExample.VersionMatrix{
					Pxc: map[string]*pbExample.Version{
						req.Apply: {
							Version:   req.Apply,
							Imagepath: "percona/percona-xtradb-cluster-operator:1.4.0-pxc8.0",
							Imagehash: "some-hash",
							Status:    "recommended",
							Critilal:  false,
						},
					},
					Proxysql: map[string]*pbExample.Version{
						"master": {
							Version:   "master",
							Imagepath: "perconalab/percona-xtradb-cluster-operator:master-proxysql",
							Imagehash: "some-hash",
							Status:    "recommended",
							Critilal:  false,
						},
					},
					Backup: map[string]*pbExample.Version{
						"master": {
							Version:   "master",
							Imagepath: "perconalab/percona-xtradb-cluster-operator:master-pxc8.0",
							Imagehash: "some-hash",
							Status:    "recommended",
							Critilal:  false,
						},
					},
					Pmm: map[string]*pbExample.Version{
						"master": {
							Version:   "master",
							Imagepath: "perconalab/percona-xtradb-cluster-operator:master-pmm",
							Imagehash: "some-hash",
							Status:    "recommended",
							Critilal:  false,
						},
					},
				},
			},
		},
	}, nil
}
