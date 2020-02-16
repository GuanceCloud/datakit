package config

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/influxdata/toml"
)

const ()

var (
	ErrNoTelegrafConf = errors.New("no telegraf config")
)

func LoadTelegrafConfigs(ctx context.Context, cfgdir string) error {

	for index, name := range SupportsTelegrafMetraicNames {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		cfgpath := filepath.Join(cfgdir, name, fmt.Sprintf(`%s.conf`, name))
		err := CheckTelegrafCfgFile(cfgpath)

		if err == nil {
			MetricsEnablesFlags[index] = true
		} else {
			MetricsEnablesFlags[index] = false
			if err != ErrNoTelegrafConf {
				return fmt.Errorf("Error loading config file %s, %s", cfgpath, err)
			}
		}

	}
	return nil
}

func CheckTelegrafCfgFile(f string) error {

	_, err := os.Stat(f)

	if err != nil {
		return ErrNoTelegrafConf
	}

	cfgdata, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	tbl, err := toml.Parse(cfgdata)
	if err != nil {
		return err
	}

	if len(tbl.Fields) == 0 {
		return ErrNoTelegrafConf
	}

	if _, ok := tbl.Fields[`inputs`]; !ok {
		return errors.New("no inputs found")
	}

	return nil
}
