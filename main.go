package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/run"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/telegrafwrap"

	"github.com/influxdata/telegraf/logger"
	winsvr "github.com/kardianos/service"
	uuid "github.com/satori/go.uuid"

	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/all"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"

	serviceutil "gitlab.jiagouyun.com/cloudcare-tools/cliutils/service"
)

var (
	workdir = "/usr/local/cloudcare/forethought/datakit/"

	flagVersion = flag.Bool("version", false, `show verison info`)

	flagInit      = flag.Bool(`init`, false, `init agent`)
	flagFtDataway = flag.String("ftdataway", ``, `address of ftdataway`)
	flagLogFile   = flag.String(`log`, ``, `log file`)
	flagLogLevel  = flag.String(`log-level`, ``, `log level`)

	flagUpgrade = flag.Bool(`upgrade`, false, `upgrade agent`)

	flagInstall     = flag.String(`install`, ``, `install datakit with systemctl or upstart`)
	flagInstallOnly = flag.Bool(`install-only`, false, `not run after installing`)

	flagCfgFile = flag.String("cfg", ``, `configure file`)
	flagCfgDir  = flag.String("config-dir", ``, `sub configuration dir`)

	fRunAsConsole = flag.Bool("console", false, "run as console application (windows only)")

	fService    = flag.String("service", "", "operate on the service (windows only)")
	fInstallDir = flag.String("installdir", `C:\Program Files (x86)\Forethought\DataFlux EBA Agent`, "install directory")
)

var (
	stop chan struct{}

	datakitConfig *config.Config
)

func main() {

	flag.Parse()

	if *flagVersion {
		fmt.Printf(`Version:        %s
Sha1:           %s
Build At:       %s
Golang Version: %s
`, git.Version, git.Sha1, git.BuildAt, git.Golang)
		return
	}

	if *flagInstall != "" {
		if err := doInstall(*flagInstall); err != nil {
			os.Exit(-1)
		}
		return
	}

	if *fService != "" && runtime.GOOS == "windows" {
		serviceCmd()
		return
	}

	exepath, err := os.Executable()
	if err != nil {
		log.Fatalln(err)
	}
	config.ExecutableDir = filepath.Dir(exepath)

	if *flagCfgFile == "" {
		*flagCfgFile = filepath.Join(config.ExecutableDir, fmt.Sprintf("%s.conf", config.ServiceName))
	}
	if *flagCfgDir == "" {
		*flagCfgDir = filepath.Join(config.ExecutableDir, "conf.d")
	}
	config.MainCfgPath = *flagCfgFile

	if *flagInit {
		if err := initialize(); err != nil {
			log.Fatalf("%s", err)
		}
		return
	} else if *flagUpgrade {

		if *flagCfgDir == "" {
			*flagCfgDir = filepath.Join(config.ExecutableDir, "conf.d")
		}

		config.InitTelegrafSamples()

		if err = config.CreatePluginConfigs(*flagCfgDir, true); err != nil {
			log.Fatalf("%s", err)
		}
		return
	}

	if runtime.GOOS == "windows" && windowsRunAsService() {

		svcConfig := &winsvr.Config{
			Name: config.ServiceName,
		}

		prg := &program{}
		s, err := winsvr.New(prg, svcConfig)
		if err != nil {
			log.Fatal("E! " + err.Error())
			return
		}

		err = s.Run()

		if err != nil {
			log.Fatalln(err.Error())
		}

	} else {
		stop = make(chan struct{})
		reloadLoop(stop)
	}
}

type program struct {
}

func (p *program) Start(s winsvr.Service) error {
	go p.run(s)
	return nil
}

func (p *program) run(s winsvr.Service) {
	stop = make(chan struct{})
	reloadLoop(stop)
}

func (p *program) Stop(s winsvr.Service) error {
	close(stop)
	return nil
}

func serviceCmd() {

	if runtime.GOOS != "windows" {
		return
	}

	svcConfig := &winsvr.Config{
		Name:        config.ServiceName,
		DisplayName: config.ServiceName,
		Description: "Collects data and publishes it to ftdataway.",
	}

	if *fService == "install" {
		if *fInstallDir == "" {
			log.Printf("installdir must not be empty")
			os.Exit(1)
			return
		}
		exepath := filepath.Join(*fInstallDir, `agent.exe`)
		_, err := os.Stat(exepath)
		if err != nil {
			log.Printf("executable file not found in %s", *fInstallDir)
			os.Exit(1)
			return
		}
		svcConfig.Executable = exepath
		svcConfig.Arguments = []string{"/config", filepath.Join(*fInstallDir, `agent.conf`)}
	}

	prg := &program{}
	s, err := winsvr.New(prg, svcConfig)
	if err != nil {
		log.Printf("Error starting service, %s ", err)
		os.Exit(1)
		return
	}

	if *fService == "status" {
		_, err := s.Status()
		if err != nil {
			log.Printf("Error get service status, %s", err)
			os.Exit(1)
			return
		}
		os.Exit(0)
		return
	}

	if *fService == "stop" || *fService == "uninstall" {
		_, err := s.Status()
		if err == winsvr.ErrNotInstalled {
			log.Printf("ok(service not found)")
			os.Exit(0)
			return
		}
	}

	err = winsvr.Control(s, *fService)
	if err != nil {
		log.Printf("E! %s", err.Error())
		os.Exit(1)
	}
	log.Println("Success!")
	os.Exit(0)
}

