package csv

import (
	"io/ioutil"
	"os"

	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
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

type rule struct {
	Metric  string    `toml:"metric" yaml:"metric"`
	Columns []*column `toml:"columns" yaml:"columns"`
}

type csv struct {
	PathEnv   string   `toml:"path_env" yaml:"-"`
	StartRows int      `toml:"start_rows" yaml:"start_rows"`
	Files     []string `toml:"files" yaml:"files"`
	Rules     []*rule  `toml:"rules" yaml:"rules"`
}

func (x *csv) Catalog() string {
	return "csv"
}

var (
	l *zap.SugaredLogger
)

func (x *csv) SampleConfig() string {
	return configSample
}

func (x *csv) Run() {

	l = logger.SLogger("csv")
	l.Info("csvkit started")

	l.Info("starting external csvkit...")

	csvConf := filepath.Join(config.InstallDir, "externals", "csv", "config.yaml")
	b, err = yaml.Marshal(x)
	if err := ioutil.WriteFile(csvConf, b, os.ModePerm); err != nil {
		l.Errorf("create csv config %s failed: %s", csvConf, err.Error())
	}

	args := []string{
		filepath.Join(config.InstallDir, "externals", "csv", "main.py"),
		"-y", csvConf,
	}
	cmd := exec.Command("python", args...)
	if x.PathEnv != "" {
		cmd.Env = []string{
			fmt.Sprintf("PATH=%s:$PATH"),
		}

		l.Infof("set PATH to %s", cmd.Env[0])
	}

	if err := cmd.Start(); err != nil {
		l.Errorf("start csv failed: %s", err.Error())
		return
	}

	proc := cmd.Process
	l.Infof("csv PID: %d", cmd.Process.Pid)
	datakit.MonitProc(cmd.Process, "csv")
}

func init() {
	inputs.Add("csv", func() inputs.Input {
		return &csv{}
	})
}
