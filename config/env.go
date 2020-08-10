package config

import (
	"os"
	"path/filepath"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func (c *Config) loadEnvs() error {

	dkid := os.Getenv("ENV_UUID")
	dwcfg := os.Getenv("ENV_DATAWAY")

	loglvl := os.Getenv("ENV_LOG_LEVEL")
	if loglvl != "" {
		c.MainCfg.LogLevel = loglvl
	}

	if dwcfg != "" {
		dw, err := ParseDataway(dwcfg)
		if err != nil {
			return err
		}

		c.MainCfg.DataWay = dw
	}

	if os.Getenv("ENV_WITHIN_DOCKER") != "" {
		c.withinDocker = true
	}

	dkhost := os.Getenv("ENV_HOSTNAME")
	if dkhost != "" {
		l.Debugf("set hostname to %s from ENV", dkhost)
		c.MainCfg.Hostname = dkhost
	} else {
		name, err := os.Hostname()
		if err != nil {
			l.Errorf("get hostname failed: %s", err.Error())
		} else {
			l.Debugf("set hostname to %s from os.Hostname()", dkhost)
			c.MainCfg.Hostname = name
		}
	}

	if c.withinDocker {
		maincfg := filepath.Join(datakit.InstallDir, "datakit.conf")
		if fi, err := os.Stat(maincfg); err != nil || fi.Size() == 0 { // create the main config

			l.Debugf("generating datakit.conf...")

			c.MainCfg.UUID = dkid
			if dkid == "" {
				c.MainCfg.UUID = cliutils.XID("dkid_")
			}

			fd, err := os.OpenFile(maincfg, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
			if err != nil {
				l.Errorf("failed to open %s: %s", maincfg, err)
				return err
			}

			defer fd.Close()

			tmp := template.New("")
			tmp, err = tmp.Parse(MainConfigTemplate)
			if err != nil {
				l.Errorf("failed to parse template: %s", err)
				return err
			}

			if err := tmp.Execute(fd, c.MainCfg); err != nil {
				l.Error(err)
				return err
			}
		}
	}

	return nil
}
