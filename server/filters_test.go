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
