// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/checkutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
)

var (
	// envVarRe is a regex to find environment variables in the config file.
	envVarRe      = regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`)
	envVarEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
)

func isValidDataway(c *Config) bool {
	if c.DataWayCfg == nil {
		l.Error("config.Cfg.DataWayCfg == nil")
		return false
	}

	if len(c.DataWayCfg.URLs) == 0 {
		l.Error("len(config.Cfg.DataWayCfg.URLs) == 0")
		return false
	}

	return true
}

func isValidSink(c *Config) bool {
	if c.Sinks == nil {
		l.Error("config.Cfg.Sinks == nil")
		return false
	}

	if len(c.Sinks.Sink) == 0 {
		l.Error("len(config.Cfg.Sinks.Sink) == 0")
		return false
	}

	empty := true
	for _, v := range c.Sinks.Sink {
		if _, ok := v["target"]; ok {
			empty = false
			break
		}
	}

	return !empty
}

func LoadCfg(c *Config, mcp string) error {
	datakit.InitDirs()

	if datakit.Docker { // only accept configs from ENV under docker(or daemon-set) mode
		if runtime.GOOS != datakit.OSWindows && runtime.GOOS != datakit.OSLinux {
			return fmt.Errorf("docker mode not supported under %s", runtime.GOOS)
		}

		if err := c.LoadEnvs(); err != nil {
			return err
		}

		// 这里暂时用 hostname 当做 datakit ID, 后续肯定会移除掉, 即 datakit ID 基本已经废弃不用了,
		// 中心最终将通过统计主机个数作为 datakit 数量来收费.
		// 由于 datakit UUID 不再重要, 出错也不管了
		_ = c.SetUUID()

		if err := CreateSymlinks(); err != nil {
			l.Warnf("CreateSymlinks: %s, ignored", err)
		}

		// We need a datakit.conf in docker mode when run datakit commands.
		// See cmd/datakit/cmds/flags.go
		if err := c.InitCfg(datakit.MainConfPath); err != nil {
			l.Warnf("InitCfg: %s, ignored", err.Error())
		}
	} else if err := c.LoadMainTOML(mcp); err != nil {
		return err
	}

	l.Debugf("apply main configure...")

	if err := c.ApplyMainConfig(); err != nil {
		return err
	}

	// check dataway and sink config
	checkutil.CheckConditionExit(func() bool {
		if !isValidDataway(c) && !isValidSink(c) {
			l.Errorf("dataway and sink not set")
			return false
		}

		return true
	})

	l.Infof("loaded main cfg: \n%s", c.String())

	// clear all samples before loading
	removeSamples()

	if err := initPluginSamples(); err != nil {
		return err
	}

	if err := initCfgSample(datakit.MainConfSamplePath); err != nil {
		l.Warnf("failed to init datakit main sample config: %s, ignored", err.Error())
	}

	if err := initPluginPipeline(); err != nil {
		l.Errorf("init plugin pipeline: %s", err.Error())
		return err
	}

	l.Infof("init %d default plugins...", len(c.DefaultEnabledInputs))
	initDefaultEnabledPlugins(c)

	loadInputsConfFromDirs(getConfRootPaths())
	return nil
}

func getConfRootPaths() []string {
	if GitHasEnabled() {
		return []string{datakit.GitReposRepoFullPath}
	} else {
		return []string{datakit.ConfdDir}
	}
}

func trimBOM(f []byte) []byte {
	return bytes.TrimPrefix(f, []byte("\xef\xbb\xbf"))
}

func feedEnvs(data []byte) []byte {
	data = trimBOM(data)

	parameters := envVarRe.FindAllSubmatch(data, -1)

	for _, parameter := range parameters {
		if len(parameter) != 3 {
			continue
		}

		var envvar []byte

		switch {
		case parameter[1] != nil:
			envvar = parameter[1]
		case parameter[2] != nil:
			envvar = parameter[2]
		default:
			continue
		}

		envval, ok := os.LookupEnv(strings.TrimPrefix(string(envvar), "$"))
		// 找到了环境变量就替换否则不替换
		if ok {
			envval = envVarEscaper.Replace(envval)
			data = bytes.Replace(data, parameter[0], []byte(envval), 1)
		}
	}

	return data
}

func GetPipelinePath(category, pipeLineName string) (string, error) {
	if pipeLineName == "" {
		// you shouldn't be here, check before you call this function.
		return "", fmt.Errorf("pipeline_empty")
	}

	pipeLineName = dkstring.TrimString(pipeLineName)

	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/509
	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/524
	if filepath.IsAbs(pipeLineName) {
		return "", fmt.Errorf("pipeline in absolutely path is discouraged")
	}

	// start search pipeline_remote
	{
		plPath := filepath.Join(datakit.PipelineRemoteDir, pipeLineName)
		if _, err := os.Stat(plPath); err == nil {
			return plPath, nil
		}
	}

	{
		files := script.SearchPlFilePathFromPlStructPath(datakit.GitReposRepoFullPath)
		if f, ok := files[category]; ok {
			if plPath, ok := f[pipeLineName]; ok {
				if _, err := os.Stat(plPath); err != nil {
					l.Errorf("os.Stat failed: %v", err)
				} else {
					return plPath, nil // return once found the pipeline file
				}
			}
		}
	}

	files := script.SearchPlFilePathFromPlStructPath(datakit.PipelineDir)
	if f, ok := files[category]; ok {
		if plPath, ok := f[pipeLineName]; ok {
			if _, err := os.Stat(plPath); err != nil {
				return "", err
			} else {
				return plPath, nil // return once found the pipeline file
			}
		}
	}

	// search datakit root pipeline
	plPath := filepath.Join(datakit.PipelineDir, pipeLineName)
	if _, err := os.Stat(plPath); err != nil {
		return "", err
	}

	return plPath, nil
}

func InitGitreposDir() {
	// search enabled gitrepos
	for _, v := range Cfg.GitRepos.Repos {
		if !v.Enable {
			continue
		}
		v.URL = dkstring.TrimString(v.URL)
		if v.URL == "" {
			continue
		}
		repoName, err := path.GetGitPureName(v.URL)
		if err != nil {
			continue
		}
		datakit.GitReposRepoName = repoName
		datakit.GitReposRepoFullPath, err = GetGitRepoDir(repoName)
		if err != nil {
			l.Errorf("GetGitRepoDir failed: err = %v, repoName = %s", err, repoName)
		}
	}
}

func GetNamespacePipelineFiles(namespace string) ([]string, error) {
	if namespace == datakit.StrGitRepos {
		return path.GetSuffixFilesFromDirDeepOne(datakit.GitReposRepoFullPath, datakit.StrPipelineFileSuffix)
	}
	return nil, fmt.Errorf("invalid namespace")
}
