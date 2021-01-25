package install

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
	l                = logger.DefaultSLogger("install")
	lagacyInstallDir = ""

	OSArch = runtime.GOOS + "/" + runtime.GOARCH

	InstallDir    = ""
	DataWayHTTP   = ""
	DataWayWsPort = ""
	GlobalTags    = ""
	Port          = 9529
	DatakitName   = ""
	EnableInputs  = ""
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
			dc, err = datakit.ParseDataway(dwhttp, "")
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
		dc, err = datakit.ParseDataway(DataWayHTTP, DataWayWsPort)
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

	// prepare dataway info
	datakit.Cfg.MainCfg.DataWay = getDataWayCfg()

	// accept any install options
	if GlobalTags != "" {
		datakit.Cfg.MainCfg.GlobalTags = datakit.ParseGlobalTags(GlobalTags)
	}

	datakit.Cfg.MainCfg.HTTPBind = fmt.Sprintf("0.0.0.0:%d", Port)
	datakit.Cfg.MainCfg.InstallDate = time.Now()

	if DatakitName != "" {
		datakit.Cfg.MainCfg.Name = DatakitName
	}

	// XXX: load old datakit UUID file: reuse datakit UUID installed before
	if data, err := ioutil.ReadFile(datakit.UUIDFile); err != nil {
		datakit.Cfg.MainCfg.UUID = cliutils.XID("dkid_")
		if err := datakit.CreateUUIDFile(datakit.Cfg.MainCfg.UUID); err != nil {
			l.Fatalf("create datakit id failed: %s", err.Error())
		}
	} else {
		datakit.Cfg.MainCfg.UUID = string(data)
	}

	defaultHostInputs := "cpu,disk,diskio,mem,swap,system,hostobject"
	if EnableInputs == "" {
		EnableInputs = defaultHostInputs
	} else {
		EnableInputs = EnableInputs + "," + defaultHostInputs
	}

	datakit.Cfg.EnableDefaultsInputs(EnableInputs)

	// build datakit main config
	if err := datakit.Cfg.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	l.Infof("installing service %s...", datakit.ServiceName)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install service: %s, ignored", err.Error())
	}
}

func updateLagacyConfig(dir string) {
	cfgdata, err := ioutil.ReadFile(filepath.Join(dir, "datakit.conf"))
	if err != nil {
		l.Fatalf("read lagacy datakit.conf failed: %s", err.Error())
	}

	var maincfg datakit.MainConfig
	if _, err = bstoml.Decode(string(cfgdata), &maincfg); err != nil {
		l.Fatalf("unmarshal failed: %s", err.Error())
	}

	maincfg.Log = filepath.Join(InstallDir, "datakit.log") // reset log path
	maincfg.DeprecatedConfigDir = ""                       // remove conf.d config: we use static conf.d dir, *not* configurable

	// split origin ftdataway into dataway object
	var dwcfg *datakit.DataWayCfg
	if maincfg.DeprecatedFtGateway != "" {
		if dwcfg, err = datakit.ParseDataway(maincfg.DeprecatedFtGateway, ""); err != nil {
			l.Fatal(err)
		} else {
			maincfg.DeprecatedFtGateway = "" // deprecated
			maincfg.DataWay = dwcfg
		}
	}

	cfgdata, err = datakit.TomlMarshal(maincfg)
	if err != nil {
		l.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.Join(dir, "datakit.conf"), cfgdata, os.ModePerm); err != nil {
		l.Fatal(err)
	}
}

func upgradeMainConfigure(cfg *datakit.Config, mcp string) {

	datakit.MoveDeprecatedMainCfg()

	mcdata, err := ioutil.ReadFile(mcp)
	if err != nil {
		l.Fatalf("ioutil.ReadFile(): %s", err.Error())
	}

	if _, err := bstoml.Decode(string(mcdata), cfg.MainCfg); err != nil {
		l.Fatalf("unmarshal main cfg failed %s", err.Error())
	}

	mc := cfg.MainCfg

	if mc.DataWay.URL == "" { // use old-version configure fields to build @URL
		mc.DataWay.URL = fmt.Sprintf("%s://%s", mc.DataWay.DeprecatedScheme, mc.DataWay.DeprecatedHost)
	}

	if mc.DataWay.DeprecatedToken != "" {
		mc.DataWay.URL += fmt.Sprintf("?token=%s", mc.DataWay.DeprecatedToken)
	}

	// clear deprecated fields
	mc.DataWay.DeprecatedToken = ""
	mc.DataWay.DeprecatedHost = ""
	mc.DataWay.DeprecatedScheme = ""

	if err := cfg.InitCfg(mcp); err != nil {
		l.Fatal(err)
	}
}

func UpgradeDatakit(svc service.Service) {

	var lagacyServiceFiles []string = nil

	switch OSArch {

	case datakit.OSArchWinAmd64, datakit.OSArchWin386:
		lagacyInstallDir = `C:\Program Files\Forethought\datakit`
		if _, err := os.Stat(lagacyInstallDir); err != nil {
			lagacyInstallDir = `C:\Program Files (x86)\Forethought\datakit`
		}

	case datakit.OSArchLinuxArm,
		datakit.OSArchLinuxArm64,
		datakit.OSArchLinux386,
		datakit.OSArchLinuxAmd64,
		datakit.OSArchDarwinAmd64:

		lagacyInstallDir = `/usr/local/cloudcare/forethought/datakit`
		lagacyServiceFiles = []string{"/lib/systemd/system/datakit.service", "/etc/systemd/system/datakit.service"}
	default:
		l.Fatalf("%s not support", OSArch)
	}

	if _, err := os.Stat(lagacyInstallDir); err != nil {
		l.Debug("no lagacy datakit installed")

		// generate new main configure
		upgradeMainConfigure(datakit.Cfg, datakit.MainConfPath)
		return
	}

	stopLagacyDatakit(svc)
	updateLagacyConfig(lagacyInstallDir)

	// uninstall service, remove old datakit.service file(for UNIX OS)
	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service %s, ignored", err.Error())
	}

	for _, sf := range lagacyServiceFiles {
		if _, err := os.Stat(sf); err == nil {
			if err := os.Remove(sf); err != nil {
				l.Fatalf("remove %s failed: %s", sf, err.Error())
			}
		}
	}

	os.RemoveAll(InstallDir) // clean new install dir if exists

	// move all lagacy datakit files to new install dir
	if err := os.Rename(lagacyInstallDir, InstallDir); err != nil {
		l.Fatalf("remove %s failed: %s", InstallDir, err.Error())
	}

	for _, dir := range []string{datakit.TelegrafDir, datakit.DataDir, datakit.LuaDir, datakit.ConfdDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			l.Fatalf("create %s failed: %s", dir, err)
		}
	}

	l.Infof("installing service %s...", datakit.ServiceName)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install service: %s, ignored", err.Error())
	}
}

func stopLagacyDatakit(svc service.Service) {
	switch OSArch {
	case datakit.OSArchWinAmd64, datakit.OSArchWin386:

		if err := service.Control(svc, "stop"); err != nil {
			l.Warnf("stop service: %s, ignored", err.Error())
		}

	default:
		cmd := exec.Command(`stop`, []string{datakit.ServiceName}...) //nolint:gosec
		if _, err := cmd.Output(); err != nil {
			l.Debugf("upstart stop datakit failed, try systemctl...")
		} else {
			return
		}

		cmd = exec.Command("systemctl", []string{"stop", datakit.ServiceName}...) //nolint:gosec
		if _, err := cmd.Output(); err != nil {
			l.Debugf("systemctl stop datakit failed, ignored")
		}
	}
}
