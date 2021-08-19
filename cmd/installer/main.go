package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

var (
	oldInstallDir      = "/usr/local/cloudcare/dataflux/datakit"
	oldInstallDirWin   = `C:\Program Files\dataflux\datakit`
	oldInstallDirWin32 = `C:\Program Files (x86)\dataflux\datakit`

	DataKitBaseURL = ""
	DataKitVersion = ""
	dataUrl        = "https://" + path.Join(DataKitBaseURL, "data.tar.gz")
	datakitUrl     = "https://" + path.Join(DataKitBaseURL,
		fmt.Sprintf("datakit-%s-%s-%s.tar.gz",
			runtime.GOOS,
			runtime.GOARCH,
			DataKitVersion))

	flagDKUpgrade,
	flagInstallOnly,
	flagOffline, // deprecated
	flagDownloadOnly, // deprecated
	flagInfo,
	flagOTA bool
	flagDataway,
	flagEnableInputs,
	flagDatakitName,
	flagGlobalTags,
	flagProxy,
	flagDatakitHTTPListen,
	flagNamespace,
	flagInstallLog,
	flagCloudProvider string
	flagDatakitHTTPPort int

	l = logger.DefaultSLogger("installer")
)

const (
	datakitBin = "datakit"
)

func init() {
	flag.BoolVar(&flagDKUpgrade, "upgrade", false, ``)
	flag.BoolVar(&flagInstallOnly, "install-only", false, "install only                                                                                                                                                                                                                                                                                                                     , not start")
	flag.BoolVar(&flagOTA, "ota", false, "auto update")
	flag.StringVar(&flagDataway, "dataway", "", `address of dataway                                                                                                                                                                          ( http://IP:Port?token                                                                                                        = xxx) , port default 9528`)
	flag.StringVar(&flagEnableInputs, "enable-inputs", "", `default enable inputs                                                                                                                                                                 ( comma splited                                                                                                                            , example: cpu                                                                                                                                , mem                     , disk)`)
	flag.StringVar(&flagDatakitName, "name", "", `specify DataKit name                                                                                                                                                                                                                                                                                                             , example: prod-env-datakit`)
	flag.StringVar(&flagGlobalTags, "global-tags", "", `enable global tags                                                                                                                                                                                                                                                                                                               , example: host                                                                                                          = __datakit_hostname , ip     = __datakit_ip`)
	flag.StringVar(&flagProxy, "proxy", "", "http proxy http://ip:port for datakit")
	flag.StringVar(&flagDatakitHTTPListen, "listen", "localhost", "datakit HTTP listen")
	flag.StringVar(&flagNamespace, "namespace", "", "datakit namespace")
	flag.StringVar(&flagInstallLog, "install-log", "", "install log")
	flag.StringVar(&flagCloudProvider, "cloud-provider", "", "specify cloud provider                                                                                                                                                               ( accept aliyun/tencent/aws)")
	flag.IntVar(&flagDatakitHTTPPort, "port", 9529, "datakit HTTP port")
	flag.BoolVar(&flagInfo, "info", false, "show installer info")

	flag.BoolVar(&flagOffline, "offline", false, "-offline option removed")
	flag.BoolVar(&flagDownloadOnly, "download-only", false, "-download-only option removed")
}

func downloadFiles() {
	dl.CurDownloading = "datakit"
	if err := dl.Download(datakitUrl, datakit.InstallDir, true, false); err != nil {
		l.Fatal(err)
	}

	fmt.Printf("\n")

	dl.CurDownloading = "data"
	if err := dl.Download(dataUrl, datakit.InstallDir, true, false); err != nil {
		l.Fatal(err)
	}

	fmt.Printf("\n")
}

