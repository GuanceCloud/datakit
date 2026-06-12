// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
)

type DebugOptions struct {
	LogPath                 string
	UploadLog               bool
	GlobConf                string
	RegexConf               string
	PromConf                string
	BugReport               bool
	BugreportOSS            string
	BugreportDataway        string
	BugreportDatawayEnabled bool
	BugreportDisableProfile bool
	BugreportNMetrics       int
	BugreportTag            string
	InputConf               string
	HTTPListen              string
	Filter                  string
	Data                    string
	KVFile                  string
}

func RunDebug(opts DebugOptions) error {
	ConfigureCommandLog(opts.LogPath)

	switch {
	case opts.Filter != "":
		if err := debugFilter([]byte(opts.Filter),
			[]byte(opts.Data)); err != nil {
			cp.Errorf("[E] %s\n", err.Error())
			return err
		}
		return nil

	case opts.InputConf != "":

		// Try load global settings, we need to load global-host/env tags
		// and applied to collected points. This makes the testing points
		// are the same as real point.
		tryLoadMainCfg()
		if err := config.Cfg.ApplyMainConfig(); err != nil {
			cp.Warnf("ApplyMainConfig: %s, ignored\n", err)
		}

		if err := debugInput(opts.InputConf, opts.KVFile, opts.HTTPListen); err != nil {
			cp.Errorf("[E] %s\n", err.Error())
			return err
		}

		return nil

	case opts.BugReport:
		tryLoadMainCfg()
		if err := bugReport(opts); err != nil {
			cp.Errorf("[E] export DataKit info failed: %s\n", err.Error())
			return err
		}
		return nil

	case opts.GlobConf != "":
		if err := globPath(opts.GlobConf); err != nil {
			cp.Errorf("[E] %s\n", err)
			return err
		}
		return nil

	case opts.RegexConf != "":
		if err := regexMatching(opts.RegexConf); err != nil {
			cp.Errorf("[E] %s\n", err)
			return err
		}
		return nil

	case opts.PromConf != "":
		if err := promDebugger(opts.PromConf); err != nil {
			cp.Errorf("[E] %s\n", err)
			return err
		}
		return nil

	case opts.UploadLog:
		tryLoadMainCfg()
		cp.Infof("Upload log start...\n")
		if err := uploadLog(config.Cfg.Dataway.URLs); err != nil {
			cp.Errorf("[E] upload log failed : %s\n", err.Error())
			return err
		}
		cp.Infof("Upload ok.\n")
		return nil
	}

	return fmt.Errorf("unknown debug option")
}
