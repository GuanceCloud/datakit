package install

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/service"
)

var (
	l = logger.DefaultSLogger("install")

	DefaultHostInputs          = []string{"cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes"}
	DefaultHostInputsWithLinux = []string{"cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes", "docker"}

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

	// accept any install options
	if GlobalTags != "" {
		mc.GlobalTags = config.ParseGlobalTags(GlobalTags)
	}

	mc.HTTPListen = fmt.Sprintf("localhost:%d", Port)
	mc.InstallDate = time.Now()

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

	// build datakit main config
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}
}

func upgradeMainConfig(c *config.Config) (*config.Config, error) {

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

	mc := config.Cfg
	if err := mc.LoadMainTOML(datakit.MainConfPath); err == nil {
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
