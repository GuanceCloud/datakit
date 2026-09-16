// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package cmds

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestDarwinExternalInstallDir(t *testing.T) {
	for _, arch := range []string{datakit.OSArchDarwinAmd64, datakit.OSArchDarwinArm64} {
		t.Run(arch, func(t *testing.T) {
			require.Contains(t, ExternalInstallDir, arch)
			require.Equal(t, "/usr/local/", ExternalInstallDir[arch])
		})
	}
}
