package server

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb/api"
)

//go:embed metadata_test
var testFs embed.FS

func TestMetadata_Product(t *testing.T) {
	t.Parallel()

	fsSub, err := fs.Sub(testFs, "metadata_test")
	require.NoError(t, err)

	tests := []struct {
		name    string
		product string
		wantErr bool
		want    *pbVersion.MetadataResponse
	}{
		{
			name:    "returns correct everest information",
			product: "everest",
			want: &pbVersion.MetadataResponse{
				Versions: []*pbVersion.MetadataVersion{
					{
						Version: "0.6.0",
						Recommended: map[string]string{
							"cli": "0.4.0",
							"k8s": "1.27",
						},
						Supported: map[string]string{
							"pg":  "2.1.0",
							"pxc": "0.0.0",
						},
					},
					{
						Version: "0.7.0",
						Recommended: map[string]string{
							"cli": "0.5.0",
							"k8s": "1.28",
						},
						Supported: map[string]string{
							"pg":  "2.2.0 || 2.3.0",
							"pxc": "^1.2.3",
						},
					},
				},
			},
		},
		{
			name:    "returns correct kilimanjaro information",
			product: "kilimanjaro",
			want: &pbVersion.MetadataResponse{
				Versions: []*pbVersion.MetadataVersion{
					{
						Version: "1.2.0",
						Recommended: map[string]string{
							"cli": "1.2.0",
						},
					},
					{
						Version: "1.3.0",
						Recommended: map[string]string{
							"cli": "1.3.0",
						},
					},
				},
			},
		},
		{
			name:    "returns empty for non-existent product",
			product: "does not exist",
			want: &pbVersion.MetadataResponse{
				Versions: []*pbVersion.MetadataVersion{},
			},
		}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m, err := NewMetadata(fsSub)
			require.NoError(t, err)

			got, err := m.Product(tt.product)
			if (err != nil) != tt.wantErr {
				t.Errorf("Metadata.Product() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(got, tt.want, protocmp.Transform()); diff != "" {
				t.Errorf("Metadata.Product() diff %s", diff)
			}
		})
	}
}
