package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	nhttp "net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/blang/semver/v4"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
)

var (
	flagVersion = flag.Bool("version", false, `show version info`)
	flagDocker  = flag.Bool("docker", false, "run within docker")

	// tool-commands supported in datakit
	flagCmd             = flag.Bool("cmd", false, "run datakit under command line mode")
	flagPipeline        = flag.String("pl", "", "pipeline script to test(name only, do not use file path)")
	flagText            = flag.String("txt", "", "text string for the pipeline or grok(json or raw text)")
	flagGrokq           = flag.Bool("grokq", false, "query groks interactively")
	flagMan             = flag.Bool("man", false, "read manuals of inputs")
	flagCheckUpdate     = flag.Bool("check-update", false, "update datakit new version if available")
	flagAcceptRCVersion = flag.Bool("accept-rc-version", false, "accept RC version if available")
	flagExportMan       = flag.String("export-man", "", "export all inputs and related manuals to specified path")
)

var (
	l = logger.DefaultSLogger("main")

	ReleaseType = ""
)

func main() {
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
		ver, err := getOnlineVersion()
		if err != nil {
			fmt.Printf("Get online version failed: \n%s\n", err.Error())
			os.Exit(-1)
		}
		curver, err := getLocalVersion()
		if err != nil {
			fmt.Printf("Get online version failed: \n%s\n", err.Error())
			os.Exit(-1)
		}

		if isNewVersion(ver, curver, true) {
			fmt.Printf("\n\nNew version available: %s, commit %s (release at %s)\n",
				ver.version, ver.Commit, ver.ReleaseDate)
			switch runtime.GOOS {
			case "windows":
				cmdWin := fmt.Sprintf(winUpgradeCmd, ver.downloadURL)
				fmt.Printf("\nUpgrade:\n\t%s\n\n", cmdWin)
			default:
				cmd := fmt.Sprintf(unixUpgradeCmd, ver.downloadURL)
				fmt.Printf("\nUpgrade:\n\t%s\n\n", cmd)
			}
		}

		os.Exit(0)
	}

	if *flagCheckUpdate {

		logger.SetGlobalRootLogger(datakit.OTALogFile, logger.DEBUG, logger.OPT_DEFAULT)
		l = logger.SLogger("ota")

		ver, err := getOnlineVersion()
		if err != nil {
			l.Errorf("Get online version failed: \n%s\n", err.Error())
			os.Exit(0)
		}
		curver, err := getLocalVersion()
		if err != nil {
			l.Errorf("Get online version failed: \n%s\n", err.Error())
			os.Exit(-1)
		}

		if isNewVersion(ver, curver, *flagAcceptRCVersion) {
			l.Infof("New version available: %s, commit %s (release at %s)",
				ver.version, ver.Commit, ver.ReleaseDate)
			os.Exit(-1)
		}

		os.Exit(0)
	}

	datakit.EnableUncheckInputs = (ReleaseType == "all")

	if *flagCmd {
		runDatakitWithCmd()
		os.Exit(0)
	}

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

	for k, v := range tgi.TelegrafInputs {
		sample := v.SampleConfig()
		if err := ioutil.WriteFile(filepath.Join(fpath, k+".conf"), []byte(sample), os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func run() {

	if datakit.Cfg.MainCfg.EnablePProf {
		go func() {
			if err := nhttp.ListenAndServe(":6060", nil); err != nil {
				l.Fatalf("pprof server error: %s", err.Error())
			}
		}()
	}

	inputs.StartTelegraf()

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
			datakit.Quit()
		}

	case <-datakit.StopCh:
		l.Infof("service stopping")
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
	go func() {
		http.Start(datakit.Cfg.MainCfg.HTTPBind)
	}()

	return nil
}

func runDatakitWithCmd() {
	if *flagPipeline != "" {
		cmds.PipelineDebugger(*flagPipeline, *flagText)
		return
	}

	if *flagGrokq {
		cmds.Grokq()
		return
	}

	if *flagMan {
		cmds.Man()
		return
	}

	if *flagExportMan != "" {
		if err := cmds.ExportMan(*flagExportMan); err != nil {
			l.Error(err)
		}
		return
	}
}

type datakitVerInfo struct {
	VersionString string `json:"version"`
	Commit        string `json:"commit"`
	ReleaseDate   string `json:"date_utc"`
	downloadURL   string `json:"-"`

	version *semver.Version
}

func (vi *datakitVerInfo) String() string {
	return fmt.Sprintf("datakit %s/%s", vi.VersionString, vi.Commit)
}

func (vi *datakitVerInfo) parse() error {
	vi.VersionString = strings.TrimPrefix(vi.VersionString, "v") // older version has prefix `v'
	v, err := semver.Parse(vi.VersionString)
	if err != nil {
		return err
	}
	vi.version = &v
	return nil
}

func getOnlineVersion() (*datakitVerInfo, error) {
	nhttp.DefaultTransport.(*nhttp.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := nhttp.Get("https://static.dataflux.cn/datakit/version")
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

	ver.downloadURL = fmt.Sprintf("https://static.dataflux.cn/datakit/installer-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		ver.downloadURL += ".exe"
	}
	return &ver, nil
}

func getLocalVersion() (*datakitVerInfo, error) {
	v := &datakitVerInfo{VersionString: strings.TrimPrefix(git.Version, "v"), Commit: git.Commit, ReleaseDate: git.BuildAt}
	if err := v.parse(); err != nil {
		return nil, err
	}
	return v, nil
}

func isNewVersion(newVer, curVer *datakitVerInfo, acceptRC bool) bool {

	if newVer.version.Compare(*curVer.version) > 0 { // new version
		if len(newVer.version.Pre) == 0 {
			return true
		}

		if acceptRC {
			return true
		}
	}

	return false
}
