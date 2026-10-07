package main

import (
	"io/fs"
	"testing"

	"github.com/Percona-Lab/percona-version-service/server"
	"github.com/stretchr/testify/require"
)

func TestBackend_create(t *testing.T) {
	metadataSub, err := fs.Sub(metaSources, "sources/metadata")
	require.NoError(t, err)

	releaseNotesSub, err := fs.Sub(releaseNoteSources, "sources/release-notes")
	require.NoError(t, err)

	_, err = server.New(metadataSub, releaseNotesSub)
	require.NoError(t, err)
}
