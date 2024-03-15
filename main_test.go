package main

import (
	"io/fs"
	"testing"

	"github.com/Percona-Lab/percona-version-service/server"
	"github.com/stretchr/testify/require"
)

func TestBackend_create(t *testing.T) {
	sub, err := fs.Sub(metaSources, "sources/metadata")
	require.NoError(t, err)

	_, err = server.New(sub)
	require.NoError(t, err)
}
