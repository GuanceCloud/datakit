package cmds

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

var (
	oldInstallDir      = "/usr/local/cloudcare/dataflux/datakit"
	oldInstallDirWin   = `C:\Program Files\dataflux\datakit`
	oldInstallDirWin32 = `C:\Program Files (x86)\dataflux\datakit`
)

const (
	datakitBin = "datakit"
)

func mvOldDatakit(svc service.Service) {
	olddir := oldInstallDir
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case datakit.OSArchWinAmd64:
		olddir = oldInstallDirWin
	case datakit.OSArchWin386:
		olddir = oldInstallDirWin32
	}

	if _, err := os.Stat(olddir); err != nil {
		l.Debugf("deprecated install path %s not exists, ignored", olddir)
		return
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service datakit failed: %s, ignored", err.Error())
	}

	if err := os.Rename(olddir, datakit.InstallDir); err != nil {
		l.Fatalf("move %s -> %s failed: %s", olddir, datakit.InstallDir, err.Error())
	}
}

func runInstaller() {

	if FlagInstallLog == "" {
		lopt := logger.OPT_DEFAULT | logger.OPT_STDOUT
		if runtime.GOOS != "windows" { // disable color on windows(some color not working under windows)
			lopt |= logger.OPT_COLOR
		}

		if err := logger.SetGlobalRootLogger("", logger.DEBUG, lopt); err != nil {
			l.Warnf("set root log failed: %s", err.Error())
		}
	} else {
		l.Infof("set log file to %s", FlagInstallLog)
		if err := logger.SetGlobalRootLogger(FlagInstallLog, logger.DEBUG, logger.OPT_DEFAULT); err != nil {
			l.Errorf("set root log failed: %s", err.Error())
		}
	}

	l = logger.SLogger("installer")

	dkservice.ServiceExecutable = filepath.Join(datakit.InstallDir, datakitBin)
	if runtime.GOOS == datakit.OSWindows {
		dkservice.ServiceExecutable += ".exe"
	}

	svc, err := dkservice.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	l.Info("stoping datakit...")
	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}

	if FlagProxy != "" {
		if _, err := url.Parse(FlagProxy); err != nil {
			l.Warnf("bad proxy config expect http://ip:port given %s", FlagProxy)
		}
	}

	// 迁移老版本 datakit 数据目录
	mvOldDatakit(svc)
	config.InitDirs()

	// create install dir if not exists
	if err := os.MkdirAll(datakit.InstallDir, 0775); err != nil {
		l.Fatal(err)
	}

	if FlagDKUpgrade { // upgrade new version
		l.Infof("Upgrading to version %s...", ReleaseVersion)
		if err := upgradeDatakit(svc); err != nil {
			l.Fatalf("upgrade datakit: %s, ignored", err.Error())
		}
	} else { // install new datakit
		l.Infof("Installing version %s...", ReleaseVersion)
		installNewDatakit(svc)
	}

	if FlagInstallOnly {
		l.Infof("starting service %s...", dkservice.ServiceName)
		if err = service.Control(svc, "start"); err != nil {
			l.Warnf("star service: %s, ignored", err.Error())
		}
	}

	config.CreateSymlinks()

	if FlagDKUpgrade { // upgrade new version
		l.Info(":) Upgrade Success!")
	} else {
		l.Info(":) Install Success!")
	}

	promptReferences()
}

func promptReferences() {
	fmt.Printf("\n\tVisit http://localhost:%d/man/changelog to see DataKit change logs.\n", FlagDatakitHTTPPort)
	fmt.Printf("\tVisit http://localhost:%d/monitor to see DataKit running status.\n", FlagDatakitHTTPPort)
	fmt.Printf("\tVisit http://localhost:%d/man to see DataKit manuals.\n\n", FlagDatakitHTTPPort)
}

func upgradeDatakit(svc service.Service) error {

	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}

	mc := config.Cfg

	if err := mc.LoadMainTOML(datakit.MainConfPath); err == nil {
		mc, _ = upgradeMainConfig(mc)

		if FlagOTA {
			l.Debugf("set auto update flag")
			mc.AutoUpdate = FlagOTA
		}

		writeDefInputToMainCfg(mc)
	} else {
		l.Warnf("load main config: %s, ignored", err.Error())
		return err
	}

	for _, dir := range []string{datakit.DataDir, datakit.ConfdDir} {
		if err := os.MkdirAll(dir, datakit.ConfPerm); err != nil {
			return err
		}
	}

	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install datakit service: %s, ignored", err.Error())
	}

	return nil
}

