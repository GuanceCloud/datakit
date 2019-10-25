package config

import (
	"bytes"
	"io/ioutil"
	"log"

	"github.com/alecthomas/template"
	yaml "gopkg.in/yaml.v2"
)

var (
	Cfg MainConfig

	MainCfgTemplate = `
log: {{.Log}}
log_level: {{.LogLevel}}
installdir: {{.InstallDir}}
`
)

type MainConfig struct {
	Log        string `yaml:"log,omitempty"`
	LogLevel   string `yaml:"log_level,omitempty"`
	InstallDir string `yaml:"installdir,omitempty"`

	Binlog *BinlogConfig `yaml:"binlog,omitempty"`
}

func (c *MainConfig) GenInitCfg(f string) string {

	t := template.New(``)

	t, err := t.Parse(MainCfgTemplate)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	buf := bytes.NewBufferString("")

	if err := t.Execute(buf, c); err != nil {
		log.Fatalf("%s", err.Error())
	}

	cfg := buf.String()

	cfg += "\nbinlog:\n"
	cfg += GenBinlogInitCfg()

	if err = ioutil.WriteFile(f, []byte(cfg), 0664); err != nil {
		log.Fatalf("%s", err.Error())
	}

	return cfg
}

func LoadConfig(f string) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		return err
	}

	if Cfg.Binlog != nil {
		if err := Cfg.Binlog.Check(); err != nil {
			return err
		}
	}

	return nil
}

func (c *MainConfig) Dump(f string) error {

	dt, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(f, dt, 0664)
}
