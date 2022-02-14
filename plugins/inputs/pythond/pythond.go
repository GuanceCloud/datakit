// Package pythond contains pythond collector
package pythond

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
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
`  // configSample

	pyCli = `
import os
import sys
import time
import importlib
import threading
import argparse
import logging
from logging.handlers import RotatingFileHandler

import sys
sys.path.append(${PythonCorePath})
sys.path.extend(${CustomerDefinedScriptRoot})

from datakit_framework import DataKitFramework

PY2 = sys.version_info[0] == 2
PY3 = sys.version_info[0] == 3

logger = logging.getLogger('pythond_cli')

def init_log():
    log_path = os.path.join(os.path.expanduser('~'), "_datakit_pythond_cli.log")
    print(log_path)
    logger.setLevel(logging.DEBUG)
    handler = RotatingFileHandler(log_path, maxBytes=100000, backupCount=10)
    logger.addHandler(handler)

def mylog(msg, *args, **kwargs):
    logger.debug(msg, *args, **kwargs)

class RunThread (threading.Thread):
	__plugin = DataKitFramework()
	__interval = 10

	def __init__(self, plugin):
		threading.Thread.__init__(self)
		self.__plugin = plugin
		if self.__plugin.interval:
			self.__interval = self.__plugin.interval

	def run(self):
		if self.__plugin:
			while True:
				try:
					self.__plugin.run()
				except:
					mylog("Unexpected error:", sys.exc_info()[0], self.__plugin.__name)
				time.sleep(self.__interval)

def search_plugin(plugin_path):
	try:
		mod = importlib.import_module(plugin_path)
	except ModuleNotFoundError:
		mylog(plugin_path + " not found.")
		return

	plugins = []

	for _, v in mod.__dict__.items():
		if v is not DataKitFramework and type(v).__name__ == 'type' and issubclass(v, DataKitFramework):
			plugin = v()
			# return plugin
			plugins.append(plugin)

	return plugins

def main(*args):
    plugins = []
    threads = []

    for arg in args:
        plg = search_plugin(arg)
        if plg and len(plg) > 0:
            plugins.extend(plg)

    for plg in plugins:
        thd = RunThread(plg)
        thd.start()
        threads.append(thd)

    for t in threads:
        t.join()

if __name__ == '__main__':
	parser = argparse.ArgumentParser(description="datakit framework")
	parser.add_argument('--logname', '-l', help='indicates datakit framework log name, required')
	args = parser.parse_args()
	if args.logname:
		DataKitFramework.log_name = args.logname
	else:
		print("need logname")
		sys.exit(-1)

	init_log()
	main(${CustomerDefinedScriptName})
`  // pyCli

) // const

var (
	inputName = "pythond"
	l         = logger.DefaultSLogger(inputName)
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

func (*PythonDInput) AvailableArchs() []string { return datakit.AllArch }

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

	l.Debugf("starting cmd %s, envs: %+#v", pe.cmd.String(), pe.cmd.Env)
	if err := pe.cmd.Start(); err != nil {
		l.Errorf("start pythond input %s failed: %s", pe.Name, err.Error())
		return err
	}

	go func() {
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
					return
				}

			case <-datakit.Exit.Wait():
				return

			case <-pe.semStop.Wait():
				return
			}
		}
	}()

	return nil
}

//------------------------------------------------------------------------------

type mockFolderList interface {
	GetFolderList(root string, deep int) (folders, files []string, err error)
}

type folderListMocker struct{}

func (*folderListMocker) GetFolderList(root string, deep int) (folders, files []string, err error) {
	return path.GetFolderList(root, deep)
}

func getPyModules(root string, mlist mockFolderList) []string {
	_, files, err := mlist.GetFolderList(root, 2)
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

type mockPath interface {
	IsDir(path string) bool
}

type pathMocker struct{}

func (*pathMocker) IsDir(ph string) bool {
	return path.IsDir(ph)
}

// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/509
func searchPythondDir(dirName string, enabledRepos []string, mp mockPath) string {
	for _, v := range enabledRepos {
		destPath := filepath.Join(datakit.GitReposDir, v, datakit.GitRepoSubDirNamePythond, dirName)
		if mp.IsDir(destPath) {
			return destPath
		}
	}
	return filepath.Join(datakit.PythonDDir, dirName)
}

//------------------------------------------------------------------------------

type mockPathEx interface {
	IsDir(path string) bool
	FileExist(filename string) bool
}

type pathMockExer struct{}

func (*pathMockExer) IsDir(ph string) bool {
	return path.IsDir(ph)
}

func (*pathMockExer) FileExist(ph string) bool {
	return datakit.FileExist(ph)
}

// Splicing Python related module information
func getScriptNameRoot(dirs []string, mp mockPath, mpEx mockPathEx, mlist mockFolderList) (scriptName, scriptRoot string, err error) {
	var pyModules, modulesRoot []string
	enabledRepos := config.GitEnabledRepoNames()
	for _, v := range dirs {
		var pythonPath string
		if len(enabledRepos) != 0 {
			// enabled git
			if filepath.IsAbs(v) {
				pythonPath = v
			} else {
				pythonPath = searchPythondDir(v, enabledRepos, mp)
			}
		} else {
			// not enabled git
			pythonPath = filepath.Join(datakit.PythonDDir, v)
		}

		if mpEx.IsDir(pythonPath) {
			pyModules = append(pyModules, getPyModules(pythonPath, mlist)...)
			modulesRoot = append(modulesRoot, pythonPath)
		} else if mpEx.FileExist(pythonPath) {
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
	l = logger.SLogger(inputName)

	l.Infof("starting pythond input %s...", pe.Name)

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
	if pe.scriptName, pe.scriptRoot, err = getScriptNameRoot(pe.Dirs, &pathMocker{}, &pathMockExer{}, &folderListMocker{}); err != nil {
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

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &PythonDInput{
			semStop: cliutils.NewSem(),
		}
	})
}
