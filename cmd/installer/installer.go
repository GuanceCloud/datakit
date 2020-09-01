package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/influxdata/toml"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	DataKitBaseURL = ""
	DataKitVersion = ""
	installDir     = ""

	datakitUrl = "https://" + path.Join(DataKitBaseURL,
		fmt.Sprintf("datakit-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion))

	telegrafUrl = "https://" + path.Join(DataKitBaseURL,
		"telegraf",
		fmt.Sprintf("agent-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH))

	curDownloading = ""

	osarch = runtime.GOOS + "/" + runtime.GOARCH
	//svc              service.Service
	lagacyInstallDir = ""

	l *logger.Logger
)

var (
	flagUpgrade      = flag.Bool("upgrade", false, ``)
	flagDataway      = flag.String("dataway", "", `address of dataway(http://IP:Port/v1/write/metric), port default 9528`)
	flagInfo         = flag.Bool("info", false, "show installer info")
	flagDownloadOnly = flag.Bool("download-only", false, `download datakit only, not install`)

	flagEnableInputs = flag.String("enable-inputs", "", `default enable inputs(comma splited, example: cpu,mem,disk)`)
	flagDatakitName  = flag.String("name", "", `specify DataKit name, example: prod-env-datakit`)
	flagGlobalTags   = flag.String("global-tags", "", `enable global tags, example: host=$datakit_hostname,from=$datakit_id`)
	flagPort         = flag.Int("port", 9529, "datakit HTTP port")

	flagOffline = flag.Bool("offline", false, "offline install mode")
	flagSrcs    = flag.String("srcs", fmt.Sprintf("./datakit-%s-%s-%s.tar.gz,./agent-%s-%s.tar.gz",
		runtime.GOOS, runtime.GOARCH, DataKitVersion, runtime.GOOS, runtime.GOARCH),
		`local path of datakit and agent install files`)
)

const (
	datakitBin = "datakit"
	dlDatakit  = "datakit"
	dlAgent    = "agent"
)

func main() {
	preInit()
	flag.Parse()
	datakit.InitDirs()
	applyFlags()

	// create install dir if not exists
	if err := os.MkdirAll(installDir, 0775); err != nil {
		l.Fatal(err)
	}

	datakit.ServiceExecutable = filepath.Join(installDir, datakitBin)
	if runtime.GOOS == datakit.OSWindows {
		datakit.ServiceExecutable += ".exe"
	}

	svc, err := datakit.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	l.Info("stoping datakit...")
	_ = stopDataKitService(svc) // stop service if installed before

	if *flagOffline && *flagSrcs != "" {
		for _, f := range strings.Split(*flagSrcs, ",") {
			extractDatakit(f, installDir)
		}
	} else {
		curDownloading = dlDatakit
		doDownload(datakitUrl, installDir)

		curDownloading = dlAgent
		doDownload(telegrafUrl, installDir)
	}

	if *flagUpgrade { // upgrade new version
		l.Infof("Upgrading to version %s...", DataKitVersion)
		migrateLagacyDatakit(svc)
	} else { // install new datakit
		l.Infof("Installing version %s...", DataKitVersion)
		installNewDatakit(svc)
	}

	l.Infof("starting service %s...", datakit.ServiceName)
	if err = startDatakitService(svc); err != nil {
		l.Fatalf("fail to star service %s: %s", datakit.ServiceName, err.Error())
	}

	if *flagUpgrade { // upgrade new version
		l.Info(":) Upgrade Success!")
	} else {
		l.Info(":) Install Success!")
	}

	localIP, err := datakit.LocalIP()
	if err != nil {
		l.Info("get local IP failed: %s", err.Error())
	} else {
		fmt.Printf("\n\tVisit http://%s:%d/stats to see DataKit running status.\n\n", localIP, *flagPort)
	}
}

func getDataWayCfg() *datakit.DataWayCfg {
	if *flagDataway == "" {
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
		dwcfg, err := datakit.ParseDataway(*flagDataway)
		if err != nil {
			l.Fatal(err)
		}
		return dwcfg
	}
}

func installNewDatakit(svc service.Service) {

	_ = uninstallDataKitService(svc) // uninstall service if installed before

	// prepare dataway info
	datakit.Cfg.MainCfg.DataWay = getDataWayCfg()

	// accept any install options
	if *flagGlobalTags != "" {
		datakit.Cfg.MainCfg.GlobalTags = datakit.ParseGlobalTags(*flagGlobalTags)
	}

	datakit.Cfg.MainCfg.HTTPBind = fmt.Sprintf("0.0.0.0:%d", *flagPort)
	datakit.Cfg.MainCfg.InstallDate = time.Now()

	if *flagDatakitName != "" {
		datakit.Cfg.MainCfg.Name = *flagDatakitName
	}

	datakit.Cfg.MainCfg.UUID = cliutils.XID("dkid_")

	datakit.Cfg.EnableDefaultsInputs(*flagEnableInputs)

	// build datakit main config
	if err := datakit.Cfg.InitCfg(); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	l.Infof("installing service %s...", datakit.ServiceName)
	if err := installDatakitService(svc); err != nil {
		l.Warnf("fail to install service %s: %s, ignored", datakit.ServiceName, err.Error())
	}
}

func applyFlags() {

	if *flagInfo {
		fmt.Printf(`
       Version: %s
      Build At: %s
Golang Version: %s
       BaseUrl: %s
       DataKit: %s
      Telegraf: %s
`, git.Version, git.BuildAt, git.Golang, DataKitBaseURL, datakitUrl, telegrafUrl)
		os.Exit(0)
	}

	if *flagDownloadOnly {
		curDownloading = dlDatakit
		doDownload(datakitUrl, fmt.Sprintf("datakit-%s-%s-%s.tar.gz",
			runtime.GOOS, runtime.GOARCH, DataKitVersion))

		curDownloading = dlAgent
		doDownload(telegrafUrl, fmt.Sprintf("agent-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH))

		os.Exit(0)
	}

	switch osarch {

	case datakit.OSArchWinAmd64:
		installDir = `C:\Program Files\dataflux\datakit`

	case datakit.OSArchWin386:
		installDir = `C:\Program Files (x86)\dataflux\datakit`

	case datakit.OSArchLinuxArm,
		datakit.OSArchLinuxArm64,
		datakit.OSArchLinux386,
		datakit.OSArchLinuxAmd64,
		datakit.OSArchDarwinAmd64:
		installDir = `/usr/local/cloudcare/dataflux/datakit`

	default:
		// TODO: more os/arch support
	}
}

func readInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	txt, err := reader.ReadString('\n')
	if err != nil {
		l.Fatal(err)
	}

	return strings.TrimSpace(txt)
}

func _doDownload(r io.Reader, to string) {

	f, err := os.OpenFile(to, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		l.Fatal(err)
	}

	if _, err := io.Copy(f, r); err != nil { //nolint:gosec
		l.Fatal(err)
	}

	f.Close()
}

func doExtract(r io.Reader, to string) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		l.Fatal(err)
	}

	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		switch {
		case err == io.EOF:
			return
		case err != nil:
			l.Fatal(err)
		case hdr == nil:
			continue
		}

		target := filepath.Join(to, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					l.Fatal(err)
				}
			}

		case tar.TypeReg:

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				l.Fatal(err)
			}

			// TODO: lock file before extracting, to avoid `text file busy` error
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				l.Fatal(err)
			}

			if _, err := io.Copy(f, tr); err != nil { //nolint:gosec
				l.Fatal(err)
			}

			f.Close()
		}
	}
}

