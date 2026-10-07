package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb/api"
)

func TestJsonpbObjectMarshalerEmitsZeroValues(t *testing.T) {
	tests := map[string]struct {
		req  *pbVersion.ApplyRequest
		want map[string]any
	}{
		"unset telemetry fields are emitted as zero values": {
			req: &pbVersion.ApplyRequest{Product: "psmdb-operator"},
			want: map[string]any{
				"product":                 "psmdb-operator",
				"vectorSearchEnabled":     false,
				"clusterSyncEnabled":      false,
				"encryptionEnabled":       false,
				"encryptionExplicitlySet": false,
				"ociBackupEnabled":        false,
				"alibabaBackupEnabled":    false,
				"arbiterEnabled":          false,
				"nonVotingEnabled":        false,
				"mongoTlsMode":            "",
				"mongosSize":              float64(0),
			},
		},
		"set telemetry fields are preserved": {
			req: &pbVersion.ApplyRequest{
				Product:                 "psmdb-operator",
				VectorSearchEnabled:     true,
				ClusterSyncEnabled:      true,
				EncryptionEnabled:       true,
				EncryptionExplicitlySet: true,
				OciBackupEnabled:        true,
				AlibabaBackupEnabled:    true,
				ArbiterEnabled:          true,
				NonVotingEnabled:        true,
				MongoTlsMode:            "requireTLS",
				MongosSize:              3,
			},
			want: map[string]any{
				"vectorSearchEnabled":     true,
				"clusterSyncEnabled":      true,
				"encryptionEnabled":       true,
				"encryptionExplicitlySet": true,
				"ociBackupEnabled":        true,
				"alibabaBackupEnabled":    true,
				"arbiterEnabled":          true,
				"nonVotingEnabled":        true,
				"mongoTlsMode":            "requireTLS",
				"mongosSize":              float64(3),
			},
		},
		"encryption disabled but explicitly set is distinguishable": {
			req: &pbVersion.ApplyRequest{
				EncryptionEnabled:       false,
				EncryptionExplicitlySet: true,
			},
			want: map[string]any{
				"encryptionEnabled":       false,
				"encryptionExplicitlySet": true,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			b, err := (&jsonpbObjectMarshaler{pb: tc.req}).MarshalJSON()
			require.NoError(t, err)

			var got map[string]any
			require.NoError(t, json.Unmarshal(b, &got))

			for key, want := range tc.want {
				require.Contains(t, got, key)
				assert.Equal(t, want, got[key])
			}
		})
	}
}
