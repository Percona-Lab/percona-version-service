package server

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"

	pbExample "github.com/Percona-Lab/percona-version-service/proto"
)

// Backend implements the protobuf interface
type Backend struct {
	mu    *sync.RWMutex
	users []*pbExample.User
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{
		mu: &sync.RWMutex{},
	}
}

// AddUser adds a user to the in-memory store.
func (b *Backend) AddUser(ctx context.Context, _ *pbExample.EmptyRequest) (*pbExample.User, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	user := &pbExample.User{
		Id: uuid.Must(uuid.NewV4()).String(),
	}
	b.users = append(b.users, user)

	return user, nil
}

// ListUsers lists all users in the store.
func (b *Backend) ListUsers(_ *pbExample.ListUsersRequest, srv pbExample.UserService_ListUsersServer) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, user := range b.users {
		err := srv.Send(user)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) Apply(context.Context, *pbExample.EmptyRequest) (*pbExample.VersionResponse, error) {
	return &pbExample.VersionResponse{
		Versions: []*pbExample.OperatorVersion{
			{
				Operator: "1.5.0",
				Database: "pxc",
				Matrix: map[string]*pbExample.OperatorVersion_VersionMap{
					"pxc": {
						Value: map[string]*pbExample.Version{
							"8.0.18-9.3": {
								Imagepath: "http://hub.docker.com",
								Imagehash: "dhjsgflshjdgkljsdlkfj",
								Status:    "avaliable",
								Critilal:  false,
							},
						},
					},
				},
			},
		},
	}, nil
}