func main() {

	flag.Parse()

	if flagInfo {
		fmt.Printf(`
Version: %s
Build At: %s
Golang Version: %s
BaseUrl: %s
DataKit: %s
`, datakit.Version, git.BuildAt, git.Golang, datakitUrl, dataUrl)
		os.Exit(0)
	}

	if flagInstallLog == "" {
		if err := logger.InitRoot(
			&logger.Option{
				Level: logger.DEBUG,
				Flags: logger.OPT_DEFAULT | logger.OPT_STDOUT}); err != nil {
			l.Errorf("set root log faile: %s", err.Error())
		}
	} else {
		l.Infof("set log file to %s", flagInstallLog)

		if err := logger.InitRoot(&logger.Option{
			Path:  flagInstallLog,
			Level: logger.DEBUG,
			Flags: logger.OPT_DEFAULT}); err != nil {
			l.Errorf("set root log faile: %s", err.Error())
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

	if flagProxy != "" {

		if !strings.HasPrefix(flagProxy, "http") {
			flagProxy = "http://" + flagProxy
		}

		if _, err := url.Parse(flagProxy); err != nil {
			l.Warnf("bad proxy config expect http://ip:port given %s", flagProxy)
		} else {
			l.Infof("set proxy to %s", flagProxy)
		}
	}

	// 迁移老版本 datakit 数据目录
	mvOldDatakit(svc)

	downloadFiles() // download 过程直接覆盖已有安装

	datakit.InitDirs()

	if flagDKUpgrade { // upgrade new version
		l.Infof("Upgrading to version %s...", DataKitVersion)
		if err := upgradeDatakit(svc); err != nil {
			l.Fatalf("upgrade datakit: %s, ignored", err.Error())
		}
	} else { // install new datakit
		l.Infof("Installing version %s...", DataKitVersion)
		installNewDatakit(svc)
	}

	if !flagInstallOnly {
		l.Infof("starting service %s...", dkservice.ServiceName)
		if err = service.Control(svc, "start"); err != nil {
			l.Warnf("star service: %s, ignored", err.Error())
		}
	} else {
		l.Infof("only install service %s, NOT started", dkservice.ServiceName)
	}

	config.CreateSymlinks()

	if flagDKUpgrade {
		l.Info(":) Upgrade Success!")
	} else {
		l.Info(":) Install Success!")
	}

	promptReferences()
}

func promptReferences() {
	fmt.Printf("\n\tVisit http://%s:%d/man/changelog to see DataKit change logs.\n",
		flagDatakitHTTPListen,
		flagDatakitHTTPPort)
	fmt.Printf("\tVisit http://%s:%d/monitor to see DataKit running status.\n",
		flagDatakitHTTPListen,
		flagDatakitHTTPPort)
	fmt.Printf("\tVisit http://%s:%d/man to see DataKit manuals.\n\n",
		flagDatakitHTTPListen,
		flagDatakitHTTPPort)
}

func upgradeDatakit(svc service.Service) error {

	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}

	mc := config.Cfg

	if err := mc.LoadMainTOML(datakit.MainConfPath); err == nil {
		mc, _ = upgradeMainConfig(mc)

		if flagOTA {
			l.Debugf("set auto update flag")
			mc.AutoUpdate = flagOTA
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
	if flagOTA {
		l.Debugf("set auto update flag")
		mc.AutoUpdate = flagOTA
	}

	// accept any install options
	if flagGlobalTags != "" {
		l.Infof("set global tags...")
		mc.GlobalTags = config.ParseGlobalTags(flagGlobalTags)

		l.Infof("set global tags %+#v", mc.GlobalTags)
	}

	mc.Namespace = flagNamespace
	mc.HTTPAPI.Listen = fmt.Sprintf("%s:%d", flagDatakitHTTPListen, flagDatakitHTTPPort)
	mc.InstallDate = time.Now()

	if flagDatakitName != "" {
		mc.Name = flagDatakitName
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

	if flagEnableInputs == "" {
		flagEnableInputs = strings.Join(hostInputs, ",")
	} else {
		flagEnableInputs = flagEnableInputs + "," + strings.Join(hostInputs, ",")
	}

	mc.EnableDefaultsInputs(flagEnableInputs)

	switch flagCloudProvider {
	case "aliyun", "tencent", "aws":

		l.Infof("try set cloud provider to %s...", flagCloudProvider)

		if conf, err := preEnableHostobjectInput(flagCloudProvider); err != nil {
			l.Fatalf("failed to init hostobject conf: %s", err.Error())
		} else {
			cfgpath := filepath.Join(datakit.ConfdDir, "host", "hostobject.conf")
			if err := ioutil.WriteFile(cfgpath, conf, datakit.ConfPerm); err != nil {
				l.Fatalf("failed to init hostobject conf: %s", err.Error())
			}
		}

		l.Infof("set cloud provider to %s ok", flagCloudProvider)

	case "": //pass

	default:
		l.Warnf("unknown cloud provider %s, ignored", flagCloudProvider)
	}

	l.Debugf("main config:\n%s", mc.String())

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
		c.DataWay.HttpProxy = flagProxy
	}

	return c, nil
}

func getDataWayCfg() *dataway.DataWayCfg {
	dw := &dataway.DataWayCfg{}

	if flagDataway != "" {
		dw.URLs = strings.Split(flagDataway, ",")
		if err := dw.Apply(); err != nil {
			l.Fatal(err)
		}

		if flagProxy != "" {
			l.Debugf("set proxy to %s", flagProxy)
			dw.HttpProxy = flagProxy
		}
	} else {
		l.Fatal("should not been here")
	}

	return dw
}

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
