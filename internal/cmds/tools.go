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

	apmInstaller "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/installer"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type ToolOptions struct {
	GrokQ                         bool
	LogPath                       string
	ShowCloudInfo                 bool
	IPInfo                        string
	WorkspaceInfo                 bool
	DumpSamples                   string
	DefaultMainConf               bool
	ParseLineProtocol             string
	JSON                          bool
	UpdateIPDB                    bool
	ParseKVFile                   string
	KVFile                        string
	RemoveApmAutoInject           bool
	ChangeDockerContainersRuntime string
	IngestionCanary               bool
	IngestionCanaryIndex          string
}

// RunTool runs the selected tool command.
//
//nolint:unparam
func RunTool(opts ToolOptions) error {
	ConfigureCommandLog(opts.LogPath)

	switch {
	case opts.UpdateIPDB:
		if err := updateIPDB(); err != nil {
			return err
		}
		return nil

	case opts.ParseLineProtocol != "":
		if err := parseLineProto(opts.ParseLineProtocol, opts.JSON); err != nil {
			return err
		}
		return nil

	case opts.GrokQ:
		grokq()
		return nil

	case opts.DefaultMainConf:

		defconf := datakit.MainConfSample(datakit.BrandDomain)
		cp.Println(defconf)
		return nil

	case opts.ShowCloudInfo:
		tryLoadMainCfg()
		info, err := showCloudInfo()
		if err != nil {
			return fmt.Errorf("get cloud info failed: %w", err)
		}

		var keys []string
		for k := range info {
			keys = append(keys, k)
		}

		sort.Strings(keys)
		for _, k := range keys {
			cp.Infof("\t% 24s: %v\n", k, info[k])
		}

		return nil

	case opts.IPInfo != "":
		tryLoadMainCfg()
		x, err := ipInfo(opts.IPInfo)
		if err != nil {
			return fmt.Errorf("get IP info failed: %w", err)
		} else {
			for k, v := range x {
				cp.Infof("\t% 8s: %s\n", k, v)
			}
		}

		return nil

	case opts.WorkspaceInfo:
		tryLoadMainCfg()
		requrl := fmt.Sprintf("http://%s%s", config.Cfg.HTTPAPI.Listen, workspace)
		body, err := doWorkspace(requrl)
		if err != nil {
			return fmt.Errorf("get worksapceInfo fail %w", err)
		}
		outputWorkspaceInfo(body)
		return nil

	case opts.DumpSamples != "":
		tryLoadMainCfg()
		fpath := opts.DumpSamples

		if err := os.MkdirAll(fpath, datakit.ConfPerm); err != nil {
			return err
		}

		for k, v := range inputs.AllInputs {
			sample := v().SampleConfig()
			if err := os.WriteFile(filepath.Join(fpath, k+".conf"),
				[]byte(sample), datakit.ConfPerm); err != nil {
				return err
			}
		}
		return nil

	case opts.ParseKVFile != "":
		tryLoadMainCfg()
		kvPath := opts.KVFile
		if kvPath == "" {
			kvPath = datakit.KVFile
		}
		kv := config.GetKV()
		if err := kv.LoadKVFile(kvPath); err != nil {
			return fmt.Errorf("load kv file failed: %w", err)
		}
		data, err := os.ReadFile(filepath.Clean(opts.ParseKVFile))
		if err != nil {
			return fmt.Errorf("read file failed: %w", err)
		}

		replacedData, err := kv.ReplaceKV(string(data))
		if err != nil {
			return fmt.Errorf("replace kv failed: %w", err)
		}

		cp.Printf("%s", replacedData)
		return nil

	case opts.RemoveApmAutoInject:
		// cleanup apm inject
		if err := apmInstaller.Uninstall(
			apmInstaller.WithInstallDir(datakit.InstallDir)); err != nil {
			return fmt.Errorf("remove failed: %w", err)
		}
		if err := unsetDKConfAPMInst(datakit.MainConfPath); err != nil {
			return fmt.Errorf("clean up datakit config failed: %w", err)
		}
		return nil

	case opts.ChangeDockerContainersRuntime != "":
		var from, to string
		switch opts.ChangeDockerContainersRuntime {
		case apmInstaller.RuntimeDkRunc:
			from, to = apmInstaller.RuntimeRunc, apmInstaller.RuntimeDkRunc
		case apmInstaller.RuntimeRunc:
			from, to = apmInstaller.RuntimeDkRunc, apmInstaller.RuntimeRunc
		}
		if err := apmInstaller.ChangeDockerHostConfigRunc(from, to, ""); err != nil {
			return fmt.Errorf("change runtime of all containers from %s to %s failed: %w", from, to, err)
		}
		cp.Infof("change runtime of all containers from %s to %s succeeded\n",
			from, to)
		return nil

	case opts.IngestionCanary:
		tryLoadMainCfg()
		if err := runIngestionCanaryTool(opts.IngestionCanaryIndex); err != nil {
			return fmt.Errorf("ingestion canary tool failed: %w", err)
		}
		return nil
	}

	return fmt.Errorf("unknown tool option")
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

	if err := cfg.TryUpgradeCfg(path, config.DKConfBackupName(path)); err != nil {
		return err
	}
	return nil
}
