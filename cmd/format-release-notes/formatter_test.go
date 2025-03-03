package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatReleaseNotesMarkdown(t *testing.T) {
	type testCase struct {
		name     string
		markdown []byte
		expected []byte
	}
	cases := []testCase{
		{
			name: "transforms relative links",
			markdown: []byte(`### PMM 2.42.0

Welcome to PMM [v3.0.0](../index.md)

- [Helm](../install-pmm/install-pmm-server/baremetal/helm/index.md) (Technical Preview) | 1. Quick<br>2. Simple<br>3. Cloud-compatible <br> 4. Rootless| Requires running a Kubernetes cluster.

![Service Accounts page](../images/Service_Accounts.png)`),
			expected: []byte(`### PMM 2.42.0

Welcome to PMM [v3.0.0](https://docs.percona.com/percona-monitoring-and-management/3/index.html)

- [Helm](https://docs.percona.com/percona-monitoring-and-management/3/install-pmm/install-pmm-server/baremetal/helm/index.html) (Technical Preview) | 1. Quick<br>2. Simple<br>3. Cloud-compatible <br> 4. Rootless| Requires running a Kubernetes cluster.

![Service Accounts page](https://docs.percona.com/percona-monitoring-and-management/3/images/Service_Accounts.png)` + "\n"),
		},
		{
			name: "transforms icon variables",
			markdown: []byte(`Navigate to the **Main** menu and hover on the {{icon.inventory}} _Dashboards_ icon.
2. Click **New folder**.
3. Provide a name for your folder, and then select **Create**.
4. Navigate to {{icon.inventory}} _Dashboards_ from the **Main** menu and click **Browse**.`),
			expected: []byte("Navigate to the **Main** menu and hover on the <i class=\"uil uil-clipboard-notes\"></i> *Dashboards* icon." +
				" 2. Click **New folder**." +
				" 3. Provide a name for your folder, and then select **Create**." +
				" 4. Navigate to <i class=\"uil uil-clipboard-notes\"></i> *Dashboards* from the **Main** menu and click **Browse**.\n"),
		},
		{
			name: "transforms admonition variables",
			markdown: []byte(
				`
!!! hint alert alert-success "Tip"
Tip
!!! hint alert alert-success "Tips"
- One
- Two
	`),
			expected: []byte(
				`### Tip

Tip

### Tips
- One
- Two` + "\n"),
		}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := FormatReleaseNotes(tc.markdown)
			require.NoError(t, err)
			assert.Equal(t, string(tc.expected), string(output))
		})
	}
}

func TestExtractRelativeURL(t *testing.T) {
	type testCases struct {
		name           string
		link           string
		isRelativeLink bool
	}

	tests := []testCases{
		{
			name:           "relative link returns true",
			link:           "../index.md",
			isRelativeLink: true,
		},
		{
			name:           "absolute link returns false",
			link:           "https://docs.percona.com/index.md",
			isRelativeLink: false,
		},
		{
			name:           "hash links returns true",
			link:           "#heading",
			isRelativeLink: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, isRelativeLink := extractRelativeURL(tt.link)
			assert.Equal(t, tt.isRelativeLink, isRelativeLink)
		})
	}
}
