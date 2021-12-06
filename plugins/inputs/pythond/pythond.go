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
    if PY3:
        try:
            mod = importlib.import_module(plugin_path)
        except ModuleNotFoundError:
            mylog(plugin_path + " not found.")
            return
    elif PY2:
        try:
            mod = importlib.import_module(plugin_path)
        except ImportError:
            mylog(plugin_path + " not found.")
            return

    plugins = []

    for _, v in mod.__dict__.items():
        if v is not DataKitFramework and type(v).__name__ == 'type' and issubclass(v, DataKitFramework):
            plugin = v()
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

	cmd      *exec.Cmd
	duration time.Duration

	semStop     *cliutils.Sem // start stop signal
	replacePair map[string]string
	scriptName  string
	scriptRoot  string
}

func (*PythonDInput) Catalog() string { return inputName }

func (*PythonDInput) SampleConfig() string { return configSample }

func (*PythonDInput) AvailableArchs() []string { return datakit.AllArch }

func (*PythonDInput) SampleMeasurement() []inputs.Measurement { return []inputs.Measurement{} }

func (pe *PythonDInput) start() error {
	pe.replacePair = map[string]string{
		"PythonCorePath":                "\"" + datakit.PythonCoreDir + "\"",
		"CustomerDefinedScriptRoot":     pe.scriptRoot,
		"CustomerDefinedScriptName":     "\"" + pe.scriptName + "\"",
		"CustomerDefinedScriptInterval": fmt.Sprintf("%.1f", pe.duration.Seconds()),
	}

	cli := os.Expand(pyCli, func(k string) string { return pe.replacePair[k] })
	l.Debugf("cli = %s", cli)

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

func getPyModules(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return []string{}
	}

	var arr []string

	for _, f := range files {
		if f.IsDir() {
			// read file names
			ff, err := ioutil.ReadDir(filepath.Join(dir, f.Name()))
			if err != nil {
				continue
			}
			for _, v := range ff {
				if v.IsDir() {
					continue
				}
				if v.Name() == "__init__.py" {
					continue
				}
				ext := filepath.Ext(v.Name())
				if strings.ToLower(ext) != ".py" {
					continue
				}
				arr = append(arr, fmt.Sprintf("%s.%s", f.Name(),
					path.GetPureNameFromExt(v.Name())))
			}
		} else {
			if f.Name() == "__init__.py" {
				continue
			}
			ext := filepath.Ext(f.Name())
			if strings.ToLower(ext) == ".py" {
				arr = append(arr, path.GetPureNameFromExt(f.Name()))
			}
		}
	}

	return arr
}

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

	var pyModules, modulesRoot []string
	for _, v := range pe.Dirs {
		if config.GitHasEnabled() {
			v = filepath.Join(datakit.GitReposDir, v)
		} else {
			v = filepath.Join(datakit.PythonDDir, v)
		}

		// l.Debugf("v = %s", v)

		if path.IsDir(v) {
			pyModules = append(pyModules, getPyModules(v)...)
			modulesRoot = append(modulesRoot, v)
		} else if datakit.FileExist(v) {
			pyModules = append(pyModules, path.GetPureNameFromExt(v))
		}
	}

	pyModules = dkstring.GetUniqueArray(pyModules)
	modulesRoot = dkstring.GetUniqueArray(modulesRoot)

	// l.Debugf("pyModules = %v, modulesRoot = %v", pyModules, modulesRoot)

	pe.scriptName = strings.Join(pyModules, "\", \"")
	pe.scriptRoot = "['" + strings.Join(modulesRoot, "', '") + "']"

	// l.Debugf("pe.scriptName = %v, pe.scriptRoot = %v", pe.scriptName, pe.scriptRoot)

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
