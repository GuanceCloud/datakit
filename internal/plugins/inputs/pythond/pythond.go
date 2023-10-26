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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	configSample = `
[[inputs.pythond]]
  # Python input name
  name = 'some-python-inputs'  # required

  # System environments to run Python
  #envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]

  # Python path(recomment abstract Python path)
  cmd = "python3" # required. python3 is recommended.

  # Python scripts relative path
  dirs = []`
)

var (
	inputName           = "pythond"
	l                   = logger.DefaultSLogger(inputName)
	onceReleasePrefiles sync.Once
	onceSetLog          sync.Once
)

type Input struct {
	Name string            `toml:"name"`
	Cmd  string            `toml:"cmd"`
	Dirs []string          `toml:"dirs"`
	Envs []string          `toml:"envs"`
	Tags map[string]string `toml:"tags"` // TODO

	cmd    *exec.Cmd
	feeder io.Feeder // TODO

	semStop    *cliutils.Sem // start stop signal
	scriptName string
	scriptRoot string
}

func (*Input) Catalog() string { return inputName }

func (*Input) SampleConfig() string { return configSample }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleMeasurement() []inputs.Measurement { return []inputs.Measurement{} }

func getCliPyScript(scriptRoot, scriptName string) string {
	replacePair := map[string]string{
		"PythonCorePath":            "\"" + datakit.PythonCoreDir + "\"",
		"CustomerDefinedScriptRoot": scriptRoot,
		"CustomerDefinedScriptName": "\"" + scriptName + "\"",
	}

	return os.Expand(pyCli, func(k string) string { return replacePair[k] })
}

func (ipt *Input) start() error {
	cli := getCliPyScript(ipt.scriptRoot, ipt.scriptName)

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

	ipt.cmd = exec.Command(ipt.Cmd, pyTmpFle.Name(), fmt.Sprintf("--logname=%s", ipt.Name)) //nolint:gosec
	if ipt.Envs != nil {
		ipt.cmd.Env = ipt.Envs
	}

	stdout, err := ipt.cmd.StdoutPipe()
	if err != nil {
		l.Errorf("cmd.StdoutPipe failed: %s", err.Error())
		return err
	}
	ipt.cmd.Stderr = ipt.cmd.Stdout

	l.Infof("starting cmd %s, envs: %+#v", ipt.cmd.String(), ipt.cmd.Env)
	if err := ipt.cmd.Start(); err != nil {
		l.Errorf("start pythond input %s failed: %s", ipt.Name, err.Error())
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

			case <-ipt.semStop.Wait():
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
	l.Infof("[11] files = %v", files)
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
		l.Infof("[12] pureName = %s", pureName)

		parentFull := filepath.Dir(v)
		if parentFull == root {
			arr = append(arr, pureName)
		} else {
			parent := path.GetParentDirName(v)
			l.Infof("[13] parent = %s", parent)
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
	_folders, _files, _err := path.GetFolderList(root, deep)
	if err != nil {
		return nil, nil, _err
	}
	// exclude filename contains ..
	for _, v := range _files {
		if !strings.Contains(v, "..") {
			files = append(files, v)
		}
	}
	return _folders, files, nil
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
				l.Infof(" [1] pythonPath = %s", pythonPath)
			} else {
				pythonPath = searchPythondDir(v, []string{datakit.GitReposRepoName}, ipd)
				l.Infof(" [2] pythonPath = %s", pythonPath)
			}
		} else {
			// not enabled git
			pythonPath = filepath.Join(datakit.PythonDDir, v)
			l.Infof(" [3] pythonPath = %s", pythonPath)
		}

		if ipd.IsDir(pythonPath) {
			pyModules = append(pyModules, getPyModules(pythonPath, ipd)...)
			modulesRoot = append(modulesRoot, pythonPath)
			l.Infof(" [4] pyModules   = %v", pyModules)
			l.Infof(" [5] modulesRoot = %v", modulesRoot)
		} else if ipd.FileExist(pythonPath) {
			pyModules = append(pyModules, path.GetPureNameFromExt(pythonPath))
			l.Infof(" [6] pyModules = %v", pyModules)
		}
	}

	pyModules = dkstring.GetUniqueArray(pyModules)
	modulesRoot = dkstring.GetUniqueArray(modulesRoot)
	l.Infof(" [7] pyModules   = %v", pyModules)
	l.Infof(" [8] modulesRoot = %v", modulesRoot)

	if len(pyModules) == 0 || len(modulesRoot) == 0 {
		err = fmt.Errorf("pyModules or modulesRoot empty")
		return "", "", err
	}

	scriptName = strings.Join(pyModules, "\", \"")
	scriptRoot = "['" + strings.Join(modulesRoot, "', '") + "']"
	l.Infof(" [9] scriptName = %s", scriptName)
	l.Infof("[10] scriptRoot = %s", scriptRoot)

	return scriptName, scriptRoot, nil
}

//------------------------------------------------------------------------------

func (ipt *Input) Run() {
	setLog()
	l.Infof("starting pythond input %s...", ipt.Name)

	onceReleasePrefiles.Do(func() {
		if err := ReleaseFiles(); err != nil {
			l.Errorf("pythond release prefiles failed: %v", err)
		}
	})

	// check
	ipt.Name = dkstring.TrimString(ipt.Name)
	if ipt.Name == "" {
		l.Error("name should not be empty.")
		return
	}
	if len(ipt.Dirs) == 0 {
		l.Error("dirs should not be empty.")
		return
	}

	var err error
	if ipt.scriptName, ipt.scriptRoot, err = getScriptNameRoot(ipt.Dirs, &pythondImpl{}); err != nil {
		l.Error(err)
		return
	}

	l.Debugf("pe.scriptName = %v, pe.scriptRoot = %v", ipt.scriptName, ipt.scriptRoot)

	for {
		if err := ipt.start(); err != nil { // start failed, retry
			time.Sleep(time.Second)
			continue
		}
		break
	}

	if err := ipt.MonitProc(); err != nil { // blocking here...
		l.Errorf("datakit.MonitProc: %s", err.Error())
	}
}

func (ipt *Input) MonitProc() error {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	if ipt.cmd.Process == nil {
		return fmt.Errorf("invalid proc %s", ipt.Name)
	}

	for {
		select {
		case <-tick.C:
			p, err := os.FindProcess(ipt.cmd.Process.Pid)
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
			if err := ipt.stop(); err != nil { // XXX: should we wait here?
				return err
			}
			return nil

		case <-ipt.semStop.Wait():
			if err := ipt.stop(); err != nil { // XXX: should we wait here?
				return err
			}
			return nil
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) stop() error {
	if err := ipt.cmd.Process.Kill(); err != nil {
		l.Errorf("Input kill failed: %v", err)
		return err
	}
	return nil
}

func setLog() {
	onceSetLog.Do(func() {
		l = logger.SLogger(inputName)
	})
}

func defaultInput() *Input {
	return &Input{
		feeder:  io.DefaultFeeder(),
		semStop: cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
