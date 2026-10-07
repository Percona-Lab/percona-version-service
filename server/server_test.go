package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb/api"
)

func TestApplyRequestLoggingOmitsUnsetFields(t *testing.T) {
	tests := map[string]struct {
		req        *pbVersion.ApplyRequest
		wantFields map[string]any
		wantAbsent []string
	}{
		"set telemetry fields are preserved": {
			req: &pbVersion.ApplyRequest{
				Product:              "psmdb-operator",
				VectorSearchEnabled:  true,
				ClusterSyncEnabled:   true,
				EncryptionEnabled:    true,
				OciBackupEnabled:     true,
				AlibabaBackupEnabled: true,
				ArbiterEnabled:       true,
				NonVotingEnabled:     true,
				MongoTlsMode:         "requireTLS",
				MongosSize:           3,
			},
			wantFields: map[string]any{
				"vectorSearchEnabled":  true,
				"clusterSyncEnabled":   true,
				"encryptionEnabled":    true,
				"ociBackupEnabled":     true,
				"alibabaBackupEnabled": true,
				"arbiterEnabled":       true,
				"nonVotingEnabled":     true,
				"mongoTlsMode":         "requireTLS",
				"mongosSize":           float64(3),
			},
		},
		"unset telemetry fields are omitted entirely": {
			req:        &pbVersion.ApplyRequest{Product: "psmdb-operator", OperatorVersion: "1.23.0"},
			wantFields: map[string]any{"product": "psmdb-operator", "operatorVersion": "1.23.0"},
			wantAbsent: []string{
				"vectorSearchEnabled", "clusterSyncEnabled", "encryptionEnabled",
				"ociBackupEnabled", "alibabaBackupEnabled",
				"arbiterEnabled", "nonVotingEnabled", "mongoTlsMode", "mongosSize",
			},
		},
		"pxc request carries no mongo or pg specific fields": {
			req: &pbVersion.ApplyRequest{
				Product:         "pxc-operator",
				OperatorVersion: "1.20.0",
				ClusterSize:     3,
				BackupsEnabled:  true,
				ProxysqlVersion: "2.7.1",
			},
			wantFields: map[string]any{
				"product":         "pxc-operator",
				"proxysqlVersion": "2.7.1",
				"backupsEnabled":  true,
				"clusterSize":     float64(3),
			},
			wantAbsent: []string{"arbiterEnabled", "nonVotingEnabled", "mongoTlsMode", "mongosSize", "distribution"},
		},
		"pg request carries its own distribution and no mongo fields": {
			req: &pbVersion.ApplyRequest{
				Product:         "pg-operator",
				OperatorVersion: "3.1.0",
				Distribution:    "percona",
			},
			wantFields: map[string]any{"product": "pg-operator", "distribution": "percona"},
			wantAbsent: []string{"vectorSearchEnabled", "arbiterEnabled", "mongoTlsMode", "mongosSize"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			b, err := (&jsonpbObjectMarshaler{pb: tc.req}).MarshalJSON()
			require.NoError(t, err)

			var got map[string]any
			require.NoError(t, json.Unmarshal(b, &got))

			for key, want := range tc.wantFields {
				require.Contains(t, got, key)
				assert.Equal(t, want, got[key])
			}
			for _, key := range tc.wantAbsent {
				assert.NotContains(t, got, key)
			}
		})
	}
}
