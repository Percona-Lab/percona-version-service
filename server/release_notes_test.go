package server

import (
	"embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/fs"
	"testing"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb/api"
)

//go:embed release_notes_test
var testReleaseNotesFS embed.FS

func TestGetReleaseNote(t *testing.T) {
	t.Parallel()
	sub, err := fs.Sub(testReleaseNotesFS, "release_notes_test")
	require.NoError(t, err)

	tests := []struct {
		name             string
		product          string
		version          string
		expectErr        bool
		expectedResponse *pbVersion.GetReleaseNotesResponse
	}{
		{
			name:      "returns correct release notes for pmm-server 2.42.0",
			product:   "pmm",
			version:   "2.42.0",
			expectErr: false,
			expectedResponse: &pbVersion.GetReleaseNotesResponse{
				Version:     "2.42.0",
				Product:     "pmm",
				ReleaseNote: "### PMM 2.42.0\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewReleaseNotes(sub)

			got, err := r.GetReleaseNote(tt.product, tt.version)
			if tt.expectErr {
				assert.NotNil(t, err)
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResponse.ReleaseNote, got.ReleaseNote)
		})
	}
}
