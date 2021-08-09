package install

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
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
	l = logger.DefaultSLogger("install")

	DefaultHostInputs          = []string{"cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes"}
	DefaultHostInputsWithLinux = []string{"cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes", "container"}

	OSArch = runtime.GOOS + "/" + runtime.GOARCH

	DataWayHTTP   = ""
	GlobalTags    = ""
	Port          = 9529
	Listen        = "localhost"
	CloudProvider = ""
	DatakitName   = ""
	EnableInputs  = ""
	Namespace     = ""
	OTA           = false
	Proxy         = ""
)

func readInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	txt, err := reader.ReadString('\n')
	if err != nil {
		l.Fatal(err)
	}

	return strings.TrimSpace(txt)
}

func getDataWayCfg() *dataway.DataWayCfg {
	dw := &dataway.DataWayCfg{}

	if DataWayHTTP == "" {

		for {
			dwhttp := readInput("Please set DataWay HTTP URL(http[s]://host:port?token=xxx) > ")

			dwurls := strings.Split(dwhttp, ",")
			dw.URLs = dwurls
			if err := dw.Apply(); err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}

			break
		}
	} else {
		dw.URLs = strings.Split(DataWayHTTP, ",")
		if err := dw.Apply(); err != nil {
			l.Fatal(err)
		}
	}

	return dw
}

func InstallNewDatakit(svc service.Service) {

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service: %s, ignored", err.Error())
	}

	mc := config.Cfg

	// prepare dataway info
	mc.DataWay = getDataWayCfg()
	if OTA {
		l.Debugf("set auto update flag")
		mc.AutoUpdate = OTA
	}

	// accept any install options
	if GlobalTags != "" {
		mc.GlobalTags = config.ParseGlobalTags(GlobalTags)
	}

	mc.Namespace = Namespace
	mc.HTTPAPI.Listen = fmt.Sprintf("%s:%d", Listen, Port)
	mc.InstallDate = time.Now()

	if mc.DataWay != nil {
		mc.DataWay.HttpProxy = Proxy
	}

	if DatakitName != "" {
		mc.Name = DatakitName
	}

	writeDefInputToMainCfg(mc)

	l.Infof("installing service %s...", dkservice.ServiceName)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install service: %s, ignored", err.Error())
	}
}

func writeDefInputToMainCfg(mc *config.Config) {

	var hostInputs = DefaultHostInputs
	if runtime.GOOS == datakit.OSLinux {
		hostInputs = DefaultHostInputsWithLinux
	}

	if EnableInputs == "" {
		EnableInputs = strings.Join(hostInputs, ",")
	} else {
		EnableInputs = EnableInputs + "," + strings.Join(hostInputs, ",")
	}

	mc.EnableDefaultsInputs(EnableInputs)

	switch CloudProvider {
	case "aliyun", "tencent", "aws":
		if conf, err := preEnableHostobjectInput(CloudProvider); err != nil {
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

## Ignore mount points by filesystem type. Default ingore following FS types
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
		c.DataWay.HttpProxy = Proxy
	}

	return c, nil
}

func UpgradeDatakit(svc service.Service) error {

	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}

	mc := config.Cfg

	if err := mc.LoadMainTOML(datakit.MainConfPath); err == nil {
		mc, _ = upgradeMainConfig(mc)

		if OTA {
			l.Debugf("set auto update flag")
			mc.AutoUpdate = OTA
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

func Init() {
	l = logger.SLogger("install")
}
