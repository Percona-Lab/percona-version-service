package api_tests

import (
	"strings"
	"testing"
	"time"

	"github.com/Percona-Lab/percona-version-service/client"
	"github.com/Percona-Lab/percona-version-service/client/models"
	"github.com/Percona-Lab/percona-version-service/client/version_service"
)

func Test_apply_should_return_just_one_version(t *testing.T) {
	cli := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:    "0.0.0.0:11000",
		Schemes: []string{"http"},
	})

	pxcParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.5.0",
		Product:         "pxc-operator",
	}
	pxcParams.WithTimeout(2 * time.Second)

	pxcResp, err := cli.VersionService.VersionServiceApply(pxcParams)
	if err != nil {
		t.Fatal(err)
	}

	if len(pxcResp.Payload.Versions[0].Matrix.Pxc) != 1 ||
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
	if err != nil {
		t.Fatal(err)
	}

	if len(psmdbResp.Payload.Versions[0].Matrix.Mongod) != 1 ||
		len(psmdbResp.Payload.Versions[0].Matrix.Backup) != 1 ||
		len(psmdbResp.Payload.Versions[0].Matrix.Pmm) != 1 ||
		len(psmdbResp.Payload.Versions[0].Matrix.Operator) != 1 {
		t.Error("more than one version returned")
	}
}

func Test_apply_pxc_should_return_same_major_version(t *testing.T) {
	cli := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:    "0.0.0.0:11000",
		Schemes: []string{"http"},
	})

	pxcParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.5.0",
		Product:         "pxc-operator",
	}
	pxcParams.WithTimeout(2 * time.Second)

	for _, v := range []string{"5.7", "8.0"} {
		pxcParams.DatabaseVersion = &v
		pxcResp, err := cli.VersionService.VersionServiceApply(pxcParams)
		if err != nil {
			t.Fatal(err)
		}

		k := getVersion(pxcResp.Payload.Versions[0].Matrix.Pxc)
		if !strings.HasPrefix(k, v) {
			t.Errorf("wrong version returned: %s", k)
		}
	}
}

func Test_apply_psmdb_should_return_same_major_version(t *testing.T) {
	cli := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:    "0.0.0.0:11000",
		Schemes: []string{"http"},
	})

	psmdbParams := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.5.0",
		Product:         "psmdb-operator",
	}
	psmdbParams.WithTimeout(2 * time.Second)

	for _, v := range []string{"4.2", "4.0", "3.6"} {
		psmdbParams.DatabaseVersion = &v
		psmdbResp, err := cli.VersionService.VersionServiceApply(psmdbParams)
		if err != nil {
			t.Fatal(err)
		}

		k := getVersion(psmdbResp.Payload.Versions[0].Matrix.Mongod)
		if !strings.HasPrefix(k, v) {
			t.Errorf("wrong version returned: %s", k)
		}
	}
}

func getVersion(v map[string]models.VersionVersion) string {
	for k := range v {
		return k
	}
	return ""
}
