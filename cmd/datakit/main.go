package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	nhttp "net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/blang/semver/v4"
	flag "github.com/spf13/pflag"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/install"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
)

var (
	flagVersion = flag.BoolP("version", "v", false, `show version info`)
	flagDocker  = flag.Bool("docker", false, "run within docker")

	// tool-commands supported in datakit
	flagCmd      = flag.Bool("cmd", false, "run datakit under command line mode")
	flagPipeline = flag.String("pl", "", "pipeline script to test(name only, do not use file path)")
	flagText     = flag.String("txt", "", "text string for the pipeline or grok(json or raw text)")

	flagGrokq           = flag.Bool("grokq", false, "query groks interactively")
	flagMan             = flag.Bool("man", false, "read manuals of inputs")
	flagOTA             = flag.Bool("update", false, "update datakit new version if available")
	flagAcceptRCVersion = flag.Bool("accept-rc-version", false, "during OTA, accept RC version if available")

	flagShowTestingVersions = flag.Bool("show-testing-version", false, "show testing versions on -version flag")

	flagInstallExternal = flag.String("install", "", "install external tool/software")
	flagStart      = flag.Bool("start", false, "start datakit")
	flagStop       = flag.Bool("stop", false, "stop datakit")
	flagRestart    = flag.Bool("restart", false, "restart datakit")
	flagReload     = flag.Bool("reload", false, "reload datakit")
	flagReloadPort = flag.Int("reload-port", 9529, "datakit http server port")
	flagExportMan  = flag.String("export-manuals", "", "export all inputs and related manuals to specified path")
	flagIgnoreMans = flag.String("ignore-manuals", "", "disable exporting specified manuals, i.e., --ignore-manuals nginx,redis,mem")
)

var (
	l = logger.DefaultSLogger("main")

	ReleaseType = ""
)

func main() {
	flag.CommandLine.MarkHidden("cmd")
	flag.CommandLine.MarkHidden("show-testing-version")
	flag.CommandLine.SortFlags = false
	flag.ErrHelp = errors.New("") // disable `pflag: help requested`

	flag.Parse()

	applyFlags()

	tryLoadConfig()

	// This may throw `Unix syslog delivery error` within docker, so we just
	// start the entry under docker.
	if *flagDocker {
		run()
	} else {
		datakit.Entry = run
		if err := datakit.StartService(); err != nil {
			l.Errorf("start service failed: %s", err.Error())
			return
		}
	}

	l.Info("datakit exited")
}

const (
	winUpgradeCmd = `Import-Module bitstransfer; ` +
		`start-bitstransfer -source %s -destination .dk-installer.exe; ` +
		`.dk-installer.exe -upgrade; ` +
		`rm .dk-installer.exe`
	unixUpgradeCmd = `sudo -- sh -c ` +
		`"curl %s -o dk-installer ` +
		`&& chmod +x ./dk-installer ` +
		`&& ./dk-installer -upgrade ` +
		`&& rm -rf ./dk-installer"`
)

