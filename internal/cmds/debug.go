// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"os"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
)

func runDebugFlags() error {
	switch {
	case *flagDebugFilter != "":
		if err := debugFilter([]byte(*flagDebugFilter),
			[]byte(*flagDebugData)); err != nil {
			cp.Errorf("[E] %s\n", err.Error())
			return err
		}
		return nil

	case *flagDebugInputConf != "":

		// Try load global settings, we need to load global-host/env tags
		// and applied to collected points. This makes the testing points
		// are the same as real point.
		tryLoadMainCfg()
		if err := config.Cfg.ApplyMainConfig(); err != nil {
			cp.Warnf("ApplyMainConfig: %s, ignored\n", err)
		}

		if err := debugInput(*flagDebugInputConf); err != nil {
			cp.Errorf("[E] %s\n", err.Error())
			return err
		}

		return nil

	case *flagDebugBugReport:
		tryLoadMainCfg()
		if err := bugReport(); err != nil {
			cp.Errorf("[E] export DataKit info failed: %s\n", err.Error())
			return err
		}
		return nil

	case *flagDebugGlobConf != "":
		if err := globPath(*flagDebugGlobConf); err != nil {
			cp.Errorf("[E] %s\n", err)
			return err
		}
		return nil

	case *flagDebugRegexConf != "":
		if err := regexMatching(*flagDebugRegexConf); err != nil {
			cp.Errorf("[E] %s\n", err)
			return err
		}
		return nil

	case *flagDebugPromConf != "":
		if err := promDebugger(*flagDebugPromConf); err != nil {
			cp.Errorf("[E] %s\n", err)
			return err
		}
		return nil

	case *flagDebugLoadLog:
		tryLoadMainCfg()
		cp.Infof("Upload log start...\n")
		if err := uploadLog(config.Cfg.Dataway.URLs); err != nil {
			cp.Errorf("[E] upload log failed : %s\n", err.Error())
			return err
		}
		cp.Infof("Upload ok.\n")
		return nil
	}

	return fmt.Errorf("unknown debug option: %s", os.Args[1])
}
