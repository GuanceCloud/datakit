package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/install"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

var (
	DataKitBaseURL = ""
	DataKitVersion = ""

	oldInstallDir      = "/usr/local/cloudcare/dataflux/datakit"
	oldInstallDirWin   = `C:\Program Files\dataflux\datakit`
	oldInstallDirWin32 = `C:\Program Files (x86)\dataflux\datakit`

	datakitUrl = "https://" + path.Join(DataKitBaseURL,
		fmt.Sprintf("datakit-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion))

	dataUrl = "https://" + path.Join(DataKitBaseURL, "data.tar.gz")

	l = logger.DefaultSLogger("installer")
)

var (
	flagUpgrade     = flag.Bool("upgrade", false, ``)
	flagDatawayHTTP = flag.String("dataway", "", `address of dataway(http://IP:Port?token=xxx), port default 9528`)

	flagInfo         = flag.Bool("info", false, "show installer info")
	flagDownloadOnly = flag.Bool("download-only", false, `download datakit only, not install`)
	flagInstallOnly  = flag.Bool("install-only", false, "install only, not start")

	flagEnableInputs = flag.String("enable-inputs", "", `default enable inputs(comma splited, example: cpu,mem,disk)`)
	flagDatakitName  = flag.String("name", "", `specify DataKit name, example: prod-env-datakit`)
	flagGlobalTags   = flag.String("global-tags", "", `enable global tags, example: host=__datakit_hostname,ip=__datakit_ip`)

	flagPort      = flag.Int("port", 9529, "datakit HTTP port")
	flagListen    = flag.String("listen", "localhost", "datakit HTTP listen")
	flagNamespace = flag.String("namespace", "", "datakit namespace")

	flagInstallLog    = flag.String("install-log", "", "install log")
	flagOTA           = flag.Bool("ota", false, "auto update")
	flagCloudProvider = flag.String("cloud-provider", "", "specify cloud provider(accept aliyun/tencent/aws)")

	flagOffline = flag.Bool("offline", false, "offline install mode")
	flagSrcs    = flag.String("srcs", fmt.Sprintf("./datakit-%s-%s-%s.tar.gz,./data.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion), `local path of datakit and agent install files`)

	flagPreserveEnvRegex = flag.String("preserve-env-regex", "", "preserve existing environment variables")
)

const (
	datakitBin = "datakit"
	dlDatakit  = "datakit"
	dlData     = "data"
)

func mvOldDatakit(svc service.Service) {
	olddir := oldInstallDir
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case datakit.OSArchWinAmd64:
		olddir = oldInstallDirWin
	case datakit.OSArchWin386:
		olddir = oldInstallDirWin32
	}

	if _, err := os.Stat(olddir); err != nil {
		l.Debugf("deprecated install path %s not exists, ingored", olddir)
		return
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service datakit failed: %s, ignored", err.Error())
	}

	if err := os.Rename(olddir, datakit.InstallDir); err != nil {
		l.Fatalf("move %s -> %s failed: %s", olddir, datakit.InstallDir, err.Error())
	}
}

func filterEnvs(regex string) (res map[string]string) {
	re, err := regexp.Compile(regex)
	if err != nil {
		l.Fatalf("invalid regex: %s", err.Error())
	}

	res = map[string]string{}

	envs := os.Environ()
	for _, env := range envs {
		if re.MatchString(env) {
			l.Debugf("get env %s", env)
			arr := strings.SplitN(env, "=", 2)
			if len(arr) != 2 {
				continue
			}
			res[arr[0]] = arr[1]
		}
	}
	return
}

func main() {

	flag.Parse()

	if *flagOTA {
		install.OTA = true
	}

	if *flagInstallLog == "" {
		lopt := logger.OPT_DEFAULT | logger.OPT_STDOUT
		if runtime.GOOS != "windows" { // disable color on windows(some color not working under windows)
			lopt |= logger.OPT_COLOR
		}

		if err := logger.SetGlobalRootLogger("", logger.DEBUG, lopt); err != nil {
			l.Warnf("set root log failed: %s", err.Error())
		}
	} else {
		l.Infof("set log file to %s", *flagInstallLog)
		if err := logger.SetGlobalRootLogger(*flagInstallLog, logger.DEBUG, logger.OPT_DEFAULT); err != nil {
			l.Errorf("set root log failed: %s", err.Error())
		}
		install.Init()
	}

	l = logger.SLogger("installer")

	if *flagPreserveEnvRegex != "" {
		dkservice.ServiceEnvs = filterEnvs(*flagPreserveEnvRegex)
		l.Infof("added %d envs", len(dkservice.ServiceEnvs))
	}

	dkservice.ServiceExecutable = filepath.Join(datakit.InstallDir, datakitBin)
	if runtime.GOOS == datakit.OSWindows {
		dkservice.ServiceExecutable += ".exe"
	}

	svc, err := dkservice.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	l.Info("stoping datakit...")
	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}

	// 迁移老版本 datakit 数据目录
	mvOldDatakit(svc)

	config.InitDirs()
	applyFlags()

	// create install dir if not exists
	if err := os.MkdirAll(datakit.InstallDir, 0775); err != nil {
		l.Fatal(err)
	}

	if *flagOffline && *flagSrcs != "" {
		for _, f := range strings.Split(*flagSrcs, ",") {
			_ = install.ExtractDatakit(f, datakit.InstallDir)
		}
	} else {
		install.CurDownloading = dlDatakit
		if err := install.Download(datakitUrl, datakit.InstallDir, true, false); err != nil {
			return
		}

		fmt.Printf("\n")

		install.CurDownloading = dlData
		if err := install.Download(dataUrl, datakit.InstallDir, true, false); err != nil {
			return
		}
		fmt.Printf("\n")
	}

	if *flagUpgrade { // upgrade new version
		l.Infof("Upgrading to version %s...", DataKitVersion)
		if err := install.UpgradeDatakit(svc); err != nil {
			l.Fatalf("upgrade datakit: %s, ignored", err.Error())
		}
	} else { // install new datakit
		l.Infof("Installing version %s...", DataKitVersion)
		install.InstallNewDatakit(svc)
	}

	if !*flagInstallOnly {
		l.Infof("starting service %s...", dkservice.ServiceName)
		if err = service.Control(svc, "start"); err != nil {
			l.Warnf("star service: %s, ignored", err.Error())
		}
	}

	config.CreateSymlinks()

	if *flagUpgrade { // upgrade new version
		l.Info(":) Upgrade Success!")
	} else {
		l.Info(":) Install Success!")
	}

	promptReferences()
}

func promptReferences() {
	fmt.Printf("\n\tVisit http://localhost:%d/man/changelog to see DataKit change logs.\n", *flagPort)
	fmt.Printf("\tVisit http://localhost:%d/monitor to see DataKit running status.\n", *flagPort)
	fmt.Printf("\tVisit http://localhost:%d/man to see DataKit manuals.\n\n", *flagPort)
}

func applyFlags() {

	if *flagInfo {
		fmt.Printf(`
       Version: %s
      Build At: %s
Golang Version: %s
       BaseUrl: %s
       DataKit: %s
`, datakit.Version, git.BuildAt, git.Golang, DataKitBaseURL, datakitUrl)
		os.Exit(0)
	}

	if *flagDownloadOnly {
		install.DownloadOnly = true

		install.CurDownloading = dlDatakit

		if err := install.Download(datakitUrl,
			fmt.Sprintf("datakit-%s-%s-%s.tar.gz",
				runtime.GOOS, runtime.GOARCH, DataKitVersion), true, true); err != nil {
			return
		}
		fmt.Printf("\n")

		install.CurDownloading = dlData
		if err := install.Download(dataUrl, "data.tar.gz", true, true); err != nil {
			return
		}

		fmt.Printf("\n")

		os.Exit(0)
	}

	install.DataWayHTTP = *flagDatawayHTTP
	install.GlobalTags = *flagGlobalTags
	install.Port = *flagPort
	install.Listen = *flagListen
	install.CloudProvider = *flagCloudProvider
	install.DatakitName = *flagDatakitName
	install.EnableInputs = *flagEnableInputs
	install.Namespace = *flagNamespace
}
