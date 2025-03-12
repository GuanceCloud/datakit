// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	apmInj "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/utils"
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
		cp.Println(defconf.String())
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
			if err := os.WriteFile(filepath.Join(fpath, k+".conf"),
				[]byte(sample), datakit.ConfPerm); err != nil {
				panic(err)
			}
		}
		os.Exit(0)

	case *flagToolParseKVFile != "":
		tryLoadMainCfg()
		kvPath := *flagToolKVFile
		if kvPath == "" {
			kvPath = datakit.KVFile
		}
		kv := config.GetKV()
		if err := kv.LoadKVFile(kvPath); err != nil {
			cp.Errorf("load kv file failed: %s\n", err.Error())
			os.Exit(-1)
		}
		data, err := os.ReadFile(filepath.Clean(*flagToolParseKVFile))
		if err != nil {
			cp.Errorf("read file failed: %s\n", err.Error())
			os.Exit(-1)
		}

		replacedData, err := kv.ReplaceKV(string(data))
		if err != nil {
			cp.Errorf("replace kv failed: %s\n", err.Error())
			os.Exit(-1)
		}

		cp.Printf("%s", replacedData)
		os.Exit(0)

	case *flagToolRemoveApmAutoInject:
		// cleanup apm inject
		if err := apmInj.Uninstall(
			apmInj.WithInstallDir(datakit.InstallDir)); err != nil {
			cp.Errorf("remove failed: %s\n", err.Error())
		}
		if err := unsetDKConfAPMInst(datakit.MainConfPath); err != nil {
			cp.Errorf("clean up datakit config failed: %s\n", err.Error())
		}
		os.Exit(0)

	case *flagToolChangeDockerContainersRuntime != "":
		var from, to string
		switch *flagToolChangeDockerContainersRuntime {
		case apmInj.RuntimeDkRunc:
			from, to = apmInj.RuntimeRunc, apmInj.RuntimeDkRunc
		case apmInj.RuntimeRunc:
			from, to = apmInj.RuntimeDkRunc, apmInj.RuntimeRunc
		}
		if err := apmInj.ChangeDockerHostConfigRunc(from, to, ""); err != nil {
			cp.Errorf("change runtime of all containers from %s to %s failed: %s\n",
				from, to, err.Error())
		} else {
			cp.Infof("change runtime of all containers from %s to %s succeeded\n",
				from, to)
		}
		os.Exit(0)
	}

	return fmt.Errorf("unknown tool option: %s", os.Args[2])
}

func unsetDKConfAPMInst(path string) error {
	var cfg config.Config
	err := cfg.LoadMainTOML(path)
	if err != nil {
		return err
	}

	if cfg.APMInject != nil &&
		cfg.APMInject.InstrumentationEnabled != "" &&
		cfg.APMInject.InstrumentationEnabled != "disable" {
		cfg.APMInject.InstrumentationEnabled = ""
	} else {
		return nil
	}

	if err := cfg.TryUpgradeCfg(path); err != nil {
		return err
	}
	return nil
}
