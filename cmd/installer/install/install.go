package install

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	l = logger.DefaultSLogger("install")

	DefaultHostInputs = []string{"cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes"}

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
		datakit.Cfg.DataWay.Urls = dwUrls
		dc, err = datakit.ParseDataway(datakit.Cfg.DataWay.Urls)
		if err != nil {
			l.Fatal(err)
		}

		if err := dc.Test(); err != nil {
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
	if EnableInputs == "" {
		EnableInputs = strings.Join(DefaultHostInputs, ",")
	} else {
		EnableInputs = EnableInputs + "," + strings.Join(DefaultHostInputs, ",")
	}

	mc.EnableDefaultsInputs(EnableInputs)

	// build datakit main config
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}
}

func upgradeMainConfigure(cfg *datakit.Config, mcp string) error {

	datakit.MoveDeprecatedCfg()

	mcdata, err := ioutil.ReadFile(mcp)
	if err != nil {
		return err
	}

	if _, err := bstoml.Decode(string(mcdata), cfg); err != nil {
		return err
	}

	mc := cfg

	if mc.DataWay.DeprecatedURL == "" { // use old-version configure fields to build @URL
		mc.DataWay.DeprecatedURL = fmt.Sprintf("%s://%s", mc.DataWay.DeprecatedScheme, mc.DataWay.DeprecatedHost)
	}

	if mc.DataWay.DeprecatedToken != "" {
		mc.DataWay.DeprecatedURL += fmt.Sprintf("?token=%s", mc.DataWay.DeprecatedToken)
	}

	// clear deprecated fields
	mc.DataWay.DeprecatedToken = ""
	mc.DataWay.DeprecatedHost = ""
	mc.DataWay.DeprecatedScheme = ""

	for _, v := range DefaultHostInputs {
		exists := false
		for _, iv := range mc.DefaultEnabledInputs {
			if v == iv {
				exists = true
				break
			}
		}
		if !exists {
			mc.DefaultEnabledInputs = append(mc.DefaultEnabledInputs, v)
		}
	}

	//backup datakit.conf
	backfile := mcp + fmt.Sprintf(".bkp.%v", time.Now().Unix())
	ioutil.WriteFile(backfile, mcdata, 0664)

	return cfg.InitCfg(mcp)
}

func UpgradeDatakit(svc service.Service) error {

	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}

	mc := datakit.Cfg
	if err := mc.LoadMainConfig(datakit.MainConfPath); err == nil {
		mc.DataWay.DeprecatedURL = ""

		// 将原来的日志位置，改成 /var/log/datakit 目录下(mac/linux, windows 继续维持原样)
		if runtime.GOOS != datakit.OSWindows {
			mc.Log = "/var/log/datakit/log"
			l.Debugf("set log to %s, remove ", mc.Log)

			mc.GinLog = "/var/log/datakit/gin.log"
			l.Debugf("set gin log to %s", mc.GinLog)
		}

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
