package api_tests

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Percona-Lab/percona-version-service/client"
	"github.com/Percona-Lab/percona-version-service/client/models"
	"github.com/Percona-Lab/percona-version-service/client/version_service"
)

func TestApplyShouldReturnJustOneVersion(t *testing.T) {
	cli := cli()

	pxcParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.12.0",
		Product:         "pxc-operator",
	}
	pxcParams.WithTimeout(2 * time.Second)

	pxcResp, err := cli.VersionService.VersionServiceApply(pxcParams)
	assert.NoError(t, err)

	assert.Len(t, pxcResp.Payload.Versions, 1)
	assert.Len(t, pxcResp.Payload.Versions[0].Matrix.Pxc, 1)
	assert.Len(t, pxcResp.Payload.Versions[0].Matrix.Backup, 1)
	assert.Len(t, pxcResp.Payload.Versions[0].Matrix.Proxysql, 1)
	assert.Len(t, pxcResp.Payload.Versions[0].Matrix.Pmm, 1)
	assert.Len(t, pxcResp.Payload.Versions[0].Matrix.Haproxy, 1)
	assert.Len(t, pxcResp.Payload.Versions[0].Matrix.Operator, 1)

	psmdbParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.13.0",
		Product:         "psmdb-operator",
	}
	psmdbParams.WithTimeout(2 * time.Second)

	psmdbResp, err := cli.VersionService.VersionServiceApply(psmdbParams)
	assert.NoError(t, err)

	assert.Len(t, psmdbResp.Payload.Versions, 1)
	assert.Len(t, psmdbResp.Payload.Versions[0].Matrix.Mongod, 1)
	assert.Len(t, psmdbResp.Payload.Versions[0].Matrix.Backup, 1)
	assert.Len(t, psmdbResp.Payload.Versions[0].Matrix.Pmm, 1)
	assert.Len(t, psmdbResp.Payload.Versions[0].Matrix.Operator, 1)

	pgParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.1.0",
		Product:         "pg-operator",
	}
	pgParams.WithTimeout(2 * time.Second)

	pgResp, err := cli.VersionService.VersionServiceApply(pgParams)
	assert.NoError(t, err)

	assert.Len(t, pgResp.Payload.Versions, 1)
	assert.Len(t, pgResp.Payload.Versions[0].Matrix.Postgresql, 1)
	assert.Len(t, pgResp.Payload.Versions[0].Matrix.Pmm, 1)
	assert.Len(t, pgResp.Payload.Versions[0].Matrix.Pgbackrest, 1)
	assert.Len(t, pgResp.Payload.Versions[0].Matrix.PgbackrestRepo, 1)
	assert.Len(t, pgResp.Payload.Versions[0].Matrix.Pgbadger, 1)
	assert.Len(t, pgResp.Payload.Versions[0].Matrix.Pgbouncer, 1)
	assert.Len(t, pgResp.Payload.Versions[0].Matrix.Operator, 1)

	psParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "0.2.0",
		Product:         "ps-operator",
	}
	psParams.WithTimeout(2 * time.Second)

	psResp, err := cli.VersionService.VersionServiceApply(psParams)
	assert.NoError(t, err)

	assert.Len(t, psResp.Payload.Versions, 1)
	assert.Len(t, psResp.Payload.Versions[0].Matrix.Mysql, 1)
	assert.Len(t, psResp.Payload.Versions[0].Matrix.Pmm, 1)
	assert.Len(t, psResp.Payload.Versions[0].Matrix.Backup, 1)
	assert.Len(t, psResp.Payload.Versions[0].Matrix.Orchestrator, 1)
	assert.Len(t, psResp.Payload.Versions[0].Matrix.Router, 1)
	assert.Len(t, psResp.Payload.Versions[0].Matrix.Operator, 1)
}

