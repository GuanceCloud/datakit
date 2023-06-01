// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	"os"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
)

var l = logger.DefaultSLogger("upgrade")

func SetLog() {
	l = logger.SLogger("upgrade")
}

func Upgrade() error {
	mc := config.Cfg

	// load exists datakit.conf
	if err := mc.LoadMainTOML(datakit.MainConfPath); err == nil {
		mc = upgradeMainConfig(mc)

		if OTA {
			l.Debugf("set auto update(OTA enabled)")
			mc.AutoUpdate = OTA
		}

		writeDefInputToMainCfg(mc)
	} else {
		l.Warnf("load main config: %s, ignored", err.Error())
		return err
	}

	// build datakit main config
	if err := mc.TryUpgradeCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	for _, dir := range []string{datakit.DataDir, datakit.ConfdDir} {
		if err := os.MkdirAll(dir, datakit.ConfPerm); err != nil {
			return err
		}
	}

	return nil
}

func upgradeMainConfig(c *config.Config) *config.Config {
	// setup dataway
	if c.Dataway != nil {
		c.Dataway.DeprecatedURL = ""
		c.Dataway.HTTPProxy = Proxy

		if c.Dataway.DeprecatedHTTPTimeout != "" {
			du, err := time.ParseDuration(c.Dataway.DeprecatedHTTPTimeout)
			if err == nil {
				c.Dataway.HTTPTimeout = du
			}

			c.Dataway.DeprecatedHTTPTimeout = "" // always remove the config
		}
	}

	cp.Infof("Set log to %s\n", c.Logging.Log)
	cp.Infof("Set gin log to %s\n", c.Logging.GinLog)

	// upgrade logging settings
	if c.LogDeprecated != "" {
		c.Logging.Log = c.LogDeprecated
		c.LogDeprecated = ""
	}

	if c.LogLevelDeprecated != "" {
		c.Logging.Level = c.LogLevelDeprecated
		c.LogLevelDeprecated = ""
	}

	if c.LogRotateDeprecated != 0 {
		c.Logging.Rotate = c.LogRotateDeprecated
		c.LogRotateDeprecated = 0
	}

	if c.GinLogDeprecated != "" {
		c.Logging.GinLog = c.GinLogDeprecated
		c.GinLogDeprecated = ""
	}

	// upgrade HTTP settings
	if c.HTTPListenDeprecated != "" {
		c.HTTPAPI.Listen = c.HTTPListenDeprecated
		c.HTTPListenDeprecated = ""
	}

	if c.Disable404PageDeprecated {
		c.HTTPAPI.Disable404Page = true
		c.Disable404PageDeprecated = false
	}

	// upgrade IO settings
	if c.IOCacheCountDeprecated != 0 {
		c.IO.MaxCacheCount = c.IOCacheCountDeprecated
		c.IOCacheCountDeprecated = 0
	}

	if c.IO.MaxCacheCount < 1000 {
		c.IO.MaxCacheCount = 1000
	}

	if c.OutputFileDeprecated != "" {
		c.IO.OutputFile = c.OutputFileDeprecated
		c.OutputFileDeprecated = ""
	}

	if c.IntervalDeprecated != "" {
		c.IO.FlushInterval = c.IntervalDeprecated
		c.IntervalDeprecated = ""
	}

	if c.IO.FeedChanSize > 1 {
		c.IO.FeedChanSize = 1 // reset to 1
	}

	if c.IO.MaxDynamicCacheCountDeprecated > 0 {
		c.IO.MaxDynamicCacheCountDeprecated = 0 // clear the config
	}

	// upgrade election settings
	if c.ElectionNamespaceDeprecated != "" {
		c.Election.Namespace = c.ElectionNamespaceDeprecated
		c.ElectionNamespaceDeprecated = ""
	}

	if c.NamespaceDeprecated != "" {
		c.Election.Namespace = c.NamespaceDeprecated
		c.NamespaceDeprecated = ""
	}

	if c.GlobalEnvTagsDeprecated != nil {
		c.Election.Tags = c.GlobalEnvTagsDeprecated
		c.GlobalEnvTagsDeprecated = nil
	}

	if c.EnableElectionDeprecated {
		c.Election.Enable = true
		c.EnableElectionDeprecated = false
	}

	if c.EnableElectionTagDeprecated {
		c.Election.EnableNamespaceTag = true
		c.EnableElectionTagDeprecated = false
	}

	c.InstallVer = DataKitVersion

	// move sinkers under dataway
	if c.SinkersDeprecated != nil && len(c.SinkersDeprecated.Arr) > 0 {
		for _, x := range c.SinkersDeprecated.Arr {
			if x.URL != "" && len(x.Categories) > 0 { // make sure it's a valid(at lease seems like) sinker
				c.Dataway.Sinkers = append(c.Dataway.Sinkers, x)
			}
		}

		c.SinkersDeprecated = nil // clear
	}

	return c
}