func extractDatakit(gz, to string) {
	data, err := os.Open(gz)
	if err != nil {
		l.Fatalf("open file %s failed: %s", gz, err)
	}

	defer data.Close()

	doExtract(data, to)
}

type writeCounter struct {
	total   uint64
	current uint64
	last    float64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.current += uint64(n)
	wc.last += float64(n)
	wc.PrintProgress()
	return n, nil
}

func doDownload(from, to string) {
	resp, err := http.Get(from) //nolint:gosec
	if err != nil {
		l.Fatalf("failed to download %s: %s", from, err)
	}

	defer resp.Body.Close()
	cnt := &writeCounter{
		total: uint64(resp.ContentLength),
	}

	if *flagDownloadOnly {
		_doDownload(io.TeeReader(resp.Body, cnt), to)
	} else {
		doExtract(io.TeeReader(resp.Body, cnt), to)
	}
	fmt.Printf("\n")
}

func (wc *writeCounter) PrintProgress() {
	if wc.last > float64(wc.total)*0.01 || wc.current == wc.total { // update progress-bar each 1%
		fmt.Printf("\r%s", strings.Repeat(" ", 36))
		fmt.Printf("\rDownloading(% 7s)... %s/%s", curDownloading, humanize.Bytes(wc.current), humanize.Bytes(wc.total))
		wc.last = 0.0
	}
}

