// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pythond contains pythond collector
package pythond

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	configSample = `
[[inputs.pythond]]

	# Python 采集器名称
	name = 'some-python-inputs'  # required

	# 运行 Python 采集器所需的环境变量
	#envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]

	# Python 采集器可执行程序路径(尽可能写绝对路径)
	cmd = "python3" # required. python3 is recommended.

	# 用户脚本的相对路径(填写文件夹，填好后该文件夹下一级目录的模块和 py 文件都将得到应用)
	dirs = []
`
)

var (
	inputName           = "pythond"
	l                   = logger.DefaultSLogger(inputName)
	onceReleasePrefiles sync.Once
	onceSetLog          sync.Once
)

type PythonDInput struct {
	Name string   `toml:"name"`
	Cmd  string   `toml:"cmd"`
	Dirs []string `toml:"dirs"`
	Envs []string `toml:"envs"`

	cmd *exec.Cmd

	semStop    *cliutils.Sem // start stop signal
	scriptName string
	scriptRoot string
}

func (*PythonDInput) Catalog() string { return inputName }

func (*PythonDInput) SampleConfig() string { return configSample }

func (*PythonDInput) AvailableArchs() []string { return datakit.AllOS }

func (*PythonDInput) SampleMeasurement() []inputs.Measurement { return []inputs.Measurement{} }

func getCliPyScript(scriptRoot, scriptName string) string {
	replacePair := map[string]string{
		"PythonCorePath":            "\"" + datakit.PythonCoreDir + "\"",
		"CustomerDefinedScriptRoot": scriptRoot,
		"CustomerDefinedScriptName": "\"" + scriptName + "\"",
	}

	return os.Expand(pyCli, func(k string) string { return replacePair[k] })
}

func (pe *PythonDInput) start() error {
	cli := getCliPyScript(pe.scriptRoot, pe.scriptName)

	pyTmpFle, err := ioutil.TempFile("", "pythond_")
	if err != nil {
		l.Errorf("ioutil.TempFile failed: %s", err.Error())
		return err
	}

	n, err := pyTmpFle.WriteString(cli)
	if err != nil {
		l.Errorf("TempFile.WriteString failed: %s", err.Error())
		return err
	}

	l.Debugf("python tmp = %s, written: %d", pyTmpFle.Name(), n)

	pe.cmd = exec.Command(pe.Cmd, pyTmpFle.Name(), fmt.Sprintf("--logname=%s", pe.Name)) //nolint:gosec
	if pe.Envs != nil {
		pe.cmd.Env = pe.Envs
	}

	stdout, err := pe.cmd.StdoutPipe()
	if err != nil {
		l.Errorf("cmd.StdoutPipe failed: %s", err.Error())
		return err
	}
	pe.cmd.Stderr = pe.cmd.Stdout

	l.Infof("starting cmd %s, envs: %+#v", pe.cmd.String(), pe.cmd.Env)
	if err := pe.cmd.Start(); err != nil {
		l.Errorf("start pythond input %s failed: %s", pe.Name, err.Error())
		return err
	}

	g := datakit.G("inputs_pythond")

	g.Go(func(ctx context.Context) error {
		l.Debug("go entry")
		tick := time.NewTicker(time.Second)
		defer tick.Stop()

		defer func() {
			if pyTmpFle != nil {
				if err := pyTmpFle.Close(); err != nil {
					l.Debugf("pyTmpFle.Close failed: %v", err)
				}
			}
		}()

		for {
			select {
			case <-tick.C:
				tmp := make([]byte, 1024)
				_, err := stdout.Read(tmp)
				l.Debug(string(tmp))
				if err != nil {
					l.Debugf("stdout.Read failed: %v", err)
					return nil
				}

			case <-datakit.Exit.Wait():
				return nil

			case <-pe.semStop.Wait():
				return nil
			}
		}
	})

	return nil
}

//------------------------------------------------------------------------------

func getPyModules(root string, ipd IPythond) []string {
	_, files, err := ipd.GetFolderList(root, 2)
	if err != nil {
		l.Error(err)
		return nil
	}
	return getFilteredPyModules(files, root)
}

func getFilteredPyModules(files []string, root string) []string {
	var arr []string

	for _, v := range files {
		base := filepath.Base(v)
		if base == "__init__.py" {
			continue
		}

		ext := filepath.Ext(base)
		if strings.ToLower(ext) != ".py" {
			continue
		}

		pureName := path.GetPureNameFromExt(base)

		parentFull := filepath.Dir(v)
		if parentFull == root {
			arr = append(arr, pureName)
		} else {
			parent := path.GetParentDirName(v)
			arr = append(arr, fmt.Sprintf("%s.%s", parent, pureName))
		}
	}

	return arr
}

//------------------------------------------------------------------------------

// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/509
func searchPythondDir(dirName string, enabledRepos []string, ipd IPythond) string {
	for _, v := range enabledRepos {
		destPath := filepath.Join(datakit.GitReposDir, v, datakit.GitRepoSubDirNamePythond, dirName)
		if ipd.IsDir(destPath) {
			return destPath
		}
	}
	return filepath.Join(datakit.PythonDDir, dirName)
}

//------------------------------------------------------------------------------

type IPythond interface {
	IsDir(path string) bool
	FileExist(filename string) bool
	GetFolderList(root string, deep int) (folders, files []string, err error)
	GitHasEnabled() bool
}

type pythondImpl struct{}

func (*pythondImpl) IsDir(ph string) bool {
	return path.IsDir(ph)
}

func (*pythondImpl) FileExist(ph string) bool {
	return datakit.FileExist(ph)
}

func (*pythondImpl) GetFolderList(root string, deep int) (folders, files []string, err error) {
	return path.GetFolderList(root, deep)
}

func (*pythondImpl) GitHasEnabled() bool {
	return config.GitHasEnabled()
}

//------------------------------------------------------------------------------

// Splicing Python related module information.
func getScriptNameRoot(dirs []string, ipd IPythond) (scriptName, scriptRoot string, err error) {
	var pyModules, modulesRoot []string
	for _, v := range dirs {
		var pythonPath string
		if ipd.GitHasEnabled() {
			// enabled git
			if filepath.IsAbs(v) {
				pythonPath = v
			} else {
				pythonPath = searchPythondDir(v, []string{datakit.GitReposRepoName}, ipd)
			}
		} else {
			// not enabled git
			pythonPath = filepath.Join(datakit.PythonDDir, v)
		}

		if ipd.IsDir(pythonPath) {
			pyModules = append(pyModules, getPyModules(pythonPath, ipd)...)
			modulesRoot = append(modulesRoot, pythonPath)
		} else if ipd.FileExist(pythonPath) {
			pyModules = append(pyModules, path.GetPureNameFromExt(pythonPath))
		}
	}

	pyModules = dkstring.GetUniqueArray(pyModules)
	modulesRoot = dkstring.GetUniqueArray(modulesRoot)

	if len(pyModules) == 0 || len(modulesRoot) == 0 {
		err = fmt.Errorf("pyModules or modulesRoot empty")
		return "", "", err
	}

	scriptName = strings.Join(pyModules, "\", \"")
	scriptRoot = "['" + strings.Join(modulesRoot, "', '") + "']"

	return scriptName, scriptRoot, nil
}

//------------------------------------------------------------------------------

func (pe *PythonDInput) Run() {
	setLog()
	l.Infof("starting pythond input %s...", pe.Name)

	onceReleasePrefiles.Do(func() {
		if err := ReleaseFiles(); err != nil {
			l.Errorf("pythond release prefiles failed: %v", err)
		}
	})

	// check
	pe.Name = dkstring.TrimString(pe.Name)
	if pe.Name == "" {
		l.Error("name should not be empty.")
		return
	}
	if len(pe.Dirs) == 0 {
		l.Error("dirs should not be empty.")
		return
	}

	var err error
	if pe.scriptName, pe.scriptRoot, err = getScriptNameRoot(pe.Dirs, &pythondImpl{}); err != nil {
		l.Error(err)
		return
	}

	l.Debugf("pe.scriptName = %v, pe.scriptRoot = %v", pe.scriptName, pe.scriptRoot)

	for {
		if err := pe.start(); err != nil { // start failed, retry
			time.Sleep(time.Second)
			continue
		}
		break
	}

	if err := pe.MonitProc(); err != nil { // blocking here...
		l.Errorf("datakit.MonitProc: %s", err.Error())
	}
}

func (pe *PythonDInput) MonitProc() error {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	if pe.cmd.Process == nil {
		return fmt.Errorf("invalid proc %s", pe.Name)
	}

	for {
		select {
		case <-tick.C:
			p, err := os.FindProcess(pe.cmd.Process.Pid)
			if err != nil {
				continue
			}

			switch runtime.GOOS {
			case datakit.OSWindows:

			default:
				if err := p.Signal(syscall.Signal(0)); err != nil {
					return err
				}
			}

		case <-datakit.Exit.Wait():
			if err := pe.stop(); err != nil { // XXX: should we wait here?
				return err
			}
			return nil

		case <-pe.semStop.Wait():
			if err := pe.stop(); err != nil { // XXX: should we wait here?
				return err
			}
			return nil
		}
	}
}

func (pe *PythonDInput) Terminate() {
	if pe.semStop != nil {
		pe.semStop.Close()
	}
}

func (pe *PythonDInput) stop() error {
	if err := pe.cmd.Process.Kill(); err != nil {
		l.Errorf("PythonDInput kill failed: %v", err)
		return err
	}
	return nil
}

func setLog() {
	onceSetLog.Do(func() {
		l = logger.SLogger(inputName)
	})
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &PythonDInput{
			semStop: cliutils.NewSem(),
		}
	})
}