func installNewDatakit(svc service.Service) {

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service: %s, ignored", err.Error())
	}

	mc := config.Cfg

	// prepare dataway info
	mc.DataWay = getDataWayCfg()
	if FlagOTA {
		l.Debugf("set auto update flag")
		mc.AutoUpdate = FlagOTA
	}

	// accept any install options
	if FlagGlobalTags != "" {
		mc.GlobalTags = config.ParseGlobalTags(FlagGlobalTags)
	}

	mc.Namespace = FlagNamespace
	mc.HTTPAPI.Listen = fmt.Sprintf("%s:%d", FlagDatakitHTTPListen, FlagDatakitHTTPPort)
	mc.InstallDate = time.Now()

	if mc.DataWay != nil {
		mc.DataWay.HttpProxy = FlagProxy
	}

	if FlagDatakitName != "" {
		mc.Name = FlagDatakitName
	}

	writeDefInputToMainCfg(mc)

	l.Infof("installing service %s...", dkservice.ServiceName)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install service: %s, ignored", err.Error())
	}
}

var (
	defaultHostInputs          = []string{"cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes"}
	defaultHostInputsWithLinux = []string{"cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes", "container"}
)

func writeDefInputToMainCfg(mc *config.Config) {

	var hostInputs = defaultHostInputs
	if runtime.GOOS == datakit.OSLinux {
		hostInputs = defaultHostInputsWithLinux
	}

	if FlagEnableInputs == "" {
		FlagEnableInputs = strings.Join(hostInputs, ",")
	} else {
		FlagEnableInputs = FlagEnableInputs + "," + strings.Join(hostInputs, ",")
	}

	mc.EnableDefaultsInputs(FlagEnableInputs)

	switch FlagCloudProvider {
	case "aliyun", "tencent", "aws":
		if conf, err := preEnableHostobjectInput(FlagCloudProvider); err != nil {
			l.Fatalf("failed to init hostobject conf: %s", err.Error())
		} else {
			cfgpath := filepath.Join(datakit.ConfdDir, "host", "hostobject.conf")
			if err := ioutil.WriteFile(cfgpath, conf, datakit.ConfPerm); err != nil {
				l.Fatalf("failed to init hostobject conf: %s", err.Error())
			}
		}
	default:
		// pass
	}

	// build datakit main config
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}
}

func preEnableHostobjectInput(cloud string) ([]byte, error) {
	// I don't want to import hostobject input, cause the installer binary bigger
	sample := []byte(`
[inputs.hostobject]

#pipeline = '' # optional

## Datakit does not collect network virtual interfaces under the linux system.
## Setting enable_net_virtual_interfaces to true will collect network virtual interfaces stats for linux.
# enable_net_virtual_interfaces = true

## Ignore mount points by filesystem type. Default ignored following FS types
# ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "autofs", "squashfs", "aufs"]


[inputs.hostobject.tags] # (optional) custom tags
# cloud_provider = "aliyun" # aliyun/tencent/aws
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...`)

	conf := bytes.Replace(sample,
		[]byte(`# cloud_provider = "aliyun"`),
		[]byte(fmt.Sprintf(`  cloud_provider = "%s"`, cloud)),
		-1)

	return conf, nil
}

func upgradeMainConfig(c *config.Config) (*config.Config, error) {

	if c.DataWay != nil {
		c.DataWay.DeprecatedURL = ""
	}

	// XXX: 无脑更改日志位置
	switch runtime.GOOS {
	case datakit.OSWindows:
		c.Logging.Log = filepath.Join(datakit.InstallDir, "log")
		c.Logging.GinLog = filepath.Join(datakit.InstallDir, "gin.log")
	default:
		c.Logging.Log = "/var/log/datakit/log"
		c.Logging.GinLog = "/var/log/datakit/gin.log"
	}
	l.Debugf("set log to %s, remove ", c.Logging.Log)
	l.Debugf("set gin log to %s", c.Logging.GinLog)

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

	if c.HTTPListenDeprecated != "" {
		c.HTTPAPI.Listen = c.HTTPListenDeprecated
		c.HTTPListenDeprecated = ""
	}

	if c.Disable404PageDeprecated {
		c.HTTPAPI.Disable404Page = true
		c.Disable404PageDeprecated = false
	}

	if c.IOCacheCountDeprecated != 0 {
		c.IOConf.MaxCacheCount = c.IOCacheCountDeprecated
		c.IOCacheCountDeprecated = 0
	}

	if c.OutputFileDeprecated != "" {
		c.IOConf.OutputFile = c.OutputFileDeprecated
		c.OutputFileDeprecated = ""
	}

	if c.DataWay != nil {
		c.DataWay.HttpProxy = FlagProxy
	}

	return c, nil
}

func getDataWayCfg() *dataway.DataWayCfg {
	dw := &dataway.DataWayCfg{}

	if FlagDataway != "" {
		dw.URLs = strings.Split(FlagDataway, ",")
		if err := dw.Apply(); err != nil {
			l.Fatal(err)
		}
	} else {
		l.Fatal("should not been here")
	}

	return dw
}
