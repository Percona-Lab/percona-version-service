package api_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Percona-Lab/percona-version-service/client/version_service"
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

		_, err := cli.VersionService.VersionServiceMetadata(params)
		assert.NoError(t, err)
	}
}
