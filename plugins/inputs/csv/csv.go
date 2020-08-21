package csv

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type column struct {
	Index    int    `toml:"index" yaml:"index"`
	Name     string `toml:"name" yaml:"name"`                     // as tag/field name
	NaAction string `toml:"na_action" yaml:"na_action"`           // ignore/drop/abort
	Type     string `toml:"type,omitempty" yaml:"type,omitempty"` // value type: int/float/str

	AsTag   bool `toml:"as_tag,omitempty" yaml:"as_tag,omitempty"`
	AsField bool `toml:"as_field,omitempty" yaml:"as_field,omitempty"`
	AsTime  bool `tome:"as_time,omitempty" yaml:"as_time,omitempty"`

	TimeFormat    string `toml:"time_format,omitempty" yaml:"time_format,omitempty"`       // time format within csv
	TimePrecision string `toml:"time_precision,omitempty" yaml:"time_precision,omitempty"` // h/m/s/ms/us/ns
}

type CSV struct {
	PathEnv   string    `toml:"path_env" yaml:"-"`
	StartRows int       `toml:"start_rows" yaml:"start_rows"`
	File      string    `toml:"file" yaml:"file"`
	Metric    string    `toml:"metric" yaml:"metric"`
	Columns   []*column `toml:"columns" yaml:"columns"`
}

var (
	l         *logger.Logger
	inputName = "csv"
)

func (x *CSV) Catalog() string {
	return "csv"
}

func (x *CSV) SampleConfig() string {
	return configSample
}

func (x *CSV) Run() {
	l = logger.SLogger("csv")
	l.Info("csv started")

	b, err := yaml.Marshal(x)
	if err != nil {
		l.Error(err)
		return
	}

	encodeStr := base64.StdEncoding.EncodeToString(b)
	args := []string{
		filepath.Join(datakit.InstallDir, "externals", "csv", "main.py"),
		"-y", encodeStr,
	}
	cmd := exec.Command("python", args...)
	if x.PathEnv != "" {
		cmd.Env = []string{
			fmt.Sprintf("PATH=%s:$PATH", x.PathEnv),
		}

		l.Infof("set PATH to %s", cmd.Env[0])
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ch := make(chan interface{})
	go func() {
		if err := cmd.Run(); err != nil {
			l.Errorf("start csv failed: %s", err.Error())
		}
		close(ch)
	}()

	time.Sleep(time.Duration(2) * time.Second)
	l.Infof("csv PID: %d", cmd.Process.Pid)
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
				l.Errorf("signal 0 to %s failed: %s", "csv", err)
			}

		case <-datakit.Exit.Wait():
			if err := cmd.Process.Kill(); err != nil {
				l.Warnf("killing %s failed: %s, ignored", "csv", err)
			}
			l.Infof("killing %s (pid: %d) ok", "csv", cmd.Process.Pid)
			return

		case <-ch:
			l.Info("csvkit exit")
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &CSV{}
	})
}
