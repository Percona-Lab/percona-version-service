package api_tests

import (
	"testing"
	"time"

	"github.com/Percona-Lab/percona-version-service/client"
	"github.com/Percona-Lab/percona-version-service/client/version_service"
)

func Test_apply_should_return_just_one_version(t *testing.T) {
	cli := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:    "0.0.0.0:11000",
		Schemes: []string{"http"},
	})
	params := &version_service.VersionServiceApplyParams{
		Apply:           "latest",
		OperatorVersion: "1.5.0",
		Product:         "pxc-operator",
	}
	params.WithTimeout(2 * time.Second)

	resp, err := cli.VersionService.VersionServiceApply(params)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Payload.Versions[0].Matrix.Pxc) != 1 {
		t.Error("more than one version returned")
	}
}
