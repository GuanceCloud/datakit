package config

import (
	"os"
	"path/filepath"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func (c *Config) loadEnvs() error {

	c.MainCfg.UUID = os.Getenv("ENV_UUID")
	dwcfg := os.Getenv("ENV_DATAWAY")

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

	if c.withinDocker {
		maincfg := filepath.Join(datakit.InstallDir, "datakit.conf")
		if _, err := os.Stat(maincfg); err != nil { // create the main config

			l.Debugf("generating datakit.conf...")

			if c.MainCfg.UUID == "" {
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
