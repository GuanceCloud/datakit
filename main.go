package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/run"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/telegrafwrap"

	"github.com/influxdata/telegraf/logger"
	winsvr "github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/all"
)

var (
	workdir = "/usr/local/cloudcare/dataflux/datakit/"

	flagVersion = flag.Bool("version", false, `show verison info`)

	flagInit         = flag.Bool(`init`, false, `init agent`)
	flagDataway      = flag.String("dataway", ``, `address of ftdataway`)
	flagAgentLogFile = flag.String(`agent-log`, ``, `agent log file`)

	flagDocker = flag.Bool(`docker`, false, `run in docker`)

	flagUpgrade = flag.Bool(`upgrade`, false, `upgrade agent`)

	flagInstallOnly = flag.Bool(`install-only`, false, `not run after installing`)

	flagCfgFile = flag.String("cfg", ``, `configure file`)
	flagCfgDir  = flag.String("config-dir", ``, `sub configuration dir`)

	flagCheckConfigDir = flag.Bool("check-config-dir", false, `check datakit conf.d, list configired and mis-configured collectors`)

	fInstallDir = flag.String("installdir", `C:\Program Files (x86)\DataFlux\DataKit`, "install directory")

	fInputFilters = flag.String("input-filter", "", "filter the inputs to enable, separator is :")

	flagListCollectors = flag.Bool("list-collectors", false, `list vailable collectors`)
)

var (
	winStopCh     chan struct{}
	winStopFalgCh chan struct{}

	inputFilters = []string{}
)

func main() {

	flag.Parse()
	args := flag.Args()

	if *flagVersion || (len(args) > 0 && args[0] == "version") {
		fmt.Printf(`Version:        %s
Sha1:           %s
Build At:       %s
Golang Version: %s
`, git.Version, git.Sha1, git.BuildAt, git.Golang)
		return
	}

	if *flagListCollectors {
		for k, _ := range inputs.Inputs {
			fmt.Println(k)
		}
		return
	}

	if *flagCheckConfigDir {
		config.CheckConfd(*flagCfgDir)
		return
	}

	applyFlags()

	svcConfig := &winsvr.Config{
		Name: config.ServiceName,
	}

	prg := &program{}
	s, err := winsvr.New(prg, svcConfig)
	if err != nil {
		log.Fatal("E! " + err.Error())
		return
	}

	if err = s.Run(); err != nil {
		log.Fatalln(err.Error())
	}
}

func applyFlags() {
	//exepath, err := os.Executable()
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//config.ExecutableDir = filepath.Dir(exepath)

	//if *flagCfgFile == "" {
	//	*flagCfgFile = filepath.Join(config.ExecutableDir, fmt.Sprintf("%s.conf", config.ServiceName))
	//}

	//if *flagCfgDir == "" {
	//	*flagCfgDir = filepath.Join(config.ExecutableDir, "conf.d")
	//}
	//config.MainCfgPath = *flagCfgFile

	//if *flagAgentLogFile != "" {
	//	config.AgentLogFile = *flagAgentLogFile
	//}

	if *fInputFilters != "" {
		inputFilters = strings.Split(":"+strings.TrimSpace(*fInputFilters)+":", ":")
	}

	/*
		if *flagUpgrade {
			config.InitTelegrafSamples()

			if err = config.CreatePluginConfigs(*flagCfgDir, true); err != nil {
				log.Fatalf("%s", err)
			}
			return
		} */
}

type program struct {
}

func (p *program) Start(s winsvr.Service) error {
	go p.run(s)
	return nil
}

func (p *program) run(s winsvr.Service) {
	winStopCh = make(chan struct{})
	winStopFalgCh = make(chan struct{})
	reloadLoop(winStopCh)
}

func (p *program) Stop(s winsvr.Service) error {
	close(winStopCh)
	<-winStopFalgCh //等待完整退出
	return nil
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
					log.Printf("Reloading config")
					<-reload
					reload <- true
				}
				log.Printf("signal notify: %v", sig)
				cancel()
			case <-stop:
				log.Printf("service stopped")
				cancel()
			}
		}()

		loadConfig(ctx)

		if err := runTelegraf(ctx); err != nil {
			log.Fatalf("E! fail to start sub service, %s", err)
		}

		if err := runDatakit(ctx); err != nil && err != context.Canceled {
			log.Fatalf("E! datakit abort: %s", err)
		}

		telegrafwrap.Svr.StopAgent()

		close(winStopFalgCh)
	}
}

func loadConfig(ctx context.Context) {

	logConfig := logger.LogConfig{
		Debug:     (strings.ToLower(config.Cfg.MainCfg.LogLevel) == "debug"),
		Quiet:     false,
		LogTarget: logger.LogTargetFile,
		Logfile:   config.Cfg.MainCfg.Log,
	}

	logConfig.RotationMaxSize.Size = (20 << 10 << 10)
	logger.SetupLogging(logConfig)

	config.Cfg.InputFilters = inputFilters
	if err := config.LoadCfg(ctx); err != nil {
		log.Fatalf("[error] load config failed: %s", err)
	}
}

func runTelegraf(ctx context.Context) error {
	telegrafwrap.Svr.Cfg = config.Cfg
	return telegrafwrap.Svr.Start(ctx)
}

func runDatakit(ctx context.Context) error {

	ag, err := run.NewAgent(config.Cfg)
	if err != nil {
		return err
	}

	return ag.Run(ctx)
}
