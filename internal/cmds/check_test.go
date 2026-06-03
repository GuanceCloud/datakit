// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	T "testing"

	"github.com/stretchr/testify/require"
)

func TestRunCheckConfigDirWithoutConfigFlag(t *T.T) {
	err := RunCheck(CheckOptions{
		ConfigDir: t.TempDir(),
	})
	require.NoError(t, err)
}