func TestApplyPxcShouldReturnSameMajorVersion(t *testing.T) {
	cli := cli()

	pxcParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.12.0",
		Product:         "pxc-operator",
	}
	pxcParams.WithTimeout(2 * time.Second)

	for _, v := range []string{"5.7", "8.0"} {
		pxcParams.DatabaseVersion = &v
		pxcResp, err := cli.VersionService.VersionServiceApply(pxcParams)
		assert.NoError(t, err)

		k := getVersion(pxcResp.Payload.Versions[0].Matrix.Pxc)
		assert.True(t, strings.HasPrefix(k, v))
	}
}

func TestApplyPsmdbShouldReturnSameMajorVersion(t *testing.T) {
	cli := cli()

	psmdbParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.12.0",
		Product:         "psmdb-operator",
	}
	psmdbParams.WithTimeout(2 * time.Second)

	for _, v := range []string{"5.0", "4.4", "4.2"} {
		psmdbParams.DatabaseVersion = &v
		psmdbResp, err := cli.VersionService.VersionServiceApply(psmdbParams)
		assert.NoError(t, err)

		k := getVersion(psmdbResp.Payload.Versions[0].Matrix.Mongod)
		assert.True(t, strings.HasPrefix(k, v))
	}
}

func TestApplyPgShouldReturnSameMajorVersion(t *testing.T) {
	cli := cli()

	pgParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.3.0",
		Product:         "pg-operator",
	}
	pgParams.WithTimeout(2 * time.Second)

	for _, v := range []string{"12.11", "13.7", "14.4"} {
		pgParams.DatabaseVersion = &v
		psmdbResp, err := cli.VersionService.VersionServiceApply(pgParams)
		assert.NoError(t, err)

		k := getVersion(psmdbResp.Payload.Versions[0].Matrix.Postgresql)
		assert.True(t, strings.HasPrefix(k, strings.Split(v, ".")[0]))
	}
}

func TestApplyPsShouldReturnSameMajorVersion(t *testing.T) {
	cli := cli()

	params := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "0.2.0",
		Product:         "ps-operator",
	}
	params.WithTimeout(2 * time.Second)

	for _, v := range []string{"8.0.27"} {
		params.DatabaseVersion = &v
		resp, err := cli.VersionService.VersionServiceApply(params)
		assert.NoError(t, err)

		k := getVersion(resp.Payload.Versions[0].Matrix.Mysql)
		assert.True(t, strings.HasPrefix(k, strings.Split(v, ".")[0]))
	}
}

