// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
)

var l = logger.DefaultSLogger("upgrade")

func SetLog() {
	l = logger.SLogger("upgrade")
}

func CheckVersion(s string) error {
	v := version.VerInfo{VersionString: s}
	if err := v.Parse(); err != nil {
		return err
	}

	// 对 1.1.x 版本的 datakit，此处暂且认为是 stable 版本，不然
	// 无法从 1.1.x 升级到 1.2.x
	// 1.2 以后的版本（1.3/1.5/...）等均视为 unstable 版本
	if v.GetMinor() == 1 {
		return nil
	}

	if !v.IsStable() {
		if EnableExperimental != 0 {
			cp.Warnf("upgrade version is unstable\n")
		} else {
			return fmt.Errorf("upgrade to %s is not stable version, use env: <$DK_ENABLE_EXPEIMENTAL=1> to upgrade", s)
		}
	}
	return nil
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
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	for _, dir := range []string{datakit.DataDir, datakit.ConfdDir} {
		if err := os.MkdirAll(dir, datakit.ConfPerm); err != nil {
			return err
		}
	}

	installExternals := map[string]struct{}{}
	for _, v := range strings.Split(InstallExternals, ",") {
		installExternals[v] = struct{}{}
	}

	updateEBPF := false
	if runtime.GOOS == datakit.OSLinux && (runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
		if _, err := os.Stat(filepath.Join(datakit.InstallDir, "externals", "datakit-ebpf")); err == nil {
			updateEBPF = true
		}
		if _, ok := installExternals["ebpf"]; ok {
			updateEBPF = true
		}
	}

	if updateEBPF {
		cp.Infof("upgrade DataKit eBPF plugin...\n")
		// nolint:gosec
		cmd := exec.Command(filepath.Join(datakit.InstallDir, "datakit"), "install", "--ebpf")
		if msg, err := cmd.CombinedOutput(); err != nil {
			l.Warnf("upgrade external input plugin %s failed: %s msg: %s", "ebpf", err.Error(), msg)
		}
	}

	return nil
}

func upgradeMainConfig(c *config.Config) *config.Config {
	// setup dataway
	if c.DataWayCfg != nil {
		c.DataWayCfg.DeprecatedURL = ""
		c.DataWayCfg.HTTPProxy = Proxy
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
		c.IOConf.MaxCacheCount = c.IOCacheCountDeprecated
		c.IOCacheCountDeprecated = 0
	}

	if c.IOConf.MaxCacheCount < 1000 {
		c.IOConf.MaxCacheCount = 1000
	}

	if c.OutputFileDeprecated != "" {
		c.IOConf.OutputFile = c.OutputFileDeprecated
		c.OutputFileDeprecated = ""
	}

	if c.IntervalDeprecated != "" {
		c.IOConf.FlushInterval = c.IntervalDeprecated
		c.IntervalDeprecated = ""
	}

	if c.IOConf.FeedChanSize > 1 {
		c.IOConf.FeedChanSize = 1 // reset to 1
	}

	if c.IOConf.MaxDynamicCacheCountDeprecated > 0 {
		c.IOConf.MaxDynamicCacheCountDeprecated = 0 // clear the config
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

	return c
}
