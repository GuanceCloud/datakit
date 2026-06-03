// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type CheckOptions struct {
	LogPath   string
	Config    bool
	ConfigDir string
	Sample    bool
}

func RunCheck(opts CheckOptions) error {
	ConfigureCommandLog(opts.LogPath)

	switch {
	case opts.Config || opts.ConfigDir != "":
		confdir := opts.ConfigDir
		if confdir == "" {
			tryLoadMainCfg()
			confdir = datakit.ConfdDir
		}

		if err := checkConfig(confdir, ".conf"); err != nil {
			return err
		}
		return nil

	case opts.Sample:
		if err := checkSample(); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("unknown check option")
}
