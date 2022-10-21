// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func GetGitRepoDir(cloneDirName string) (string, error) {
	if cloneDirName == "" {
		// you shouldn't be here, check before you call this function.
		return "", fmt.Errorf("git repo clone dir empty")
	}
	return filepath.Join(datakit.GitReposDir, cloneDirName), nil
}

func GetGitRepoSubDir(cloneDirName, sonName string) (string, error) {
	if cloneDirName == "" {
		// you shouldn't be here, check before you call this function.
		return "", fmt.Errorf("git repo clone dir empty")
	}
	return filepath.Join(datakit.GitReposDir, cloneDirName, sonName), nil
}

var confsampleFingerprint = append([]byte(fmt.Sprintf(
	`# {"version": "%s", "desc": "do NOT edit this line"}`, datakit.Version)),
	byte('\n'))

func initDatakitConfSample(name string, c inputs.Creator) error {
	if name == datakit.DatakitInputName {
		return nil
	}

	input := c()
	catalog := input.Catalog()

	cfgpath := filepath.Join(datakit.ConfdDir, catalog, name+".conf.sample")
	l.Debugf("create datakit conf path %s", filepath.Join(datakit.ConfdDir, catalog))
	if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, catalog), datakit.ConfPerm); err != nil {
		l.Errorf("create catalog dir %s failed: %s", catalog, err.Error())
		return err
	}

	sample := input.SampleConfig()
	if sample == "" {
		return fmt.Errorf("no sample available on collector %s", name)
	}

	// 在 conf-sample 头部增加一些指纹信息.
	// 一般用户在编辑 conf 时，都是 copy 这个 sample 的。如果 sample 中带上指纹，
	// 那么最终的配置上也会带上这可能便于后续的升级，即升级程序能识别某个 conf
	// 的版本，进而进行指定的升级
	if err := ioutil.WriteFile(cfgpath, append(confsampleFingerprint, []byte(sample)...), datakit.ConfPerm); err != nil {
		l.Errorf("failed to create sample configure for collector %s: %s", name, err.Error())
		return err
	}

	return nil
}

// Creata datakit input plugin's configures if not exists.
func initPluginSamples() error {
	for name, create := range inputs.Inputs {
		if err := initDatakitConfSample(name, create); err != nil {
			return err
		}
	}
	return nil
}

func initDefaultEnabledPlugins(c *Config) {
	if len(c.DefaultEnabledInputs) == 0 {
		l.Debug("no default inputs enabled")
		return
	}

	if GitHasEnabled() {
		return // #501 issue
	}

	for _, name := range c.DefaultEnabledInputs {
		l.Debugf("init default input %s conf...", name)

		var fpath, sample string

		if c, ok := inputs.Inputs[name]; ok {
			i := c()
			sample = i.SampleConfig()

			fpath = filepath.Join(datakit.ConfdDir, i.Catalog(), name+".conf")
		} else {
			l.Warnf("input %s not found, ignored", name)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), datakit.ConfPerm); err != nil {
			l.Errorf("mkdir failed: %s, ignored", err.Error())
			continue
		}

		// check exist
		if _, err := os.Stat(fpath); err == nil {
			continue
		}

		if err := ioutil.WriteFile(fpath, []byte(sample), datakit.ConfPerm); err != nil {
			l.Errorf("write input %s config failed: %s, ignored", name, err.Error())
			continue
		}

		l.Infof("enable input %s ok", name)
	}
}

func loadInputsConfFromDirs(paths []string) {
	inputs.ResetInputs()

	l.Infof("load input confs from %s", paths)
	for _, rp := range paths {
		for name, arr := range LoadInputConf(rp) {
			for _, x := range arr {
				inputs.AddInput(name, x)
			}
		}
	}

	if GitHasEnabled() {
		enableDefaultInputs(Cfg.DefaultEnabledInputs)
	}

	inputs.AddSelf()

	inputs.Init()
}

func enableDefaultInputs(list []string) {
	for _, name := range list {
		if c, ok := inputs.Inputs[name]; ok {
			i := c()
			inputInstances, err := LoadSingleConf(i.SampleConfig(), inputs.Inputs)
			if err != nil {
				l.Errorf("LoadSingleConf failed: %v", err)
				continue
			}
			for _, arr := range inputInstances {
				for _, ipt := range arr {
					inputs.AddInput(name, ipt)
				}
			}
		}
	}
}

func ReloadCheckInputCfg() ([]inputs.Input, error) {
	var availableInputs []inputs.Input
	confRootPath := getConfRootPaths()

	for _, rp := range confRootPath {
		for _, arr := range LoadInputConf(rp) {
			availableInputs = append(availableInputs, arr...)
		}
	}

	return availableInputs, nil
}

func ReloadInputConfig() error {
	loadInputsConfFromDirs(getConfRootPaths())
	return nil
}
