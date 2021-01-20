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

	datakit.Cfg.MainCfg.UUID = cliutils.XID("dkid_")

	datakit.Cfg.EnableDefaultsInputs(EnableInputs)

	// build datakit main config
	if err := datakit.Cfg.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	//default enable host inputs when install
	enabledHostInputs()

	l.Infof("installing service %s...", datakit.ServiceName)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install service: %s, ignored", err.Error())
	}
}

func enabledHostInputs() {

	cfgs := map[string]string{
		`cpu`: `
[[inputs.cpu]]
## Whether to report per-cpu stats or not
percpu = false
## Whether to report total system cpu stats or not
totalcpu = true
## If true, collect raw CPU time metrics.
collect_cpu_time = false
## If true, compute and report the sum of all non-idle CPU states.
report_active = false
`,

		`disk`: `
[[inputs.disk]]
## By default stats will be gathered for all mount points.
## Set mount_points will restrict the stats to only the specified mount points.
# mount_points = ["/"]
	  
## Ignore mount points by filesystem type.
ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]
`,

		`diskio`: `# Read metrics about disk IO by device
[[inputs.diskio]]
## By default, telegraf will gather stats for all devices including
## disk partitions.
## Setting devices will restrict the stats to the specified devices.
# devices = ["sda", "sdb"]
## Uncomment the following line if you need disk serial numbers.
# skip_serial_number = false
#
## On systems which support it, device metadata can be added in the form of
## tags.
## Currently only Linux is supported via udev properties. You can view
## available properties for a device by running:
## 'udevadm info -q property -n /dev/sda'
## Note: Most, but not all, udev properties can be accessed this way. Properties
## that are currently inaccessible include DEVTYPE, DEVNAME, and DEVPATH.
# device_tags = ["ID_FS_TYPE", "ID_FS_USAGE"]
#
## Using the same metadata source as device_tags, you can also customize the
## name of the device via templates.
## The 'name_templates' parameter is a list of templates to try and apply to
## the device. The template may contain variables in the form of '$PROPERTY' or
## '${PROPERTY}'. The first template which does not contain any variables not
## present for the device is used as the device name tag.
## The typical use case is for LVM volumes, to get the VG/LV name instead of
## the near-meaningless DM-0 name.
# name_templates = ["$ID_FS_LABEL","$DM_VG_NAME/$DM_LV_NAME"]
`,

		`mem`: `# Read metrics about memory usage
[[inputs.mem]]
# no configuration
`,

		`swap`: `# Read metrics about swap memory usage
[[inputs.swap]]
# no configuration
`,

		`system`: `# Read metrics about system load & uptime
[[inputs.system]]
# no configuration
`,

		`hostobject`: `
[inputs.hostobject]
# ##(optional) collect interval, default is 5 miniutes
interval = '5m'

# ##(optional) 
#pipeline = ''
`,
	}

	for name, sample := range cfgs {

		fpath := filepath.Join(datakit.ConfdDir, "host", name+".conf")

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			l.Errorf("mkdir failed: %s, ignored", err.Error())
			continue
		}

		if err := ioutil.WriteFile(fpath, []byte(sample), 0664); err != nil {
			l.Errorf("write input %s config failed: %s, ignored", name, err.Error())
			continue
		}

		l.Infof("enable input %s ok", name)
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
