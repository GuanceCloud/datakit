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

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// nolint:unparam
// There may be some error returned here.
func runToolFlags() error {
	switch {
	case *flagToolUpdateIPDB:
		if err := updateIPDB(); err != nil {
			os.Exit(-1)
		} else {
			os.Exit(0)
		}

	case *flagToolParseLineProtocol != "":
		if err := parseLineProto(); err != nil {
			os.Exit(1)
		} else {
			os.Exit(0)
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

	case *flagToolWorkspaceInfo:
		tryLoadMainCfg()
		requrl := fmt.Sprintf("http://%s%s", config.Cfg.HTTPAPI.Listen, workspace)
		body, err := doWorkspace(requrl)
		if err != nil {
			cp.Errorf("get worksapceInfo fail %s\n", err.Error())
		}
		outputWorkspaceInfo(body)
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
	}

	return fmt.Errorf("unknown tool option: %s", os.Args[2])
}
