// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLocalNodeName(t *testing.T) {
	tests := []struct {
		name, current, legacy, want string
		wantErr                     bool
	}{
		{name: "legacy fallback", legacy: "legacy-node", want: "legacy-node"},
		{name: "current takes precedence", current: "current-node", legacy: "legacy-node", want: "current-node"},
		{name: "missing", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("ENV_K8S_NODE_NAME", test.current)
			t.Setenv("NODE_NAME", test.legacy)

			got, err := GetLocalNodeName()
			if test.wantErr {
				require.EqualError(t, err, "invalid ENV_K8S_NODE_NAME environment, cannot be empty")
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}
