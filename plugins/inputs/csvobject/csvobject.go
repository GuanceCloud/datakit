package csvobject

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type CsvObject struct {
	PythonEnv string `toml:"pythonEnv"`
	File      string `toml:"file,omitempty"`
	StartRows int    `toml:"startRows,omitempty"`
	Interval  string `toml:"interval,omitempty"`
	Name      string `toml:"name,omitempty"`
	Class     string `toml:"class,omitempty"`
}

const (
	configSample = `
#[[inputs.csvobject]]
#  pythonEnv = "python3"
#  file      = "/path/your/csvfile.csv"
#  startRows = 0
#  interval  = "60s"
#  name      = "objectname"
#  class     = "objectclass"
`
	defaultInterval = "0s"
)

var (
	l         *logger.Logger
	inputName = "csvobject"
)

func (_ *CsvObject) Catalog() string {
	return inputName
}

func (_ *CsvObject) SampleConfig() string {
	return configSample
}

func (x *CsvObject) Run() {
	var encodeStr string
	var intVal int
	var startCmd = "python"
	l = logger.SLogger(inputName)
	logFile := inputName + ".log"

	if b, err := toml.Marshal(x); err != nil {
		l.Error(err)
		return
	} else {
		encodeStr = base64.StdEncoding.EncodeToString(b)
	}

	if x.Interval == "" {
		x.Interval = defaultInterval
	}

	if interval, err := time.ParseDuration(x.Interval); err != nil {
		l.Error(err)
		return
	} else {
		intVal = int(interval) / 1e9
	}

	if config.Cfg.HTTPListen == "" {
		l.Errorf("missed http_server_addr configuration in datakit.conf")
		return
	}

	port := strings.Split(config.Cfg.HTTPListen, ":")[1]
	args := []string{
		filepath.Join(datakit.InstallDir, "externals", "csv", "main.py"),
		"--object", encodeStr,
		"--interval", fmt.Sprintf("%d", intVal),
		"--http", "http://127.0.0.1:" + port,
		"--log_file", filepath.Join(datakit.InstallDir, "externals", logFile),
		"--log_level", config.Cfg.LogLevel,
	}

	if x.PythonEnv != "" {
		startCmd = x.PythonEnv
	}
	l.Info("csvobject started")
	cmd := exec.Command(startCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ch := make(chan interface{})
	go func() {
		if err := cmd.Run(); err != nil {
			l.Errorf("start csvobject failed: %s", err.Error())
		}
		close(ch)
	}()

	time.Sleep(time.Duration(2) * time.Second)
	l.Infof("csvobject PID: %d", cmd.Process.Pid)
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			p, err := os.FindProcess(cmd.Process.Pid)
			if err != nil {
				l.Error(err)
				continue
			}

			if err := p.Signal(syscall.Signal(0)); err != nil {
				l.Errorf("signal 0 to %s failed: %s", inputName, err)
			}

		case <-datakit.Exit.Wait():
			if err := cmd.Process.Kill(); err != nil {
				l.Warnf("killing %s failed: %s, ignored", inputName, err)
			}
			l.Infof("killing %s (pid: %d) ok", inputName, cmd.Process.Pid)
			return

		case <-ch:
			l.Info("csvobject exit")
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &CsvObject{}
	})
}
