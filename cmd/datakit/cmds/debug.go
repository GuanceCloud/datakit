// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// nolint:unparam
// There may be some error returned here.
func runDebugFlags() error {
	if *FlagDebugWorkDir != "" {
		datakit.SetWorkDir(*FlagDebugWorkDir)
		return nil
	}

	setCmdRootLog(*flagDebugLogPath)
	switch {
	case *flagDebugDefaultMainConfig:

		defconf := config.DefaultConfig()
		fmt.Println(defconf.String())
		os.Exit(0)

	case *flagDebugCloudInfo != "":
		tryLoadMainCfg()
		info, err := showCloudInfo(*flagDebugCloudInfo)
		if err != nil {
			errorf("[E] Get cloud info failed: %s\n", err.Error())
			os.Exit(-1)
		}

		var keys []string
		for k := range info {
			keys = append(keys, k)
		}

		sort.Strings(keys)
		for _, k := range keys {
			infof("\t% 24s: %v\n", k, info[k])
		}

		os.Exit(0)

	case *flagDebugIPInfo != "":
		tryLoadMainCfg()
		x, err := ipInfo(*flagDebugIPInfo)
		if err != nil {
			errorf("[E] get IP info failed: %s\n", err.Error())
		} else {
			for k, v := range x {
				infof("\t% 8s: %s\n", k, v)
			}
		}

		os.Exit(0)

	case *flagDebugWorkspaceInfo:
		tryLoadMainCfg()
		requrl := fmt.Sprintf("http://%s%s", config.Cfg.HTTPAPI.Listen, workspace)
		body, err := doWorkspace(requrl)
		if err != nil {
			errorf("get worksapceInfo fail %s\n", err.Error())
		}
		outputWorkspaceInfo(body)
		os.Exit(0)

	case *flagDebugCheckConfig:
		confdir := FlagConfigDir
		if confdir == "" {
			tryLoadMainCfg()
			confdir = datakit.ConfdDir
		}

		if err := checkConfig(confdir, ""); err != nil {
			os.Exit(-1)
		}
		os.Exit(0)

	case *flagDebugDumpSamples != "":
		tryLoadMainCfg()
		fpath := *flagDebugDumpSamples

		if err := os.MkdirAll(fpath, datakit.ConfPerm); err != nil {
			panic(err)
		}

		for k, v := range inputs.Inputs {
			sample := v().SampleConfig()
			if err := ioutil.WriteFile(filepath.Join(fpath, k+".conf"),
				[]byte(sample), datakit.ConfPerm); err != nil {
				panic(err)
			}
		}
		os.Exit(0)

	case *flagDebugLoadLog:
		tryLoadMainCfg()
		if config.Cfg.DataWayCfg == nil {
			infof("[E] upload log failed: Dataway is required")
			os.Exit(-1)
		}
		infof("Upload log start...\n")
		if err := uploadLog(config.Cfg.DataWayCfg.URLs); err != nil {
			errorf("[E] upload log failed : %s\n", err.Error())
			os.Exit(-1)
		}
		infof("Upload ok.\n")
		os.Exit(0)

	case *flagDebugCheckSample:
		if err := checkSample(); err != nil {
			os.Exit(-1)
		}
		os.Exit(0)
	}

	return nil
}
