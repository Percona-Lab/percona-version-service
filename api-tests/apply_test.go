package api_tests

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Percona-Lab/percona-version-service/client"
	"github.com/Percona-Lab/percona-version-service/client/models"
	"github.com/Percona-Lab/percona-version-service/client/version_service"
	"github.com/stretchr/testify/assert"
)

func Test_apply_should_return_just_one_version(t *testing.T) {
	cli := cli()

	pxcParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.5.0",
		Product:         "pxc-operator",
	}
	pxcParams.WithTimeout(2 * time.Second)

	pxcResp, err := cli.VersionService.VersionServiceApply(pxcParams)
	assert.NoError(t, err)

	if len(pxcResp.Payload.Versions) != 1 ||
		len(pxcResp.Payload.Versions[0].Matrix.Pxc) != 1 ||
		len(pxcResp.Payload.Versions[0].Matrix.Backup) != 1 ||
		len(pxcResp.Payload.Versions[0].Matrix.Proxysql) != 1 ||
		len(pxcResp.Payload.Versions[0].Matrix.Pmm) != 1 ||
		len(pxcResp.Payload.Versions[0].Matrix.Haproxy) != 1 ||
		len(pxcResp.Payload.Versions[0].Matrix.Operator) != 1 {
		t.Error("more than one version returned")
	}

	psmdbParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.5.0",
		Product:         "psmdb-operator",
	}
	psmdbParams.WithTimeout(2 * time.Second)

	psmdbResp, err := cli.VersionService.VersionServiceApply(psmdbParams)
	assert.NoError(t, err)

	if len(pxcResp.Payload.Versions) != 1 ||
		len(psmdbResp.Payload.Versions[0].Matrix.Mongod) != 1 ||
		len(psmdbResp.Payload.Versions[0].Matrix.Backup) != 1 ||
		len(psmdbResp.Payload.Versions[0].Matrix.Pmm) != 1 ||
		len(psmdbResp.Payload.Versions[0].Matrix.Operator) != 1 {
		t.Error("more than one version returned")
	}
}

func Test_apply_pxc_should_return_same_major_version(t *testing.T) {
	cli := cli()

	pxcParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.5.0",
		Product:         "pxc-operator",
	}
	pxcParams.WithTimeout(2 * time.Second)

	for _, v := range []string{"5.7", "8.0"} {
		pxcParams.DatabaseVersion = &v
		pxcResp, err := cli.VersionService.VersionServiceApply(pxcParams)
		assert.NoError(t, err)

		k := getVersion(pxcResp.Payload.Versions[0].Matrix.Pxc)
		if !strings.HasPrefix(k, v) {
			t.Errorf("wrong version returned: %s", k)
		}
	}
}

func Test_apply_psmdb_should_return_same_major_version(t *testing.T) {
	cli := cli()

	psmdbParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.5.0",
		Product:         "psmdb-operator",
	}
	psmdbParams.WithTimeout(2 * time.Second)

	for _, v := range []string{"4.2", "4.0", "3.6"} {
		psmdbParams.DatabaseVersion = &v
		psmdbResp, err := cli.VersionService.VersionServiceApply(psmdbParams)
		assert.NoError(t, err)

		k := getVersion(psmdbResp.Payload.Versions[0].Matrix.Mongod)
		if !strings.HasPrefix(k, v) {
			t.Errorf("wrong version returned: %s", k)
		}
	}
}

func Test_apply_pxc_returned_versions(t *testing.T) {
	cli := cli()

	v57 := "5.7"

	cases := []struct {
		apply     string
		operator  string
		dbVersion *string
		version   string
	}{
		// test latest
		{"latest", "1.5.0", nil, "8.0.19-10.1"},
		{"latest", "1.4.0", nil, "8.0.18-9.3"},
		{"latest", "1.5.0", &v57, "5.7.30-31.43"},
		{"latest", "1.4.0", &v57, "5.7.28-31.41.2"},

		// test recommended
		{"recommended", "1.5.0", nil, "8.0.19-10.1"},
		{"recommended", "1.4.0", nil, "8.0.18-9.3"},
		{"recommended", "1.5.0", &v57, "5.7.30-31.43"},
		{"recommended", "1.4.0", &v57, "5.7.28-31.41.2"},

		// test exact
		{"5.7.28-31.41.2", "1.5.0", nil, "5.7.28-31.41.2"},
		{"5.7.28-31.41.2", "1.4.0", nil, "5.7.28-31.41.2"},
		{"8.0.19-10.1", "1.5.0", nil, "8.0.19-10.1"},
		{"8.0.18-9.3", "1.4.0", nil, "8.0.18-9.3"},
	}

	for _, c := range cases {
		params := &version_service.VersionServiceApplyParams{
			Apply:           c.apply,
			OperatorVersion: c.operator,
			Product:         "pxc-operator",
		}
		params.WithTimeout(2 * time.Second)
		if c.dbVersion != nil {
			params.DatabaseVersion = c.dbVersion
		}

		resp, err := cli.VersionService.VersionServiceApply(params)
		assert.NoError(t, err)

		v := getVersion(resp.Payload.Versions[0].Matrix.Pxc)
		assert.Equal(t, c.version, v)
	}
}

func Test_apply_psmdb_returned_versions(t *testing.T) {
	cli := cli()

	v36 := "3.6"
	v40 := "4.0"

	cases := []struct {
		apply     string
		operator  string
		dbVersion *string
		version   string
	}{
		// test latest
		{"latest", "1.5.0", nil, "4.2.8-8"},
		{"latest", "1.5.0", &v40, "4.0.19-12"},
		{"latest", "1.5.0", &v36, "3.6.18-6.0"},

		// test recommended
		{"recommended", "1.5.0", nil, "4.2.8-8"},
		{"recommended", "1.5.0", &v40, "4.0.19-12"},
		{"recommended", "1.5.0", &v36, "3.6.18-6.0"},

		// test exact
		{"4.2.7-7", "1.5.0", nil, "4.2.7-7"},
		{"4.0.18-11", "1.5.0", nil, "4.0.18-11"},
		{"3.6.18-5.0", "1.5.0", nil, "3.6.18-5.0"},
	}

	for _, c := range cases {
		params := &version_service.VersionServiceApplyParams{
			Apply:           c.apply,
			OperatorVersion: c.operator,
			Product:         "psmdb-operator",
		}
		params.WithTimeout(2 * time.Second)
		if c.dbVersion != nil {
			params.DatabaseVersion = c.dbVersion
		}

		resp, err := cli.VersionService.VersionServiceApply(params)
		assert.NoError(t, err)

		v := getVersion(resp.Payload.Versions[0].Matrix.Mongod)
		assert.Equal(t, c.version, v)
	}
}

func getVersion(v map[string]models.VersionVersion) string {
	for k := range v {
		return k
	}
	return ""
}

func cli() *client.APIVersionProto {
	host := "0.0.0.0:11000"
	if h, ok := os.LookupEnv("VS_HOST"); ok {
		host = h + ":11000"
	}

	cli := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:    host,
		Schemes: []string{"http"},
	})

	return cli
}
