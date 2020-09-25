package csvmetric

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"strings"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type TimeStamp struct {
	Column     string  `toml:"column"`
	TimeFormat string  `toml:"timeFormat,omitempty"`
	Precision  string  `toml:"precision,omitempty"`
}

type MetricField struct {
	Column     string  `toml:"column,omitempty"`
	NullOp     string  `toml:"nullOp,omitempty"`
	NullFill   string  `toml:"nullFill,omitempty"`
	Type       string  `toml:"type,omitempty"`
}
type CsvMetric struct {
	File      string         `toml:"file"`
	StartRows int            `toml:"startRows"`
	Metric    string         `toml:"metric"`
	Tags      []string       `toml:"tags,omitempty"`
	Timestamp TimeStamp      `toml:"timestamp,omitempty"`
	Field     []MetricField  `toml:"field,omitempty"`

}

const (
	configSample = `
#[[inputs.csvmetric]]
#  file      = "/path/your/csvfile.csv"
#  startRows = 0 
#  metric    = "metric-name"
#  tags      = ["column-name1","column-name2","column-name3"] 
#  [inputs.csvmetric.timestamp]
#    column     = "column"
#    timeFormat = "15/08/27 10:20:06"
#    precision  = "ns"
#
#  [[inputs.csvmetric.field]]
#    column     = "column-name3"
#    nullOp     = "ignore"
#    nullFill   = "default-value"
#    type       = "int"
#  [[inputs.csvmetric.field]]
#    column     = "column-name4"
#    nullOp     = "drop"
#    nullFill   = "default-value"
#    type       = "str"
`
)
var (
	l         *logger.Logger
	inputName = "csvmetric"
)

func (_ *CsvMetric) Catalog() string {
	return inputName
}

func (_ *CsvMetric) SampleConfig() string {
	return configSample
}

func (x *CsvMetric) Run() {
	var encodeStr string
	l = logger.SLogger(inputName)
	logFile := inputName + ".log"

	if b, err := toml.Marshal(x); err != nil {
		l.Error(err)
		return
	} else {
		encodeStr = base64.StdEncoding.EncodeToString(b)
	}

	if datakit.Cfg.MainCfg.HTTPBind == "" {
		l.Errorf("missed http_server_addr configuration in datakit.conf")
		return
	}

	port := strings.Split(datakit.Cfg.MainCfg.HTTPBind, ":")[1]
	args := []string{
		filepath.Join(datakit.InstallDir, "externals", "csv", "main.py"),
		"--metric", encodeStr,
		"--http", "http://127.0.0.1:" + port,
		"--log_file", filepath.Join(datakit.InstallDir, "externals", logFile),
		"--log_level", datakit.Cfg.MainCfg.LogLevel,
	}

	l.Info("csvmetric started")
	cmd := exec.Command("python3", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ch := make(chan interface{})
	go func() {
		if err := cmd.Run(); err != nil {
			l.Errorf("start csvmetric failed: %s", err.Error())
		}
		close(ch)
	}()

	time.Sleep(time.Duration(2) * time.Second)
	l.Infof("csvmetric PID: %d", cmd.Process.Pid)
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
			l.Info("csvmetric exit")
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &CsvMetric{}
	})
}
