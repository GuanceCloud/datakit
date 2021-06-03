package install

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	l = logger.DefaultSLogger("install")

	DefaultHostInputs          = []string{"cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes"}
	DefaultHostInputsWithLinux = []string{"cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes", "container"}

	OSArch = runtime.GOOS + "/" + runtime.GOARCH

	DataWayHTTP  = ""
	GlobalTags   = ""
	Port         = 9529
	DatakitName  = ""
	EnableInputs = ""
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

func getDataWayCfg() *datakit.DataWayCfg {
	var dc *datakit.DataWayCfg
	var err error

	if DataWayHTTP == "" {
		for {
			dwhttp := readInput("Please set DataWay HTTP URL(http[s]://host:port?token=xxx) > ")
			dwUrls := []string{dwhttp}
			dc, err = datakit.ParseDataway(dwUrls)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}
			if err := dc.Test(); err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}
			break
		}
	} else {
		dwUrls := []string{DataWayHTTP}
		datakit.Cfg.DataWay.URLs = dwUrls
		dc, err = datakit.ParseDataway(datakit.Cfg.DataWay.URLs)
		if err != nil {
			l.Fatal(err)
		}
	}

	return dc
}

func InstallNewDatakit(svc service.Service) {

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service: %s, ignored", err.Error())
	}

	mc := datakit.Cfg

	// prepare dataway info
	mc.DataWay = getDataWayCfg()

	// accept any install options
	if GlobalTags != "" {
		mc.GlobalTags = datakit.ParseGlobalTags(GlobalTags)
	}

	mc.HTTPListen = fmt.Sprintf("localhost:%d", Port)
	mc.InstallDate = time.Now()

	if DatakitName != "" {
		mc.Name = DatakitName
	}

	// XXX: load old datakit UUID file: reuse datakit UUID installed before
	if data, err := ioutil.ReadFile(datakit.UUIDFile); err != nil {
		mc.UUID = cliutils.XID("dkid_")
		if err := datakit.CreateUUIDFile(datakit.UUIDFile, mc.UUID); err != nil {
			l.Fatalf("create datakit id failed: %s", err.Error())
		}
	} else {
		mc.UUID = string(data)
	}

	writeDefInputToMainCfg(mc)

	l.Infof("installing service %s...", datakit.ServiceName)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install service: %s, ignored", err.Error())
	}
}

func writeDefInputToMainCfg(mc *datakit.Config) {

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

	// build datakit main config
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}
}

func upgradeMainConfig(c *datakit.Config) (*datakit.Config, error) {

	if c.DataWay != nil {
		c.DataWay.DeprecatedURL = ""
	}

	// XXX: 无脑更改日志位置
	switch runtime.GOOS {
	case datakit.OSWindows:
		c.Log = filepath.Join(datakit.InstallDir, "log")
		c.GinLog = filepath.Join(datakit.InstallDir, "gin.log")
	default:
		c.Log = "/var/log/datakit/log"
		c.GinLog = "/var/log/datakit/gin.log"
	}
	l.Debugf("set log to %s, remove ", c.Log)
	l.Debugf("set gin log to %s", c.GinLog)

	return c, nil
}

func UpgradeDatakit(svc service.Service) error {

	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}

	mc := datakit.Cfg
	if err := mc.LoadMainConfig(datakit.MainConfPath); err == nil {
		mc, _ = upgradeMainConfig(mc)
		writeDefInputToMainCfg(mc)
	} else {
		l.Warnf("load main config: %s, ignored", err.Error())
		return err
	}

	for _, dir := range []string{datakit.DataDir, datakit.ConfdDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
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
