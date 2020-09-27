package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/install"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	DataKitBaseURL = ""
	DataKitVersion = ""

	datakitUrl = "https://" + path.Join(DataKitBaseURL,
		fmt.Sprintf("datakit-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion))

	telegrafUrl = "https://" + path.Join(DataKitBaseURL,
		"telegraf",
		fmt.Sprintf("agent-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH))

	l *logger.Logger
)

var (
	flagUpgrade     = flag.Bool("upgrade", false, ``)
	flagDatawayHTTP = flag.String("dataway", "", `address of dataway(http://IP:Port?token=xxx), port default 9528`)
	flagDatawayWS   = flag.String("dataway-ws", "", `address of dataway websocket(ws://IP:Port?token=xxx), port default 9531`)

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
	lopt := logger.OPT_DEFAULT | logger.OPT_STDOUT
	if runtime.GOOS != "windows" { // disable color on windows(some color not working under windows)
		lopt |= logger.OPT_COLOR
	}

	logger.SetGlobalRootLogger("", logger.DEBUG, lopt)
	l = logger.SLogger("installer")

	flag.Parse()
	datakit.InitDirs()
	applyFlags()

	// create install dir if not exists
	if err := os.MkdirAll(install.InstallDir, 0775); err != nil {
		l.Fatal(err)
	}

	datakit.ServiceExecutable = filepath.Join(install.InstallDir, datakitBin)
	if runtime.GOOS == datakit.OSWindows {
		datakit.ServiceExecutable += ".exe"
	}

	svc, err := datakit.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	l.Info("stoping datakit...")
	_ = install.StopDataKitService(svc) // stop service if installed before

	if *flagOffline && *flagSrcs != "" {
		for _, f := range strings.Split(*flagSrcs, ",") {
			install.ExtractDatakit(f, install.InstallDir)
		}
	} else {
		install.CurDownloading = dlDatakit
		install.Download(datakitUrl, install.InstallDir)

		install.CurDownloading = dlAgent
		install.Download(telegrafUrl, install.InstallDir)
	}

	if *flagUpgrade { // upgrade new version
		l.Infof("Upgrading to version %s...", DataKitVersion)
		install.UpgradeDatakit(svc)
	} else { // install new datakit
		l.Infof("Installing version %s...", DataKitVersion)
		install.InstallNewDatakit(svc)
	}

	l.Infof("starting service %s...", datakit.ServiceName)
	if err = service.Control(svc, "start"); err != nil {
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
		install.DownloadOnly = true

		install.CurDownloading = dlDatakit
		install.Download(datakitUrl, fmt.Sprintf("datakit-%s-%s-%s.tar.gz",
			runtime.GOOS, runtime.GOARCH, DataKitVersion))

		install.CurDownloading = dlAgent
		install.Download(telegrafUrl, fmt.Sprintf("agent-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH))

		os.Exit(0)
	}

	switch install.OSArch {

	case datakit.OSArchWinAmd64:
		install.InstallDir = `C:\Program Files\dataflux\datakit`

	case datakit.OSArchWin386:
		install.InstallDir = `C:\Program Files (x86)\dataflux\datakit`

	case datakit.OSArchLinuxArm,
		datakit.OSArchLinuxArm64,
		datakit.OSArchLinux386,
		datakit.OSArchLinuxAmd64,
		datakit.OSArchDarwinAmd64:
		install.InstallDir = `/usr/local/cloudcare/dataflux/datakit`

	default:
		// TODO: more os/arch support
	}

	install.DataWayHTTP = *flagDatawayHTTP
	install.DataWayWs = *flagDatawayWS
	install.GlobalTags = *flagGlobalTags
	install.Port = *flagPort
	install.DatakitName = *flagDatakitName
	install.EnableInputs = *flagEnableInputs
}
