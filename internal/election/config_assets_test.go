// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOperatorElectionConfigAssetsStayInSync(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	requiredByFile := map[string][]string{
		"templates/datakit.template.yaml": {
			"ENV_ELECTION_OPERATOR_URL",
		},
		"templates/datakit-gke-autopilot.template.yaml": {
			"ENV_ELECTION_OPERATOR_URL",
		},
		"templates/charts-values.template.yaml": {
			`election_operator_url: ""`,
		},
		"templates/charts-values-gke-autopilot.template.yaml": {
			`election_operator_url: ""`,
		},
		"charts/datakit/templates/daemonset.yaml": {
			".Values.datakit.election_operator_url",
			"ENV_ELECTION_OPERATOR_URL",
		},
		"internal/export/doc/en/election.md": {
			"operator_url",
			"ENV_ELECTION_OPERATOR_URL",
		},
		"internal/export/doc/zh/election.md": {
			"operator_url",
			"ENV_ELECTION_OPERATOR_URL",
		},
		"internal/export/doc/en/datakit-helm.md": {
			`election_operator_url: ""`,
		},
		"internal/export/doc/zh/datakit-helm.md": {
			`election_operator_url: ""`,
		},
	}

	for relativePath, required := range requiredByFile {
		data, err := os.ReadFile(filepath.Join(repositoryRoot, relativePath))
		if err != nil {
			t.Fatalf("read %s: %v", relativePath, err)
		}
		content := string(data)
		for _, snippet := range required {
			if !strings.Contains(content, snippet) {
				t.Errorf("%s does not contain %q", relativePath, snippet)
			}
		}
	}
}
