package server

import (
	"testing"

	"github.com/Masterminds/semver"
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

func TestPgImageNewer(t *testing.T) {
	must := func(s string) *semver.Version {
		t.Helper()
		v, err := semver.NewVersion(s)
		require.NoError(t, err)
		return v
	}

	assert.True(t, pgImageNewer(must("17.11.1-3"), must("17.11.1")))
	assert.False(t, pgImageNewer(must("17.11.1"), must("17.11.1-3")))
	assert.True(t, pgImageNewer(must("17.11.1-10"), must("17.11.1-9")))
	assert.False(t, pgImageNewer(must("17.11.1-9"), must("17.11.1-10")))
	assert.True(t, pgImageNewer(must("18.0.0"), must("17.11.1-10")))
	assert.False(t, pgImageNewer(must("17.11.1-10"), must("18.0.0")))

	sorted, err := sortedPGImageVersionsDesc([]string{"17.11.1-9", "17.11.1", "17.11.1-10"})
	require.NoError(t, err)
	assert.Equal(t, []string{"17.11.1-10", "17.11.1-9", "17.11.1"}, []string{
		sorted[0].Original(), sorted[1].Original(), sorted[2].Original(),
	})
}

func TestPgDepFilterImageRebuildKeys(t *testing.T) {
	andRange := func(lo, hi string) map[string]interface{} {
		return map[string]interface{}{
			"and": []interface{}{
				map[string]interface{}{">=": []interface{}{map[string]interface{}{"var": "productVersion"}, lo}},
				map[string]interface{}{"<": []interface{}{map[string]interface{}{"var": "productVersion"}, hi}},
			},
		}
	}

	deps := map[string]interface{}{
		"18.6.1-3": map[string]interface{}{">=": []interface{}{map[string]interface{}{"var": "productVersion"}, "18.6.1-3"}},
		"18.6.1":   andRange("18.6.1", "18.6.1-3"),
		"17.11.1-3": andRange("17.11.1-3", "18.0"),
		"17.11.1":   andRange("17.11.1", "17.11.1-3"),
		"16.15-3":   andRange("16.15-3", "17.0"),
		"16.15":     andRange("16.15", "16.15-3"),
		"15.19-3":   andRange("15.19-3", "16.0"),
		"15.19":     andRange("15.19", "15.19-3"),
		"14.24-3":   andRange("14.24-3", "15.0"),
		"14.24":     andRange("14.24", "14.24-3"),
	}

	cases := []struct {
		productVersion string
		want           string
	}{
		{"18.6.1-3", "18.6.1-3"},
		{"18.6.1", "18.6.1"},
		{"17.11.1-3", "17.11.1-3"},
		{"17.11.1", "17.11.1"},
		{"16.15-3", "16.15-3"},
		{"16.15", "16.15"},
		{"15.19-3", "15.19-3"},
		{"15.19", "15.19"},
		{"14.24-3", "14.24-3"},
		{"14.24", "14.24"},
	}

	for _, c := range cases {
		t.Run(c.productVersion, func(t *testing.T) {
			got, err := pgDepFilter(deps, c.productVersion)
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}
}
