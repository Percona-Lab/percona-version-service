package api_tests

import (
	"testing"
	"time"

	"github.com/Percona-Lab/percona-version-service/client/version_service"
	version "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadatatRouteShouldReturnSuccessfulResponse(t *testing.T) {
	cli := cli()

	cases := []struct {
		product string
	}{
		{"non-existent-product"},
	}

	for _, c := range cases {
		params := &version_service.VersionServiceMetadataParams{
			Product: c.product,
		}
		params.WithTimeout(2 * time.Second)

		meta, err := cli.VersionService.VersionServiceMetadata(params)
		require.NoError(t, err)
		for _, m := range meta.Payload.Versions {
			_, err := version.NewSemver(m.Version)
			assert.NoError(t, err)

			// All recommended strings can be parsed as a version
			for _, v := range m.Recommended {
				_, err := version.NewSemver(v)
				assert.NoError(t, err)
			}

			// All supported strings can be parsed as constraints
			for _, s := range m.Supported {
				_, err := version.NewConstraint(s)
				assert.NoError(t, err)
			}
		}
	}
}