func TestApplyPxcReturnedVersions(t *testing.T) {
	cli := cli()

	v57 := "5.7"
	vPreRel := "5.7.31-99-99"

	cases := []struct {
		apply     string
		operator  string
		dbVersion *string
		version   string
	}{
		// test latest
		{"latest", "1.12.0", nil, "8.0.29-21.1"},
		{"latest", "1.11.0", nil, "8.0.27-18.1"},
		{"latest", "1.10.0", nil, "8.0.25-15.1"},
		{"latest", "1.9.0", nil, "8.0.23-14.1"},
		{"latest", "1.8.0", nil, "8.0.22-13.1"},
		{"latest", "1.7.0", nil, "8.0.21-12.1"},
		{"latest", "1.6.0", nil, "8.0.20-11.2"},
		{"latest", "1.5.0", nil, "8.0.20-11.2"},
		{"latest", "1.4.0", nil, "8.0.18-9.3"},
		{"latest", "1.12.0", &v57, "5.7.39-31.61"},
		{"latest", "1.11.0", &v57, "5.7.36-31.55"},
		{"latest", "1.10.0", &v57, "5.7.35-31.53"},
		{"latest", "1.9.0", &v57, "5.7.34-31.51"},
		{"latest", "1.8.0", &v57, "5.7.33-31.49"},
		{"latest", "1.7.0", &v57, "5.7.32-31.47"},
		{"latest", "1.6.0", &v57, "5.7.31-31.45.2"},
		{"latest", "1.5.0", &v57, "5.7.31-31.45.2"},
		{"latest", "1.4.0", &v57, "5.7.28-31.41.2"},

		// test latest when prerelease part in current version is bigger than in latest
		{"latest", "1.12.0", &vPreRel, "5.7.39-31.61"},
		{"latest", "1.11.0", &vPreRel, "5.7.36-31.55"},
		{"latest", "1.10.0", &vPreRel, "5.7.35-31.53"},
		{"latest", "1.9.0", &vPreRel, "5.7.34-31.51"},
		{"latest", "1.8.0", &vPreRel, "5.7.33-31.49"},
		{"latest", "1.7.0", &vPreRel, "5.7.32-31.47"},
		{"latest", "1.6.0", &vPreRel, "5.7.31-31.45.2"},
		{"latest", "1.5.0", &vPreRel, "5.7.31-31.45.2"},

		// test recommended
		{"recommended", "1.12.0", nil, "8.0.29-21.1"},
		{"recommended", "1.11.0", nil, "8.0.27-18.1"},
		{"recommended", "1.10.0", nil, "8.0.25-15.1"},
		{"recommended", "1.9.0", nil, "8.0.23-14.1"},
		{"recommended", "1.8.0", nil, "8.0.22-13.1"},
		{"recommended", "1.7.0", nil, "8.0.21-12.1"},
		{"recommended", "1.6.0", nil, "8.0.20-11.2"},
		{"recommended", "1.5.0", nil, "8.0.20-11.2"},
		{"recommended", "1.4.0", nil, "8.0.18-9.3"},
		{"recommended", "1.12.0", &v57, "5.7.39-31.61"},
		{"recommended", "1.11.0", &v57, "5.7.36-31.55"},
		{"recommended", "1.10.0", &v57, "5.7.35-31.53"},
		{"recommended", "1.9.0", &v57, "5.7.34-31.51"},
		{"recommended", "1.8.0", &v57, "5.7.33-31.49"},
		{"recommended", "1.7.0", &v57, "5.7.32-31.47"},
		{"recommended", "1.6.0", &v57, "5.7.31-31.45.2"},
		{"recommended", "1.5.0", &v57, "5.7.31-31.45.2"},
		{"recommended", "1.4.0", &v57, "5.7.28-31.41.2"},

		// test exact
		{"5.7.36-31.55", "1.12.0", nil, "5.7.36-31.55"},
		{"5.7.28-31.41.2", "1.11.0", nil, "5.7.28-31.41.2"},
		{"5.7.28-31.41.2", "1.10.0", nil, "5.7.28-31.41.2"},
		{"5.7.28-31.41.2", "1.9.0", nil, "5.7.28-31.41.2"},
		{"5.7.28-31.41.2", "1.8.0", nil, "5.7.28-31.41.2"},
		{"5.7.28-31.41.2", "1.7.0", nil, "5.7.28-31.41.2"},
		{"5.7.28-31.41.2", "1.6.0", nil, "5.7.28-31.41.2"},
		{"5.7.28-31.41.2", "1.5.0", nil, "5.7.28-31.41.2"},
		{"5.7.28-31.41.2", "1.4.0", nil, "5.7.28-31.41.2"},
		{"8.0.27-18.1", "1.12.0", nil, "8.0.27-18.1"},
		{"8.0.19-10.1", "1.11.0", nil, "8.0.19-10.1"},
		{"8.0.19-10.1", "1.10.0", nil, "8.0.19-10.1"},
		{"8.0.19-10.1", "1.9.0", nil, "8.0.19-10.1"},
		{"8.0.19-10.1", "1.8.0", nil, "8.0.19-10.1"},
		{"8.0.19-10.1", "1.7.0", nil, "8.0.19-10.1"},
		{"8.0.19-10.1", "1.6.0", nil, "8.0.19-10.1"},
		{"8.0.19-10.1", "1.5.0", nil, "8.0.19-10.1"},
		{"8.0.18-9.3", "1.4.0", nil, "8.0.18-9.3"},

		//test with suffix
		{"8.0-latest", "1.12.0", nil, "8.0.29-21.1"},
		{"8.0-latest", "1.11.0", nil, "8.0.27-18.1"},
		{"8.0-latest", "1.10.0", nil, "8.0.25-15.1"},
		{"8.0-latest", "1.9.0", nil, "8.0.23-14.1"},
		{"8.0-latest", "1.8.0", nil, "8.0.22-13.1"},
		{"8.0-latest", "1.7.0", nil, "8.0.21-12.1"},
		{"8.0-latest", "1.6.0", nil, "8.0.20-11.2"},
		{"8.0-latest", "1.5.0", nil, "8.0.20-11.2"},
		{"8.0-latest", "1.4.0", nil, "8.0.18-9.3"},
		{"5.7-latest", "1.12.0", nil, "5.7.39-31.61"},
		{"5.7-latest", "1.11.0", nil, "5.7.36-31.55"},
		{"5.7-latest", "1.10.0", nil, "5.7.35-31.53"},
		{"5.7-latest", "1.9.0", nil, "5.7.34-31.51"},
		{"5.7-latest", "1.8.0", nil, "5.7.33-31.49"},
		{"5.7-latest", "1.7.0", nil, "5.7.32-31.47"},
		{"5.7-latest", "1.6.0", nil, "5.7.31-31.45.2"},
		{"5.7-latest", "1.5.0", nil, "5.7.31-31.45.2"},
		{"5.7-latest", "1.4.0", nil, "5.7.28-31.41.2"},
		{"8.0-recommended", "1.12.0", nil, "8.0.29-21.1"},
		{"8.0-recommended", "1.11.0", nil, "8.0.27-18.1"},
		{"8.0-recommended", "1.10.0", nil, "8.0.25-15.1"},
		{"8.0-recommended", "1.9.0", nil, "8.0.23-14.1"},
		{"8.0-recommended", "1.8.0", nil, "8.0.22-13.1"},
		{"8.0-recommended", "1.7.0", nil, "8.0.21-12.1"},
		{"8.0-recommended", "1.6.0", nil, "8.0.20-11.2"},
		{"8.0-recommended", "1.5.0", nil, "8.0.20-11.2"},
		{"8.0-recommended", "1.4.0", nil, "8.0.18-9.3"},
		{"5.7-recommended", "1.12.0", nil, "5.7.39-31.61"},
		{"5.7-recommended", "1.11.0", nil, "5.7.36-31.55"},
		{"5.7-recommended", "1.10.0", nil, "5.7.35-31.53"},
		{"5.7-recommended", "1.9.0", nil, "5.7.34-31.51"},
		{"5.7-recommended", "1.8.0", nil, "5.7.33-31.49"},
		{"5.7-recommended", "1.7.0", nil, "5.7.32-31.47"},
		{"5.7-recommended", "1.6.0", nil, "5.7.31-31.45.2"},
		{"5.7-recommended", "1.5.0", nil, "5.7.31-31.45.2"},
		{"5.7-recommended", "1.4.0", nil, "5.7.28-31.41.2"},
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

func TestApplyPsmdbReturnedVersions(t *testing.T) {
	cli := cli()

	v36 := "3.6"
	v40 := "4.0"
	v42 := "4.2"
	v44 := "4.4"

	cases := []struct {
		apply     string
		operator  string
		dbVersion *string
		version   string
	}{
		// test latest
		{"latest", "1.13.0", nil, "5.0.11-10"},
		{"latest", "1.12.0", nil, "5.0.7-6"},
		{"latest", "1.11.0", nil, "5.0.4-3"},
		{"latest", "1.10.0", nil, "5.0.2-1"},
		{"latest", "1.9.0", nil, "4.4.6-8"},
		{"latest", "1.8.0", nil, "4.4.5-7"},
		{"latest", "1.7.0", nil, "4.4.3-5"},
		{"latest", "1.6.0", nil, "4.4.2-4"},
		{"latest", "1.5.0", nil, "4.2.8-8"},
		{"latest", "1.13.0", &v44, "4.4.16-16"},
		{"latest", "1.12.0", &v44, "4.4.13-13"},
		{"latest", "1.11.0", &v44, "4.4.10-11"},
		{"latest", "1.10.0", &v44, "4.4.8-9"},
		{"latest", "1.9.0", &v44, "4.4.6-8"},
		{"latest", "1.8.0", &v44, "4.4.5-7"},
		{"latest", "1.7.0", &v44, "4.4.3-5"},
		{"latest", "1.6.0", &v44, "4.4.2-4"},
		{"latest", "1.13.0", &v42, "4.2.22-22"},
		{"latest", "1.12.0", &v42, "4.2.19-19"},
		{"latest", "1.11.0", &v42, "4.2.17-17"},
		{"latest", "1.10.0", &v42, "4.2.15-16"},
		{"latest", "1.9.0", &v42, "4.2.14-15"},
		{"latest", "1.8.0", &v42, "4.2.13-14"},
		{"latest", "1.7.0", &v42, "4.2.12-13"},
		{"latest", "1.6.0", &v42, "4.2.11-12"},
		{"latest", "1.5.0", &v42, "4.2.8-8"},
		{"latest", "1.11.0", &v40, "4.0.27-22"},
		{"latest", "1.10.0", &v40, "4.0.26-21"},
		{"latest", "1.9.0", &v40, "4.0.25-20"},
		{"latest", "1.8.0", &v40, "4.0.23-18"},
		{"latest", "1.7.0", &v40, "4.0.22-17"},
		{"latest", "1.6.0", &v40, "4.0.21-15"},
		{"latest", "1.5.0", &v40, "4.0.20-13"},
		{"latest", "1.8.0", &v36, "3.6.23-13.0"},
		{"latest", "1.7.0", &v36, "3.6.21-10.0"},
		{"latest", "1.6.0", &v36, "3.6.21-10.0"},
		{"latest", "1.5.0", &v36, "3.6.19-7.0"},

		// test recommended
		{"recommended", "1.13.0", nil, "5.0.11-10"},
		{"recommended", "1.12.0", nil, "5.0.7-6"},
		{"recommended", "1.11.0", nil, "4.4.10-11"},
		{"recommended", "1.10.0", nil, "4.4.8-9"},
		{"recommended", "1.9.0", nil, "4.4.6-8"},
		{"recommended", "1.8.0", nil, "4.4.5-7"},
		{"recommended", "1.7.0", nil, "4.4.3-5"},
		{"recommended", "1.6.0", nil, "4.4.2-4"},
		{"recommended", "1.5.0", nil, "4.2.8-8"},
		{"recommended", "1.13.0", &v44, "4.4.16-16"},
		{"recommended", "1.12.0", &v44, "4.4.13-13"},
		{"recommended", "1.11.0", &v44, "4.4.10-11"},
		{"recommended", "1.10.0", &v44, "4.4.8-9"},
		{"recommended", "1.9.0", &v44, "4.4.6-8"},
		{"recommended", "1.8.0", &v44, "4.4.5-7"},
		{"recommended", "1.7.0", &v44, "4.4.3-5"},
		{"recommended", "1.6.0", &v44, "4.4.2-4"},
		{"recommended", "1.13.0", &v42, "4.2.22-22"},
		{"recommended", "1.12.0", &v42, "4.2.19-19"},
		{"recommended", "1.11.0", &v42, "4.2.17-17"},
		{"recommended", "1.10.0", &v42, "4.2.15-16"},
		{"recommended", "1.9.0", &v42, "4.2.14-15"},
		{"recommended", "1.8.0", &v42, "4.2.13-14"},
		{"recommended", "1.7.0", &v42, "4.2.12-13"},
		{"recommended", "1.6.0", &v42, "4.2.11-12"},
		{"recommended", "1.5.0", &v42, "4.2.8-8"},
		{"recommended", "1.11.0", &v40, "4.0.27-22"},
		{"recommended", "1.10.0", &v40, "4.0.26-21"},
		{"recommended", "1.9.0", &v40, "4.0.25-20"},
		{"recommended", "1.8.0", &v40, "4.0.23-18"},
		{"recommended", "1.7.0", &v40, "4.0.22-17"},
		{"recommended", "1.6.0", &v40, "4.0.21-15"},
		{"recommended", "1.5.0", &v40, "4.0.20-13"},
		{"recommended", "1.8.0", &v36, "3.6.23-13.0"},
		{"recommended", "1.7.0", &v36, "3.6.21-10.0"},
		{"recommended", "1.6.0", &v36, "3.6.21-10.0"},
		{"recommended", "1.5.0", &v36, "3.6.19-7.0"},

		// test exact
		{"5.0.7-6", "1.13.0", nil, "5.0.7-6"},
		{"5.0.2-1", "1.12.0", nil, "5.0.2-1"},
		{"5.0.2-1", "1.11.0", nil, "5.0.2-1"},
		{"5.0.2-1", "1.10.0", nil, "5.0.2-1"},
		{"4.4.13-13", "1.13.0", nil, "4.4.13-13"},
		{"4.4.6-8", "1.12.0", nil, "4.4.6-8"},
		{"4.4.6-8", "1.11.0", nil, "4.4.6-8"},
		{"4.4.6-8", "1.10.0", nil, "4.4.6-8"},
		{"4.4.2-4", "1.9.0", nil, "4.4.2-4"},
		{"4.4.2-4", "1.8.0", nil, "4.4.2-4"},
		{"4.4.2-4", "1.7.0", nil, "4.4.2-4"},
		{"4.4.2-4", "1.6.0", nil, "4.4.2-4"},
		{"4.2.22-22", "1.13.0", nil, "4.2.22-22"},
		{"4.2.14-15", "1.12.0", nil, "4.2.14-15"},
		{"4.2.14-15", "1.11.0", nil, "4.2.14-15"},
		{"4.2.14-15", "1.10.0", nil, "4.2.14-15"},
		{"4.2.7-7", "1.9.0", nil, "4.2.7-7"},
		{"4.2.7-7", "1.8.0", nil, "4.2.7-7"},
		{"4.2.7-7", "1.7.0", nil, "4.2.7-7"},
		{"4.2.7-7", "1.6.0", nil, "4.2.7-7"},
		{"4.2.7-7", "1.5.0", nil, "4.2.7-7"},
		{"4.0.25-20", "1.11.0", nil, "4.0.25-20"},
		{"4.0.25-20", "1.10.0", nil, "4.0.25-20"},
		{"4.0.18-11", "1.9.0", nil, "4.0.18-11"},
		{"4.0.18-11", "1.8.0", nil, "4.0.18-11"},
		{"4.0.18-11", "1.7.0", nil, "4.0.18-11"},
		{"4.0.18-11", "1.6.0", nil, "4.0.18-11"},
		{"4.0.18-11", "1.5.0", nil, "4.0.18-11"},
		{"3.6.18-5.0", "1.8.0", nil, "3.6.18-5.0"},
		{"3.6.18-5.0", "1.7.0", nil, "3.6.18-5.0"},
		{"3.6.18-5.0", "1.6.0", nil, "3.6.18-5.0"},
		{"3.6.18-5.0", "1.5.0", nil, "3.6.18-5.0"},

		//test with suffix
		{"5.0-latest", "1.13.0", nil, "5.0.11-10"},
		{"5.0-latest", "1.12.0", nil, "5.0.7-6"},
		{"5.0-latest", "1.11.0", nil, "5.0.4-3"},
		{"5.0-latest", "1.10.0", nil, "5.0.2-1"},
		{"4.4-latest", "1.13.0", nil, "4.4.16-16"},
		{"4.4-latest", "1.12.0", nil, "4.4.13-13"},
		{"4.4-latest", "1.11.0", nil, "4.4.10-11"},
		{"4.4-latest", "1.10.0", nil, "4.4.8-9"},
		{"4.4-latest", "1.9.0", nil, "4.4.6-8"},
		{"4.4-latest", "1.8.0", nil, "4.4.5-7"},
		{"4.4-latest", "1.7.0", nil, "4.4.3-5"},
		{"4.4-latest", "1.6.0", nil, "4.4.2-4"},
		{"4.2-latest", "1.5.0", nil, "4.2.8-8"},
		{"4.2-latest", "1.13.0", nil, "4.2.22-22"},
		{"4.2-latest", "1.12.0", nil, "4.2.19-19"},
		{"4.2-latest", "1.11.0", nil, "4.2.17-17"},
		{"4.2-latest", "1.10.0", nil, "4.2.15-16"},
		{"4.2-latest", "1.9.0", nil, "4.2.14-15"},
		{"4.2-latest", "1.8.0", nil, "4.2.13-14"},
		{"4.2-latest", "1.7.0", nil, "4.2.12-13"},
		{"4.2-latest", "1.6.0", nil, "4.2.11-12"},
		{"4.0-latest", "1.11.0", nil, "4.0.27-22"},
		{"4.0-latest", "1.10.0", nil, "4.0.26-21"},
		{"4.0-latest", "1.9.0", nil, "4.0.25-20"},
		{"4.0-latest", "1.8.0", nil, "4.0.23-18"},
		{"4.0-latest", "1.7.0", nil, "4.0.22-17"},
		{"4.0-latest", "1.6.0", nil, "4.0.21-15"},
		{"4.0-latest", "1.5.0", nil, "4.0.20-13"},
		{"3.6-latest", "1.8.0", nil, "3.6.23-13.0"},
		{"3.6-latest", "1.7.0", nil, "3.6.21-10.0"},
		{"3.6-latest", "1.6.0", nil, "3.6.21-10.0"},
		{"3.6-latest", "1.5.0", nil, "3.6.19-7.0"},
		{"5.0-recommended", "1.13.0", nil, "5.0.11-10"},
		{"5.0-recommended", "1.12.0", nil, "5.0.7-6"},
		{"4.4-recommended", "1.13.0", nil, "4.4.16-16"},
		{"4.4-recommended", "1.12.0", nil, "4.4.13-13"},
		{"4.4-recommended", "1.11.0", nil, "4.4.10-11"},
		{"4.4-recommended", "1.10.0", nil, "4.4.8-9"},
		{"4.4-recommended", "1.9.0", nil, "4.4.6-8"},
		{"4.4-recommended", "1.8.0", nil, "4.4.5-7"},
		{"4.4-recommended", "1.7.0", nil, "4.4.3-5"},
		{"4.4-recommended", "1.6.0", nil, "4.4.2-4"},
		{"4.2-recommended", "1.13.0", nil, "4.2.22-22"},
		{"4.2-recommended", "1.12.0", nil, "4.2.19-19"},
		{"4.2-recommended", "1.11.0", nil, "4.2.17-17"},
		{"4.2-recommended", "1.10.0", nil, "4.2.15-16"},
		{"4.2-recommended", "1.9.0", nil, "4.2.14-15"},
		{"4.2-recommended", "1.8.0", nil, "4.2.13-14"},
		{"4.2-recommended", "1.7.0", nil, "4.2.12-13"},
		{"4.2-recommended", "1.6.0", nil, "4.2.11-12"},
		{"4.2-recommended", "1.5.0", nil, "4.2.8-8"},
		{"4.0-recommended", "1.11.0", nil, "4.0.27-22"},
		{"4.0-recommended", "1.10.0", nil, "4.0.26-21"},
		{"4.0-recommended", "1.9.0", nil, "4.0.25-20"},
		{"4.0-recommended", "1.8.0", nil, "4.0.23-18"},
		{"4.0-recommended", "1.7.0", nil, "4.0.22-17"},
		{"4.0-recommended", "1.6.0", nil, "4.0.21-15"},
		{"4.0-recommended", "1.5.0", nil, "4.0.20-13"},
		{"3.6-recommended", "1.8.0", nil, "3.6.23-13.0"},
		{"3.6-recommended", "1.7.0", nil, "3.6.21-10.0"},
		{"3.6-recommended", "1.6.0", nil, "3.6.21-10.0"},
		{"3.6-recommended", "1.5.0", nil, "3.6.19-7.0"},
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

func TestApplyPGReturnedVersions(t *testing.T) {
	cli := cli()

	v12 := "12.11"
	v13 := "13.7"
	v14 := "14.4"

	cases := []struct {
		apply     string
		operator  string
		dbVersion *string
		version   string
	}{
		// test latest
		{"latest", "1.3.0", nil, "14.4"},
		{"latest", "1.3.0", &v12, "12.11"},
		{"latest", "1.3.0", &v13, "13.7"},
		{"latest", "1.3.0", &v14, "14.4"},

		// test recommended
		{"recommended", "1.3.0", nil, "14.4"},
		{"recommended", "1.3.0", &v12, "12.11"},
		{"recommended", "1.3.0", &v13, "13.7"},
		{"recommended", "1.3.0", &v14, "14.4"},

		// test exact
		{"12.8", "1.1.0", nil, "12.8"},
		{"13.5", "1.1.0", nil, "13.5"},
		{"14.1", "1.1.0", nil, "14.1"},
		{"12.11", "1.3.0", nil, "12.11"},
		{"13.7", "1.3.0", nil, "13.7"},
		{"14.4", "1.3.0", nil, "14.4"},

		//test with suffix
		{"12-latest", "1.1.0", nil, "12.8"},
		{"13-latest", "1.1.0", nil, "13.5"},
		{"14-latest", "1.1.0", nil, "14.1"},
		{"12-latest", "1.3.0", nil, "12.11"},
		{"13-latest", "1.3.0", nil, "13.7"},
		{"14-latest", "1.3.0", nil, "14.4"},
	}

	for _, c := range cases {
		params := &version_service.VersionServiceApplyParams{
			Apply:           c.apply,
			OperatorVersion: c.operator,
			Product:         "pg-operator",
		}
		params.WithTimeout(2 * time.Second)
		if c.dbVersion != nil {
			params.DatabaseVersion = c.dbVersion
		}

		resp, err := cli.VersionService.VersionServiceApply(params)
		assert.NoError(t, err)

		v := getVersion(resp.Payload.Versions[0].Matrix.Postgresql)
		assert.Equal(t, c.version, v)
	}
}

func TestApplyPSReturnedVersions(t *testing.T) {
	cli := cli()

	cases := []struct {
		apply    string
		operator string
		version  string
	}{
		// test latest
		{"latest", "0.2.0", "8.0.28-20"},

		// test recommended
		{"recommended", "0.2.0", "8.0.28-20"},

		// test exact
		{"8.0.27", "0.2.0", "8.0.27-18"},

		//test with suffix
		{"8.0-latest", "0.2.0", "8.0.28-20"},
		{"8.0-recommended", "0.2.0", "8.0.28-20"},
	}

	for _, c := range cases {
		params := &version_service.VersionServiceApplyParams{
			Apply:           c.apply,
			OperatorVersion: c.operator,
			Product:         "ps-operator",
		}
		params.WithTimeout(2 * time.Second)

		resp, err := cli.VersionService.VersionServiceApply(params)
		assert.NoError(t, err)

		v := getVersion(resp.Payload.Versions[0].Matrix.Mysql)
		assert.Equal(t, c.version, v)
	}
}

func TestPmmServerUnimplemented(t *testing.T) {
	cli := cli()

	params := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.8.0",
		Product:         "pmm-server",
	}
	params.WithTimeout(2 * time.Second)

	_, err := cli.VersionService.VersionServiceApply(params)
	assert.Error(t, err, "error expected - apply should not be implemented for pmm-server")
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