func initialize() error {
	if *flagLogFile == "" {
		*flagLogFile = filepath.Join(config.ExecutableDir, "datakit.log")
	}

	if *flagLogLevel == "" {
		*flagLogLevel = "info"
	}

	if *flagCfgDir == "" {
		*flagCfgDir = filepath.Join(config.ExecutableDir, "conf.d")
	}

	uid, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("Error creating uuid, %s", err)
	}

	maincfg := config.MainConfig{
		UUID:      "dkit_" + uid.String(),
		FtGateway: *flagFtDataway,
		Log:       *flagLogFile,
		LogLevel:  *flagLogLevel,
		ConfigDir: *flagCfgDir,
	}

	if err = config.InitMainCfg(&maincfg, config.MainCfgPath); err != nil {
		return err
	}

	config.InitTelegrafSamples()
	return config.CreatePluginConfigs(*flagCfgDir, false)
}

func reloadLoop(stop chan struct{}) {
	reload := make(chan bool, 1)
	reload <- true
	for <-reload {
		reload <- false

		ctx, cancel := context.WithCancel(context.Background())

		signals := make(chan os.Signal)
		signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
			syscall.SIGTERM, syscall.SIGINT)
		go func() {
			select {
			case sig := <-signals:
				if sig == syscall.SIGHUP {
					log.Printf("I! Reloading Telegraf config")
					<-reload
					reload <- true
				}
				cancel()
			case <-stop:
				cancel()
			}
		}()

		err := loadConfig(ctx)
		if err == nil {
			err = runTelegraf(ctx)
		}
		if err == nil {
			select {
			case <-ctx.Done():
				telegrafwrap.Svr.StopAgent()
				break
			default:
			}
			err = runAgent(ctx)
		}
		if err != nil && err != context.Canceled {
			log.Fatalf("E! [datakit] %s", err)
		}
	}
}

func loadConfig(ctx context.Context) error {
	c := config.NewConfig()

	err := c.LoadConfig(ctx, config.MainCfgPath)
	if err != nil {
		return err
	}
	datakitConfig = c

	logConfig := logger.LogConfig{
		Debug:     (strings.ToLower(datakitConfig.MainCfg.LogLevel) == "debug"),
		Quiet:     false,
		LogTarget: logger.LogTargetFile,
		Logfile:   datakitConfig.MainCfg.Log,
	}
	logConfig.RotationMaxSize.Size = (20 << 10 << 10)
	logger.SetupLogging(logConfig)

	log.Printf("%s v%s", config.ServiceName, git.Version)

	return nil
}

func runTelegraf(ctx context.Context) error {
	telegrafwrap.Svr.MainCfg = datakitConfig.MainCfg
	return telegrafwrap.Svr.Start(ctx)
}

func runAgent(ctx context.Context) error {

	pnames := []string{}
	for _, ip := range datakitConfig.Inputs {
		pnames = append(pnames, ip.Config.Name)
	}
	log.Printf("[agent] avariable inputs: %s", strings.Join(pnames, ","))

	pnames = pnames[:0]
	for _, op := range datakitConfig.Outputs {
		pnames = append(pnames, op.Config.Name)
	}
	log.Printf("[agent] avariable outputs: %s", strings.Join(pnames, ","))

	ag, err := run.NewAgent(datakitConfig)
	if err != nil {
		return err
	}

	return ag.Run(ctx)
}

func windowsRunAsService() bool {
	if *fRunAsConsole {
		return false
	}

	return !winsvr.Interactive()
}

func doInstall(serviceType string) error {

	svr := &serviceutil.Service{
		Name:        config.ServiceName,
		InstallDir:  workdir,
		Description: `Forethought Datakit`,
		StartCMD:    fmt.Sprintf("%s -cfg=%s", filepath.Join(workdir, `datakit`), *flagCfgFile),
		Type:        serviceType,
	}

	if *flagInstallOnly {
		svr.InstallOnly = true
	}

	return svr.Install()
}