func applyFlags() {

	if *flagVersion {
		fmt.Printf(`
       Version: %s
        Commit: %s
        Branch: %s
 Build At(UTC): %s
Golang Version: %s
      Uploader: %s
ReleasedInputs: %s
`, git.Version, git.Commit, git.Branch, git.BuildAt, git.Golang, git.Uploader, ReleaseType)
		vers, err := getOnlineVersions()
		if err != nil {
			fmt.Printf("Get online version failed: \n%s\n", err.Error())
			os.Exit(-1)
		}
		curver, err := getLocalVersion()
		if err != nil {
			fmt.Printf("Get local version failed: \n%s\n", err.Error())
			os.Exit(-1)
		}

		for k, v := range vers {

			if isNewVersion(v, curver, true) { // show version info, also show RC verison info
				fmt.Println("---------------------------------------------------")
				fmt.Printf("\n\n%s version available: %s, commit %s (release at %s)\n",
					k, v.version, v.Commit, v.ReleaseDate)
				switch runtime.GOOS {
				case "windows":
					cmdWin := fmt.Sprintf(winUpgradeCmd, v.downloadURL)
					fmt.Printf("\nUpgrade:\n\t%s\n\n", cmdWin)
				default:
					cmd := fmt.Sprintf(unixUpgradeCmd, v.downloadURL)
					fmt.Printf("\nUpgrade:\n\t%s\n\n", cmd)
				}
			}
		}

		os.Exit(0)
	}

	if *flagOTA {

		logger.SetGlobalRootLogger(datakit.OTALogFile, logger.DEBUG, logger.OPT_DEFAULT)
		l = logger.SLogger("ota")

		install.Init()

		l.Debugf("get online version...")
		vers, err := getOnlineVersions()
		if err != nil {
			l.Errorf("Get online version failed: \n%s\n", err.Error())
			os.Exit(0)
		}

		ver := vers["Online"]

		l.Debugf("online version: %v", ver)

		curver, err := getLocalVersion()
		if err != nil {
			l.Errorf("Get online version failed: \n%s\n", err.Error())
			os.Exit(-1)
		}

		if ver != nil && isNewVersion(ver, curver, *flagAcceptRCVersion) {
			l.Infof("New online version available: %s, commit %s (release at %s)",
				ver.version, ver.Commit, ver.ReleaseDate)
			if err := tryOTAUpdate(ver.VersionString); err != nil {
				l.Errorf("OTA failed: %s", err.Error())
				os.Exit(-1)
			}
			l.Infof("OTA success, new verison is %s", ver.VersionString)
		} else {
			l.Infof("OTA up to date(%s)", curver.VersionString)
		}

		os.Exit(0)
	}

	datakit.EnableUncheckInputs = (ReleaseType == "all")

	runDatakitWithCmd()

	if *flagDocker {
		datakit.Docker = true
	}
}

func dumpAllConfigSamples(fpath string) {

	if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
		panic(err)
	}

	for k, v := range inputs.Inputs {
		sample := v().SampleConfig()
		if err := ioutil.WriteFile(filepath.Join(fpath, k+".conf"), []byte(sample), os.ModePerm); err != nil {
			panic(err)
		}
	}

}

func run() {

	l.Info("datakit start...")
	if err := runDatakitWithHTTPServer(); err != nil {
		return
	}

	l.Info("datakit start ok. Wait signal or service stop...")

	// NOTE:
	// Actually, the datakit process been managed by system service, no matter on
	// windows/UNIX, datakit should exit via `service-stop' operation, so the signal
	// branch should not reached, but for daily debugging(ctrl-c), we kept the signal
	// exit option.
	signals := make(chan os.Signal, datakit.CommonChanCap)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-signals:
		if sig == syscall.SIGHUP {
			// TODO: reload configures
		} else {
			l.Infof("get signal %v, wait & exit", sig)
			http.HttpStop()
			datakit.Quit()
		}

	case <-datakit.StopCh:
		l.Infof("service stopping")
		http.HttpStop()
		datakit.Quit()
	}

	l.Info("datakit exit.")
}

