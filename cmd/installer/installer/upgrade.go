// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	"os"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

var l = logger.DefaultSLogger("upgrade")

func SetLog() {
	l = logger.SLogger("upgrade")
}

func (args *InstallerArgs) Upgrade(mc *config.Config, svc service.Service) (err error) {
	if args.shouldReinstallService {
		l.Info("service updated, uninstall it...")

		args.uninstallDKService(svc)

		l.Infof("re-installing service datakit...")
		if err := service.Control(svc, "install"); err != nil {
			l.Warnf("uninstall service failed %s", err.Error()) //nolint:lll
		}
	}

	// load exists datakit.conf
	if err := mc.LoadMainTOML(datakit.MainConfPath); err == nil {
		// load DK_XXX env config
		mc, err = args.LoadInstallerArgs(mc)
		if err != nil {
			return err
		}

		mc = args.upgradeMainConfig(mc)
		if err := args.WriteDefInputs(mc); err != nil {
			return err
		}
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

func (args *InstallerArgs) upgradeMainConfig(c *config.Config) *config.Config {
	if c.PointPool != nil {
		l.Infof("always disable point pool by default")
		c.PointPool.Enable = false // default disable point-pool
	}

	// setup dataway
	if c.Dataway != nil {
		c.Dataway.DeprecatedURL = ""
		c.Dataway.HTTPProxy = args.Proxy

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

	if c.HTTPAPI.RequestRateLimit == 20.0 {
		c.HTTPAPI.RequestRateLimit = 100.0
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

	if c.ResourceLimitOptions != nil {
		// During upgrading, people has set limit-cpu-max in old version, so
		// disable limit-cpu-cores.
		//
		// To override old limit-cpu-max, we have to set limit-cpu-max during
		// installing or upgrading.
		if c.ResourceLimitOptions.CPUMax > 0 {
			c.ResourceLimitOptions.CPUCores = 0
		}
	}

	c.InstallVer = args.DataKitVersion
	if javaHome := getJavaHome(); javaHome != "" {
		if c.RemoteJob == nil {
			c.RemoteJob = &io.RemoteJob{}
		}
		if c.RemoteJob.JavaHome != "" {
			c.RemoteJob.JavaHome = javaHome
		}
	}

	return c
}

func getJavaHome() string {
	return os.Getenv("JAVA_HOME")
}
