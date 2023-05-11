// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"os"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
)

func runDebugFlags() error {
	switch {
	case *flagDebugBugReport:

		tryLoadMainCfg()
		if err := bugReport(); err != nil {
			cp.Errorf("[E] export DataKit info failed: %s\n", err.Error())
		}
		os.Exit(0)

	case *flagDebugPromConf != "":

		if err := promDebugger(*flagDebugPromConf); err != nil {
			cp.Errorf("[E] %s\n", err)
			os.Exit(-1)
		}

		os.Exit(0)

	case *flagDebugLoadLog:
		tryLoadMainCfg()
		cp.Infof("Upload log start...\n")
		if err := uploadLog(config.Cfg.Dataway.URLs); err != nil {
			cp.Errorf("[E] upload log failed : %s\n", err.Error())
			os.Exit(-1)
		}
		cp.Infof("Upload ok.\n")
		os.Exit(0)

	case *flagDebugTestInput != "":
		if err := inputDebugger(*flagDebugTestInput); err != nil {
			l.Errorf("inputDebugger: %s", err)
		}
		os.Exit(0)
	}

	return fmt.Errorf("unknown debug option: %s", os.Args[1])
}