func tryLoadConfig() {
	datakit.MoveDeprecatedMainCfg()

	for {
		if err := config.LoadCfg(datakit.Cfg, datakit.MainConfPath); err != nil {
			l.Errorf("load config failed: %s", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	l = logger.SLogger("main")
}

func runDatakitWithHTTPServer() error {

	io.Start()

	if err := inputs.RunInputs(); err != nil {
		l.Error("error running inputs: %v", err)
		return err
	}

	http.Start(&http.Option{
		Bind:           datakit.Cfg.MainCfg.HTTPListen,
		GinLog:         datakit.Cfg.MainCfg.GinLog,
		GinReleaseMode: strings.ToLower(datakit.Cfg.MainCfg.LogLevel) != "debug",
		PProf:          datakit.Cfg.MainCfg.EnablePProf,
	})

	return nil
}

func runDatakitWithCmd() {
	if *flagCmd {
		l.Warn("--cmd parameter has been discarded")
		os.Exit(0)
	}

	if *flagPipeline != "" {
		cmds.PipelineDebugger(*flagPipeline, *flagText)
		os.Exit(0)
	}

	if *flagGrokq {
		cmds.Grokq()
		os.Exit(0)
	}

	if *flagMan {
		cmds.Man()
		os.Exit(0)
	}

	if *flagExportMan != "" {
		if err := cmds.ExportMan(*flagExportMan, *flagIgnoreMans); err != nil {
			l.Error(err)
		}
		os.Exit(0)
	}

	if *flagInstallExternal != "" {
		if err := cmds.InstallExternal(*flagInstallExternal); err != nil {
			l.Error(err)
		}
		os.Exit(0)
	}

	if *flagStart {
		cmds.StartDatakit()
		os.Exit(0)
	}

	if *flagStop {
		cmds.StopDatakit()
		os.Exit(0)
	}

	if *flagRestart {
		cmds.RestartDatakit()
		os.Exit(0)
	}

	if *flagReload {
		cmds.ReloadDatakit(*flagReloadPort)
		os.Exit(0)
	}
}

type datakitVerInfo struct {
	VersionString string `json:"version"`
	Commit        string `json:"commit"`
	ReleaseDate   string `json:"date_utc"`

	downloadURL        string `json:"-"`
	downloadURLTesting string `json:"-"`

	version *semver.Version
}

func (vi *datakitVerInfo) String() string {
	return fmt.Sprintf("datakit %s/%s", vi.VersionString, vi.Commit)
}

func (vi *datakitVerInfo) parse() error {
	verstr := strings.TrimPrefix(vi.VersionString, "v") // older version has prefix `v', this crash semver.Parse()
	v, err := semver.Parse(verstr)
	if err != nil {
		return err
	}
	vi.version = &v
	return nil
}

func getVersion(addr string) (*datakitVerInfo, error) {
	resp, err := nhttp.Get("http://" + path.Join(addr, "version"))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	infobody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ver datakitVerInfo
	if err = json.Unmarshal(infobody, &ver); err != nil {
		return nil, err
	}

	if err := ver.parse(); err != nil {
		return nil, err
	}
	ver.downloadURL = fmt.Sprintf("https://%s/installer-%s-%s",
		addr, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		ver.downloadURL += ".exe"
	}
	return &ver, nil
}

func getOnlineVersions() (res map[string]*datakitVerInfo, err error) {

	nhttp.DefaultTransport.(*nhttp.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	res = map[string]*datakitVerInfo{}

	onlineVer, err := getVersion("static.dataflux.cn/datakit")
	if err != nil {
		return nil, err
	}
	res["Online"] = onlineVer

	if *flagShowTestingVersions {
		testVer, err := getVersion("zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit")
		if err != nil {
			return nil, err
		}
		res["Testing"] = testVer
	}

	return
}

func getLocalVersion() (*datakitVerInfo, error) {
	v := &datakitVerInfo{VersionString: strings.TrimPrefix(git.Version, "v"), Commit: git.Commit, ReleaseDate: git.BuildAt}
	if err := v.parse(); err != nil {
		return nil, err
	}
	return v, nil
}

func isNewVersion(newVer, curver *datakitVerInfo, acceptRC bool) bool {

	if newVer.version.Compare(*curver.version) > 0 { // new version
		if len(newVer.version.Pre) == 0 {
			return true
		}

		if acceptRC {
			return true
		}
	}

	return false
}

func tryOTAUpdate(ver string) error {
	baseURL := "static.dataflux.cn/datakit"

	datakitUrl := "https://" + path.Join(baseURL,
		fmt.Sprintf("datakit-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, ver))

	dataUrl := "https://" + path.Join(baseURL, "data.tar.gz")

	l.Debugf("downloading %s to %s...", datakitUrl, datakit.InstallDir)
	if err := install.Download(datakitUrl, datakit.InstallDir, false, false); err != nil {
		return err
	}

	l.Debugf("downloading %s to %s...", dataUrl, datakit.InstallDir)
	if err := install.Download(dataUrl, datakit.InstallDir, false, false); err != nil {
		l.Errorf("download %s failed: %v, ignored", dataUrl, err)
	}

	svc, err := datakit.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return err
	}

	return install.UpgradeDatakit(svc)
}
