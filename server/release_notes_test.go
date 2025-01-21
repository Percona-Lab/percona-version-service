package server

import (
	"embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/fs"
	"path/filepath"
	"testing"

	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb/api"
)

//go:embed release_notes_test
var testReleaseNotesFS embed.FS

func TestGetReleaseNote(t *testing.T) {
	t.Parallel()
	sub, err := fs.Sub(testReleaseNotesFS, "release_notes_test")
	require.NoError(t, err)

	tests := []struct {
		name             string
		product          string
		version          string
		expectErr        bool
		expectedResponse *pbVersion.GetReleaseNotesResponse
	}{
		{
			name:      "returns correct release notes for pmm-server 2.42.0",
			product:   "pmm",
			version:   "2.42.0",
			expectErr: false,
			expectedResponse: &pbVersion.GetReleaseNotesResponse{
				Version:     "2.42.0",
				Product:     "pmm",
				ReleaseNote: "### PMM 2.42.0\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewReleaseNotes(sub)

			got, err := r.GetReleaseNote(tt.product, tt.version)
			if tt.expectErr {
				assert.NotNil(t, err)
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResponse.ReleaseNote, got.ReleaseNote)
		})
	}
}

func TestTransformMarkdownVariables(t *testing.T) {
	type testCase struct {
		name     string
		markdown string
		expected string
	}
	cases := []testCase{{
		name: "transforms icon variables",
		markdown: `Navigate to the **Main** menu and hover on the {{icon.inventory}} _Dashboards_ icon.
2.	Click **New folder**.
3.	Provide a name for your folder, and then select **Create**.
4.	Navigate to {{icon.inventory}} _Dashboards_ from the **Main** menu and click **Browse**.`,
		expected: "Navigate to the **Main** menu and hover on the <i class=\"uil uil-clipboard-notes\"></i> *Dashboards* icon." +
			"2. Click **New folder**." +
			"3. Provide a name for your folder, and then select **Create**." +
			"4. Navigate to <i class=\"uil uil-clipboard-notes\"></i> *Dashboards* from the **Main** menu and click **Browse**.\n",
	},
		{
			name: "transforms admonition variables",
			markdown: `!!! hint alert alert-success "Tip"
	    Tip
!!! hint alert alert-success "Tips"
- One
- Two
	`,
			expected: `### Tip
Tip
### Tips
- One
- Two`,
		}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := FormatReleaseNotes([]byte(tc.markdown))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(output))
		})
	}
}

func TestTransformReleaseNotes(t *testing.T) {
	t.Parallel()

	sub, err := fs.Sub(testReleaseNotesFS, "release_notes_test")
	require.NoError(t, err)

	path := filepath.Join(".", "pmm", "2.43.0-rel-links.md")
	b, err := fs.ReadFile(sub, path) //nolint:gosec
	require.NoError(t, err)

	output, err := FormatReleaseNotes(b)
	require.NoError(t, err)

	expected := "### PMM 2.42.0\n\n" +
		"Welcome to PMM [v2.42](https://github.com/percona/pmm-doc/tree/main/docs/index.md) " +
		"-![!image](https://docs.percona.com/percona-monitoring-and-management/_images/Max_Connection_Limit.png)\n"
	assert.Equal(t, expected, string(output))
}

func TestIsRelativeLink(t *testing.T) {
	type testCases struct {
		name     string
		link     string
		expected bool
	}

	tests := []testCases{
		{
			name:     "relative link returns true",
			link:     "../index.md",
			expected: true,
		},
		{
			name:     "absolute link returns false",
			link:     "https://docs.percona.com/index.md",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRelativeLink(tt.link)
			assert.Equal(t, tt.expected, got)
		})
	}
}
