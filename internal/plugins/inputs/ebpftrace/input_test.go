// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ebpftrace

import (
	T "testing"

	bstoml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestSampleConfigEscapesWindowsInstallDir(t *T.T) {
	origInstallDir := datakit.InstallDir
	t.Cleanup(func() {
		datakit.SetupWorkDir(origInstallDir)
	})

	datakit.SetupWorkDir(`C:\Program Files\datakit`)

	var conf struct {
		Inputs struct {
			EBPFTrace []Input `toml:"ebpftrace"`
		} `toml:"inputs"`
	}

	_, err := bstoml.Decode((&Input{}).SampleConfig(), &conf)
	require.NoError(t, err)
	require.Len(t, conf.Inputs.EBPFTrace, 1)
	require.Contains(t, conf.Inputs.EBPFTrace[0].DBPath, `C:\Program Files\datakit`)
}
