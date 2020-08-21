package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func (c *Config) loadEnvs() error {

	enableInputs := os.Getenv("ENV_ENABLE_INPUTS")
	if enableInputs != "" {
		EnableInputs(enableInputs)
	}

	globalTags := os.Getenv("ENV_GLOBAL_TAGS")
	if globalTags != "" {
		c.MainCfg.GlobalTags = ParseGlobalTags(globalTags)
	}

	loglvl := os.Getenv("ENV_LOG_LEVEL")
	if loglvl != "" {
		c.MainCfg.LogLevel = loglvl
	}

	dwcfg := os.Getenv("ENV_DATAWAY")
	if dwcfg != "" {
		dw, err := ParseDataway(dwcfg)
		if err != nil {
			return err
		}

		c.MainCfg.DataWay = dw
	}

	dkhost := os.Getenv("ENV_HOSTNAME")
	if dkhost != "" {
		l.Debugf("set hostname to %s from ENV", dkhost)
		c.MainCfg.Hostname = dkhost
	} else {
		c.setHostname()
	}

	if datakit.Docker {
		maincfg := filepath.Join(datakit.InstallDir, "datakit.conf")
		if fi, err := os.Stat(maincfg); err != nil || fi.Size() == 0 { // create the main config

			l.Debugf("generating datakit.conf...")

			dkid := os.Getenv("ENV_UUID")
			c.MainCfg.UUID = dkid
			if dkid == "" {
				c.MainCfg.UUID = cliutils.XID("dkid_")
			}

			cfgdata, err := buildMainCfg(c.MainCfg)
			if err != nil {
				l.Errorf("failed to build main cfg %s", err)
				return err
			}

			if err := ioutil.WriteFile(maincfg, cfgdata, os.ModePerm); err != nil {
				l.Error(err)
				return err
			}
		}
	}

	return nil
}
