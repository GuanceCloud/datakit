// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	"os"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

var l = logger.DefaultSLogger("upgrade")

func SetLog() {
	l = logger.SLogger("upgrade")
}

func Upgrade() error {
	mc := config.Cfg

	// load exists datakit.conf
	if err := mc.LoadMainTOML(datakit.MainConfPath); err == nil {
		// load DK_XXX env config
		mc = loadDKEnvCfg(mc)

		mc = upgradeMainConfig(mc)

		writeDefInputToMainCfg(mc, true)
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
	if c.PointPool != nil {
		l.Infof("default enable point pool")
		c.PointPool.Enable = false // default disable point-pool
	}

	// setup dataway
	if c.Dataway != nil {
		c.Dataway.DeprecatedURL = ""
		c.Dataway.HTTPProxy = Proxy

		if c.Dataway.ContentEncoding == "v1" {
			l.Infof("switch default content-encoding from v1 to v2")
			c.Dataway.ContentEncoding = "v2"
		}

		if c.Dataway.DeprecatedHTTPTimeout != "" {
			du, err := time.ParseDuration(c.Dataway.DeprecatedHTTPTimeout)
			if err == nil {
				c.Dataway.HTTPTimeout = du
			}

			c.Dataway.DeprecatedHTTPTimeout = "" // always remove the config
		}

		if c.Dataway.MaxRawBodySize >= dataway.DeprecatedDefaultMaxRawBodySize {
			l.Infof("to save memory, set max-raw-body-size from %d to %d",
				c.Dataway.MaxRawBodySize, dataway.DefaultMaxRawBodySize)

			c.Dataway.MaxRawBodySize = dataway.DefaultMaxRawBodySize
		}
	}

	l.Infof("Set log to %s", c.Logging.Log)
	l.Infof("Set gin log to %s", c.Logging.GinLog)

	// upgrade logging settings
	if c.LogDeprecated != "" {
		c.Logging.Log = c.LogDeprecated
		c.LogDeprecated = ""
	}

	if !c.EnablePProf { // enable pprof by default
		c.EnablePProf = true
	}

	if c.PProfListen == "" {
		c.PProfListen = "localhost:6060"
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

	if c.IntervalDeprecated != time.Duration(0) {
		c.IO.CompactInterval = c.IntervalDeprecated
		c.IntervalDeprecated = time.Duration(0)
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

	if c.ResourceLimitOptionsDeprecated != nil {
		c.ResourceLimitOptions = c.ResourceLimitOptionsDeprecated
		c.ResourceLimitOptionsDeprecated = nil
	}

	c.InstallVer = DataKitVersion

	return c
}