func stopDataKitService(s service.Service) error {

	if err := service.Control(s, "stop"); err != nil {
		l.Warnf("stop service datakit failed: %s, ignored", err.Error())
		return err
	}

	return nil
}

func uninstallDataKitService(s service.Service) error {
	if err := service.Control(s, "uninstall"); err != nil {
		l.Warnf("stop service datakit failed: %s, ignored", err.Error())
		return err
	}

	return nil
}

func installDatakitService(s service.Service) error {
	return service.Control(s, "install")
}

func startDatakitService(s service.Service) error {
	return service.Control(s, "start")
}

func stopLagacyDatakit(svc service.Service) {
	switch osarch {
	case datakit.OSArchWinAmd64, datakit.OSArchWin386:
		_ = stopDataKitService(svc)
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

func updateLagacyConfig(dir string) {
	cfgdata, err := ioutil.ReadFile(filepath.Join(dir, "datakit.conf"))
	if err != nil {
		l.Fatalf("read lagacy datakit.conf failed: %s", err.Error())
	}

	var maincfg datakit.MainConfig
	if err = toml.Unmarshal(cfgdata, &maincfg); err != nil {
		l.Fatalf("toml unmarshal failed: %s", err.Error())
	}

	maincfg.Log = filepath.Join(installDir, "datakit.log") // reset log path
	maincfg.ConfigDir = ""                                 // remove conf.d config: we use static conf.d dir, *not* configurable

	// split origin ftdataway into dataway object
	var dwcfg *datakit.DataWayCfg
	if maincfg.FtGateway != "" {
		if dwcfg, err = datakit.ParseDataway(maincfg.FtGateway); err != nil {
			l.Fatal(err)
		} else {
			maincfg.FtGateway = "" // deprecated
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

func migrateLagacyDatakit(svc service.Service) {

	var lagacyServiceFiles []string = nil

	switch osarch {

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
		l.Fatalf("%s not support", osarch)
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

	os.RemoveAll(installDir) // clean new install dir if exists

	// move all lagacy datakit files to new install dir
	if err := os.Rename(lagacyInstallDir, installDir); err != nil {
		l.Fatalf("remove %s failed: %s", installDir, err.Error())
	}

	for _, dir := range []string{datakit.TelegrafDir, datakit.DataDir, datakit.LuaDir, datakit.ConfdDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			l.Fatalf("create %s failed: %s", dir, err)
		}
	}

	l.Infof("installing service %s...", datakit.ServiceName)
	if err := installDatakitService(svc); err != nil {
		l.Warnf("fail to register service %s: %s, ignored", datakit.ServiceName, err.Error())
	}
}

func preInit() {

	lopt := logger.OPT_DEFAULT | logger.OPT_STDOUT
	if runtime.GOOS != "windows" { // disable color on windows(some color not working under windows)
		lopt |= logger.OPT_COLOR
	}

	logger.SetGlobalRootLogger("", logger.DEBUG, lopt)
	l = logger.SLogger("installer")
}
