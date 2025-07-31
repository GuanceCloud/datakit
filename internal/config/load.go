// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	plmanager "github.com/GuanceCloud/pipeline-go/manager"
	apmInstaller "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/installer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
)

var (
	// envVarRe is a regex to find environment variables in the config file.
	envVarRe      = regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`)
	re            = regexp.MustCompile(`ENC\[(.*?)\]`)
	envVarEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
)

func isValidDataway(c *Config) error {
	if c.Dataway == nil {
		return fmt.Errorf("dataway not set")
	}

	if len(c.Dataway.URLs) == 0 {
		return fmt.Errorf("dataway URL not set")
	}

	// TODO: check if dataway's sinkers ok

	return nil
}

func LoadCfg(c *Config, mcp string) error {
	datakit.InitDirs()

	if datakit.Docker { // only accept configs from ENV under docker(or daemon-set) mode
		if runtime.GOOS == datakit.OSWindows {
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

	l.Infof("apply main configure from %q...", mcp)

	if err := c.ApplyMainConfig(); err != nil {
		return err
	}

	// check dataway config
	if err := isValidDataway(c); err != nil {
		l.Errorf("ValidDataway: %s", err)
		return err
	}

	defaultKV.LoadKV()

	l.Infof("loaded main cfg: \n%s", c.String())

	// clear all samples before loading
	removeSamples()

	if err := initPluginSamples(inputs.AllInputs); err != nil {
		return err
	}

	if err := initCfgSample(datakit.MainConfSamplePath); err != nil {
		l.Warnf("failed to init datakit main sample config: %s, ignored", err.Error())
	}

	if err := initPluginPipeline(inputs.AllInputs); err != nil {
		l.Errorf("init plugin pipeline: %s", err.Error())
		return err
	}

	if c.APMInject != nil {
		if err := apmInstaller.Install(l,
			apmInstaller.WithInstallDir(datakit.InstallDir),
			apmInstaller.WithInstrumentationEnabled(
				c.APMInject.InstrumentationEnabled),
		); err != nil {
			l.Warnf("failed to install/uninstall apm inject: %s", err.Error())
		}
	}

	l.Infof("init %d default plugins...", len(c.DefaultEnabledInputs))

	if !GitHasEnabled() {
		// #501 issue
		c.initDefaultEnabledPlugins(datakit.ConfdDir, inputs.AllInputs)
	}

	c.loadInputsConfFromDirs(getConfRootPaths(), inputs.AllInputs)

	return nil
}

func getConfRootPaths() []string {
	if GitHasEnabled() {
		return []string{filepath.Join(datakit.GitReposRepoFullPath, datakit.GitRepoSubDirNameConfd)}
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
			l.Infof("load ENV %q:%q ok", envvar, envval)
			data = bytes.Replace(data, parameter[0], []byte(envval), 1)
		} else {
			l.Infof("load ENV %q failed, ignored", envvar)
		}
	}

	return data
}

func decodeEncs(data []byte) []byte {
	data = trimBOM(data)

	parameters := re.FindAllSubmatch(data, -1)

	for _, parameter := range parameters {
		if len(parameter) != 2 {
			l.Infof("%s", parameter)
			continue
		}
		envvar := string(parameter[1])
		l.Infof("envvar=%s", envvar)

		u, err := url.Parse(envvar)
		if err != nil {
			l.Errorf("%s can not parse to url", string(parameter[0]))
			return data
		}
		l.Infof("url scheme=%s host=<%s> path=<%s>", u.Scheme, u.Host, u.Path)
		encVal := ""
		switch u.Scheme {
		case "file":
			encVal = readFromFile(u.Host + u.Path)
			if encVal != "" {
				encVal = strings.TrimRight(encVal, "\n")
				l.Infof("ENC read from file,value=%s", maskPassword(encVal))
			}
		case "aes":
			l.Infof("aes decrypt key=%s text=%s", maskPassword(datakit.ConfigAESKey), u.Host+u.Path)
			encVal, err = AESDecrypt([]byte(datakit.ConfigAESKey), u.Host+u.Path)
			if err == nil {
				l.Infof("ENC from aes decrypt password= %s", maskPassword(encVal))
			} else {
				l.Errorf("aes decrypt err=%v", err)
			}
		default:
			l.Infof("unknown ENC scheme:%s,and enc=%s", u.Scheme, envvar)
		}

		if encVal != "" {
			data = bytes.Replace(data, parameter[0], []byte(encVal), 1)
		}
	}

	return data
}

func GetPipelinePath(category point.Category, pipeLineName string) (string, error) {
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
		files := plmanager.SearchWorkspaceScripts(datakit.GitReposRepoFullPath)
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

	files := plmanager.SearchWorkspaceScripts(datakit.PipelineDir)
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

// InitGitreposDir must exported because other modules, like pythond tests, would use it.
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

func fillPipelineConfig(c *Config) {
	if c.Pipeline != nil {
		c.Pipeline.EnableDebugFields = c.EnableDebugFields
	}
}
