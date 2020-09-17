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

	"github.com/influxdata/toml"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	l                = logger.DefaultSLogger("install")
	lagacyInstallDir = ""

	OSArch = runtime.GOOS + "/" + runtime.GOARCH

	InstallDir   = ""
	DataWay      = ""
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
	if DataWay == "" {
		for {
			dw := readInput("Please set DataWay request URL(http://IP:Port/v1/write/metric) > ")
			dwcfg, err := datakit.ParseDataway(dw)
			if err == nil {
				return dwcfg
			}

			fmt.Printf("%s\n", err.Error())
			continue
		}
	} else {
		dwcfg, err := datakit.ParseDataway(DataWay)
		if err != nil {
			l.Fatal(err)
		}
		return dwcfg
	}
}

func InstallNewDatakit(svc service.Service) {

	_ = uninstallDataKitService(svc) // uninstall service if installed before

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

	datakit.Cfg.MainCfg.UUID = cliutils.XID("dkid_")

	datakit.Cfg.EnableDefaultsInputs(EnableInputs)

	// build datakit main config
	if err := datakit.Cfg.InitCfg(); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	l.Infof("installing service %s...", datakit.ServiceName)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("fail to install service %s: %s, ignored", datakit.ServiceName, err.Error())
	}
}

func updateLagacyConfig(dir string) {
	cfgdata, err := ioutil.ReadFile(filepath.Join(dir, "datakit.conf"))
	if err != nil {
		l.Fatalf("read lagacy datakit.conf failed: %s", err.Error())
	}

	var maincfg datakit.MainConfig
	if err = toml.Unmarshal(cfgdata, &maincfg); err != nil {
		l.Fatalf("toml unmarshal failed: %s", err.Error())
	}

	maincfg.Log = filepath.Join(InstallDir, "datakit.log") // reset log path
	maincfg.DeprecatedConfigDir = ""                       // remove conf.d config: we use static conf.d dir, *not* configurable

	// split origin ftdataway into dataway object
	var dwcfg *datakit.DataWayCfg
	if maincfg.DeprecatedFtGateway != "" {
		if dwcfg, err = datakit.ParseDataway(maincfg.DeprecatedFtGateway); err != nil {
			l.Fatal(err)
		} else {
			maincfg.DeprecatedFtGateway = "" // deprecated
			maincfg.DataWay = dwcfg
		}
	}

	cfgdata, err = toml.Marshal(maincfg)
	if err != nil {
		l.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.Join(dir, "datakit.conf"), cfgdata, os.ModePerm); err != nil {
		l.Fatal(err)
	}
}

func MigrateLagacyDatakit(svc service.Service) {

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
		return
	}

	stopLagacyDatakit(svc)
	updateLagacyConfig(lagacyInstallDir)

	// uninstall service, remove old datakit.service file(for UNIX OS)
	_ = uninstallDataKitService(svc)
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
		l.Warnf("fail to register service %s: %s, ignored", datakit.ServiceName, err.Error())
	}
}

func stopLagacyDatakit(svc service.Service) {
	switch OSArch {
	case datakit.OSArchWinAmd64, datakit.OSArchWin386:
		_ = StopDataKitService(svc)
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

func uninstallDataKitService(s service.Service) error {
	if err := service.Control(s, "uninstall"); err != nil {
		l.Warnf("stop service datakit failed: %s, ignored", err.Error())
		return err
	}

	return nil
}

func StopDataKitService(s service.Service) error {

	if err := service.Control(s, "stop"); err != nil {
		l.Warnf("stop service datakit failed: %s, ignored", err.Error())
		return err
	}

	return nil
}
