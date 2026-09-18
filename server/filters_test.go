package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb/api"
)

func TestPMMFilter(t *testing.T) {
	tests := map[string]struct {
		versions map[string]*pbVersion.Version
		expected map[string]*pbVersion.Version
	}{
		"no versions": {
			versions: map[string]*pbVersion.Version{},
			expected: map[string]*pbVersion.Version{},
		},
		"single PMM2 version": {
			versions: map[string]*pbVersion.Version{
				"2.44.1-1": nil,
			},
			expected: map[string]*pbVersion.Version{
				"2.44.1-1": nil,
			},
		},
		"multiple PMM2 versions": {
			versions: map[string]*pbVersion.Version{
				"2.43.0":   nil,
				"2.44.0":   nil,
				"2.44.1-1": nil,
			},
			expected: map[string]*pbVersion.Version{
				"2.44.1-1": nil,
			},
		},
		"single PMM3 version": {
			versions: map[string]*pbVersion.Version{
				"3.3.1": nil,
			},
			expected: map[string]*pbVersion.Version{
				"3.3.1": nil,
			},
		},
		"multiple PMM3 versions": {
			versions: map[string]*pbVersion.Version{
				"3.2.0": nil,
				"3.3.0": nil,
				"3.3.1": nil,
			},
			expected: map[string]*pbVersion.Version{
				"3.3.1": nil,
			},
		},
		"one PMM2 and one PMM3 version": {
			versions: map[string]*pbVersion.Version{
				"2.44.1-1": nil,
				"3.3.1":    nil,
			},
			expected: map[string]*pbVersion.Version{
				"2.44.1-1": nil,
				"3.3.1":    nil,
			},
		},
		"multiple PMM2 and PMM3 versions": {
			versions: map[string]*pbVersion.Version{
				"2.43.0":   nil,
				"2.44.0":   nil,
				"2.44.1-1": nil,
				"3.2.0":    nil,
				"3.3.0":    nil,
				"3.3.1":    nil,
			},
			expected: map[string]*pbVersion.Version{
				"2.44.1-1": nil,
				"3.3.1":    nil,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := pmmFilter(tt.versions, true)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.versions)
		})
	}
}

func TestPSFilter(t *testing.T) {
	// mirrors the ps-operator 1.3.0 matrix: 8.0 and 8.4 carry a recommended
	// version, the freshly added 9.7 branch is available only
	matrix := func() map[string]*pbVersion.Version {
		return map[string]*pbVersion.Version{
			"9.7.1-1":   {Status: pbVersion.Status_available},
			"8.4.10-10": {Status: pbVersion.Status_recommended},
			"8.4.8-8":   {Status: pbVersion.Status_available},
			"8.0.46-37": {Status: pbVersion.Status_recommended},
			"8.0.45-36": {Status: pbVersion.Status_available},
		}
	}

	tests := map[string]struct {
		versions map[string]*pbVersion.Version
		apply    string
		current  string
		expected string
	}{
		"recommended stays on the latest recommended branch": {
			versions: matrix(), apply: "recommended", current: "", expected: "8.4.10-10",
		},
		"latest picks the newest branch": {
			versions: matrix(), apply: "latest", current: "", expected: "9.7.1-1",
		},
		"8.0-recommended": {
			versions: matrix(), apply: "recommended", current: "8.0", expected: "8.0.46-37",
		},
		"8.4-recommended": {
			versions: matrix(), apply: "recommended", current: "8.4", expected: "8.4.10-10",
		},
		"9.7-recommended falls back to the newest 9.7": {
			versions: matrix(), apply: "recommended", current: "9.7", expected: "9.7.1-1",
		},
		"9.7-latest": {
			versions: matrix(), apply: "latest", current: "9.7", expected: "9.7.1-1",
		},
		"branch without a recommended version resolves to its newest": {
			versions: map[string]*pbVersion.Version{
				"9.7.2-2":   {Status: pbVersion.Status_available},
				"9.7.1-1":   {Status: pbVersion.Status_available},
				"8.4.10-10": {Status: pbVersion.Status_recommended},
			},
			apply: "recommended", current: "9.7", expected: "9.7.2-2",
		},
		"disabled versions are skipped in the fallback": {
			versions: map[string]*pbVersion.Version{
				"9.7.2-2":   {Status: pbVersion.Status_disabled},
				"9.7.1-1":   {Status: pbVersion.Status_available},
				"8.4.10-10": {Status: pbVersion.Status_recommended},
			},
			apply: "recommended", current: "9.7", expected: "9.7.1-1",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := psFilter(tt.versions, tt.apply, tt.current)
			require.NoError(t, err)
			require.Len(t, tt.versions, 1)
			_, ok := tt.versions[tt.expected]
			assert.True(t, ok, "expected %s, got %v", tt.expected, keysOf(tt.versions))
		})
	}
}

func keysOf(m map[string]*pbVersion.Version) []string {
	k := make([]string, 0, len(m))
	for v := range m {
		k = append(k, v)
	}
	return k
}
