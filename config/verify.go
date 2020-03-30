package config

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/influxdata/toml"
)

var (
	ErrConfigNotFound = errors.New("config file not found")
	ErrEmptyInput     = errors.New("no input is configured")
)

func VerifyToml(file string, checkInput bool) error {

	_, err := os.Stat(file)

	if err != nil {
		if os.IsNotExist(err) {
			return ErrConfigNotFound
		}
		return err
	}

	cfgdata, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	tbl, err := toml.Parse(cfgdata)
	if err != nil {
		return err
	}

	//没有任何配置项目，可能都被注释掉了
	if len(tbl.Fields) == 0 {
		return ErrConfigNotFound
	}

	if checkInput {
		if _, ok := tbl.Fields[`inputs`]; !ok {
			return ErrEmptyInput
		}
	}

	return nil
}
