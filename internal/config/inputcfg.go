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
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const catalog = "samples"

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
	input := c()

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
func initPluginSamples(ipts map[string]inputs.Creator) error {
	for name, create := range ipts {
		if err := initDatakitConfSample(name, create); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) initDefaultEnabledPlugins(confDir string, ipts map[string]inputs.Creator) {
	if len(c.DefaultEnabledInputs) == 0 {
		l.Debug("no default inputs enabled")
		return
	}

	for _, name := range c.DefaultEnabledInputs {
		l.Debugf("init default input %s conf...", name)

		var (
			confPath, sample string
			ipt              inputs.Input
		)

		if c, ok := ipts[name]; ok {
			ipt = c()
			sample = ipt.SampleConfig()

			confPath = filepath.Join(confDir, name+".conf")
		} else {
			l.Warnf("input %s not found, ignored", name)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(confPath), datakit.ConfPerm); err != nil {
			l.Errorf("mkdir failed: %s, ignored", err.Error())
			continue
		}

		// check exist
		if fi, err := os.Stat(confPath); err == nil {
			if fi.IsDir() { // for configmap in k8s, the conf(such as zipkin.conf) is a dir.
				newfpath := filepath.Join(confDir, name+"-0xdeadbeaf"+".conf") // add suffix to filename
				l.Warnf("%q is dir, rename conf file to %q", confPath, newfpath)
				confPath = newfpath
			} else {
				l.Infof("default enabled input %q exists(%q), skipped", name, confPath)
				continue
			}
		}

		if err := ioutil.WriteFile(confPath, []byte(sample), datakit.ConfPerm); err != nil {
			l.Errorf("write input %s config failed: %s, ignored", name, err.Error())
			continue
		}

		l.Infof("enable input %s(conf: %q)ok", name, confPath)
	}
}

func (c *Config) inputDisabled(name string) bool {
	for _, elem := range c.DefaultEnabledInputs {
		if "-"+name == elem {
			return true
		}
	}
	return false
}

func (c *Config) loadInputsConfFromDirs(paths []string, ipts map[string]inputs.Creator) {
	inputs.ResetInputs()

	l.Infof("load input confs from %s", paths)
	for _, rp := range paths {
		for name, arr := range LoadInputConf(rp, ipts) {
			if c.inputDisabled(name) {
				l.Infof("input %q disabled", name)
				continue
			}

			for _, x := range arr {
				l.Infof("load input %q from conf file", name)
				inputs.AddInput(name, x)
			}
		}
	}

	if GitHasEnabled() {
		l.Infof("DefaultEnabledInputs: %s", strings.Join(Cfg.DefaultEnabledInputs, ","))
		enableDefaultInputs(c.DefaultEnabledInputs, ipts)
	}

	inputs.Init()
}

func enableDefaultInputs(list []string, ipts map[string]inputs.Creator) {
	for _, name := range list {
		if c, ok := ipts[name]; ok {
			i := c()
			inputInstances, err := LoadSingleConf(i.SampleConfig(), ipts)
			if err != nil {
				l.Errorf("LoadSingleConf failed: %v", err)
				continue
			}

			for _, arr := range inputInstances {
				for _, ipt := range arr {
					l.Infof("add input name: %s ", name)
					inputs.AddInput(name, ipt)
				}
			}
		}
	}
}

func ReloadCheckInputCfg() ([]*inputs.InputInfo, error) {
	var availableInputs []*inputs.InputInfo
	confRootPath := getConfRootPaths()

	for _, rp := range confRootPath {
		for _, arr := range LoadInputConf(rp, inputs.AllInputs) {
			availableInputs = append(availableInputs, arr...)
		}
	}

	return availableInputs, nil
}

func ReloadInputConfig() error {
	Cfg.loadInputsConfFromDirs(getConfRootPaths(), inputs.AllInputs)
	return nil
}
