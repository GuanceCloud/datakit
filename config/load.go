// Loading datakit configures

package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	// envVarRe is a regex to find environment variables in the config file.
	envVarRe      = regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`)
	envVarEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
)

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

	l.Infof("main cfg: \n%s", c.String())

	// clear all samples before loading
	removeSamples()

	if err := initPluginSamples(); err != nil {
		return err
	}

	if err := initPluginPipeline(); err != nil {
		l.Errorf("init plugin pipeline: %s", err.Error())
		return err
	}

	l.Infof("init %d default plugins...", len(c.DefaultEnabledInputs))
	initDefaultEnabledPlugins(c)

	return ReloadInputConfig()
}

func GetConfRootPaths() []string {
	var confRootPath *[]string
	if GitHasEnabled() {
		confRootPath = &[]string{datakit.GitReposRepoFullPath}
	} else {
		confRootPath = &[]string{datakit.ConfdDir}
	}
	return *confRootPath
}

func ReloadInputConfig() error {
	confRootPath := GetConfRootPaths()

	var confTables []map[string]*ast.Table

	for _, v := range confRootPath {
		m := LoadInputsConfigEx(v)
		if len(m) != 0 {
			confTables = append(confTables, m)
		}
	}

	return ReloadInputTables(confTables)
}

func ReloadInputTables(confTables []map[string]*ast.Table) error {
	defer func() {
		if GitHasEnabled() {
			// #501 issue
			for _, name := range Cfg.DefaultEnabledInputs {
				if c, ok := inputs.Inputs[name]; ok {
					i := c()
					sample := i.SampleConfig()
					tpl, err := parseCfgBytes([]byte(sample))
					if err != nil {
						l.Errorf("parseCfgBytes failed: %s", err.Error())
						continue
					}
					res := getCheckInputCfgResult(tpl)
					if !res.Runnable() {
						l.Errorf("getCheckInputCfgResult failed: input_cfg_invalid")
						continue
					}
					for _, i := range res.AvailableInputs {
						inputs.AddInput(name, i)
					}
				}
			} // if
		} // if GitHasEnabled()
	}()

	if len(confTables) == 0 {
		return nil
	}

	inputs.Init()

	// reset inputs(for reloading)
	l.Debug("reset inputs")
	inputs.ResetInputs()

	for _, v := range confTables {
		for name, creator := range inputs.Inputs {
			if !datakit.Enabled(name) {
				l.Debugf("ignore unchecked input %s", name)
				continue
			}

			doLoadInputConf(name, creator, v)
		}
	}

	if !DisableSelfInput {
		inputs.AddSelf()
	}

	return nil
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

func ParseCfgFile(f string) (*ast.Table, error) {
	data, err := ioutil.ReadFile(filepath.Clean(f))
	if err != nil {
		l.Errorf("ioutil.ReadFile: %s", err.Error())
		return nil, fmt.Errorf("read config %s failed: %w", f, err)
	}

	return parseCfgBytes(data)
}

func parseCfgBytes(bys []byte) (*ast.Table, error) {
	data := feedEnvs(bys)

	tbl, err := toml.Parse(data)
	if err != nil {
		l.Errorf("parse toml %s failed", string(data))
		return nil, err
	}

	return tbl, nil
}

func ReloadCheckPipelineCfg(iputs []inputs.Input) (*tailer.Option, error) {
	for _, v := range iputs {
		if inp, ok := v.(inputs.PipelineInput); ok {
			opts := inp.GetPipeline()
			for _, vv := range opts {
				if vv.Pipeline == "" {
					continue
				}
				pFullPath, err := GetPipelinePath(vv.Pipeline)
				if err != nil {
					return nil, err
				}
				pl, err := pipeline.NewPipelineByScriptPath(pFullPath, false)
				if err != nil {
					return vv, err
				}
				if pl == nil {
					return vv, fmt.Errorf("pipeline_file_error")
				}
			}
		}
	}

	return nil, nil
}

func GetPipelinePath(pipeLineName string) (string, error) {
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
		plPath := filepath.Join(datakit.GitReposRepoFullPath, pipeLineName)
		if _, err := os.Stat(plPath); err != nil {
			l.Errorf("os.Stat failed: %v", err)
		} else {
			return plPath, nil // return once found the pipeline file
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
	switch namespace {
	case datakit.StrPipelineRemote:
		return path.GetSuffixFilesFromDirDeepOne(datakit.PipelineRemoteDir, datakit.StrPipelineFileSuffix)
	case datakit.StrGitRepos:
		return path.GetSuffixFilesFromDirDeepOne(datakit.GitReposRepoFullPath, datakit.StrPipelineFileSuffix)
	}
	return nil, fmt.Errorf("invalid namespace")
}

type CheckedInputCfgResult struct {
	Failed  int
	Unknown int
	Passed  int
	Ignored int

	AvailableInputs []inputs.Input
}

func (r *CheckedInputCfgResult) Runnable() bool {
	return r.Failed == 0
}

func ReloadCheckInputCfg() ([]inputs.Input, error) {
	var availableInputs []inputs.Input
	confRootPath := GetConfRootPaths()
	confSuffix := ".conf"

	for _, v := range confRootPath {
		iputs, err := CheckInputCfgEx(v, confSuffix)
		if err != nil {
			return nil, err
		}
		availableInputs = append(availableInputs, iputs...)
	}

	return availableInputs, nil
}

func CheckInputCfgEx(rootPath, suffix string) ([]inputs.Input, error) {
	var availableInputs []inputs.Input
	fps := SearchDir(rootPath, suffix)

	for _, fp := range fps {
		tpl, err := ParseCfgFile(fp)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", err, fp)
		} else {
			res := getCheckInputCfgResult(tpl)
			if !res.Runnable() {
				return nil, fmt.Errorf("input_cfg_invalid: %s", fp)
			}
			availableInputs = append(availableInputs, res.AvailableInputs...)
		}
	}

	return availableInputs, nil
}

func getCheckInputCfgResult(tpl *ast.Table) *CheckedInputCfgResult {
	res := CheckedInputCfgResult{}

	if len(tpl.Fields) == 0 {
		res.Failed++
		return &res
	}

	for field, node := range tpl.Fields {
		switch field {
		default:
			// not input
			res.Ignored++
			return &res

		case "inputs": //nolint:goconst
			stbl, ok := node.(*ast.Table)
			if !ok {
				// bad toml node
				res.Failed++
			} else {
				for inputName, v := range stbl.Fields {
					if c, ok := inputs.Inputs[inputName]; !ok {
						// unknown input
						res.Unknown++
					} else {
						iputs, err := TryUnmarshal(v, inputName, c)
						if err != nil {
							res.Failed++
							continue
						}
						res.Passed++
						res.AvailableInputs = append(res.AvailableInputs, iputs...)
					} // if c, ok := inputs.Inputs[inputName];
				} // for inputName, v := range stbl.Fields
			} // stbl, ok := node.(*ast.Table)
		} // switch field
	} // for field, node := range tpl.Fields

	return &res
}
