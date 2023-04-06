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
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// nolint:unparam
// There may be some error returned here.
func runToolFlags() error {
	switch {
	case *flagToolPromConf != "":
		if err := promDebugger(*flagToolPromConf); err != nil {
			cp.Errorf("[E] %s\n", err)
			os.Exit(-1)
		}

		os.Exit(0)

	case *flagToolParseLineProtocol != "":
		if err := parseLineProto(); err != nil {
			os.Exit(1)
		} else {
			os.Exit(-1)
		}

	case *flagToolSetupCompleterScripts:
		setupCompleterScripts()
		os.Exit(0)

	case *flagToolCompleterScripts:
		showCompletionScripts()
		os.Exit(0)

	case *flagToolGrokQ:
		grokq()
		os.Exit(0)

	case *flagToolDefaultMainConfig:

		defconf := config.DefaultConfig()
		fmt.Println(defconf.String())
		os.Exit(0)

	case *flagToolCloudInfo:
		tryLoadMainCfg()
		info, err := showCloudInfo()
		if err != nil {
			cp.Errorf("[E] Get cloud info failed: %s\n", err.Error())
			os.Exit(-1)
		}

		var keys []string
		for k := range info {
			keys = append(keys, k)
		}

		sort.Strings(keys)
		for _, k := range keys {
			cp.Infof("\t% 24s: %v\n", k, info[k])
		}

		os.Exit(0)

	case *flagToolIPInfo != "":
		tryLoadMainCfg()
		x, err := ipInfo(*flagToolIPInfo)
		if err != nil {
			cp.Errorf("[E] get IP info failed: %s\n", err.Error())
		} else {
			for k, v := range x {
				cp.Infof("\t% 8s: %s\n", k, v)
			}
		}

		os.Exit(0)

	case *flagToolBugReport:
		tryLoadMainCfg()
		if err := bugReport(); err != nil {
			cp.Errorf("[E] export DataKit info failed: %s\n", err.Error())
		}
		os.Exit(0)

	case *flagToolWorkspaceInfo:
		tryLoadMainCfg()
		requrl := fmt.Sprintf("http://%s%s", config.Cfg.HTTPAPI.Listen, workspace)
		body, err := doWorkspace(requrl)
		if err != nil {
			cp.Errorf("get worksapceInfo fail %s\n", err.Error())
		}
		outputWorkspaceInfo(body)
		os.Exit(0)

	case *flagToolCheckConfig:
		confdir := FlagConfigDir
		if confdir == "" {
			tryLoadMainCfg()
			confdir = datakit.ConfdDir
		}

		if err := checkConfig(confdir, ".conf"); err != nil {
			os.Exit(-1)
		}
		os.Exit(0)

	case *flagToolTestSNMP != "":
		if !datakit.FileExist(*flagToolTestSNMP) {
			cp.Errorf("[E] File not exist: %s\n", *flagToolTestSNMP)
			return nil
		}

		if err := testSNMP(*flagToolTestSNMP); err != nil {
			os.Exit(-1)
		}
		os.Exit(0)

	case *flagToolDumpSamples != "":
		tryLoadMainCfg()
		fpath := *flagToolDumpSamples

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

	case *flagToolLoadLog:
		tryLoadMainCfg()
		cp.Infof("Upload log start...\n")
		if err := uploadLog(config.Cfg.Dataway.URLs); err != nil {
			cp.Errorf("[E] upload log failed : %s\n", err.Error())
			os.Exit(-1)
		}
		cp.Infof("Upload ok.\n")
		os.Exit(0)

	case *flagToolCheckSample:
		if err := checkSample(); err != nil {
			os.Exit(-1)
		}
		os.Exit(0)
	}

	return fmt.Errorf("unknown tool: %s", os.Args[2])
}
