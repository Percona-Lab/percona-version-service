package api_tests

import (
	"testing"
	"time"

	"github.com/Percona-Lab/percona-version-service/client/version_service"
	"github.com/stretchr/testify/assert"
)

func TestProductRouteShouldReturnRigthProductName(t *testing.T) {
	cli := cli()

	cases := []struct {
		product string
	}{
		{"pxc-operator"},
		{"psmdb-operator"},
		{"postgresql-operator"},
	}

	for _, c := range cases {
		params := &version_service.VersionServiceProductParams{
			Product: c.product,
		}
		params.WithTimeout(2 * time.Second)

		resp, err := cli.VersionService.VersionServiceProduct(params)
		assert.NoError(t, err)

		assert.NotEmpty(t, resp.Payload.Versions)

		for _, v := range resp.Payload.Versions {
			assert.Equal(t, c.product, v.Product)
		}
	}
}
