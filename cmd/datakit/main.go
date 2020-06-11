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

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/all"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/run"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/telegrafwrap"
)

var (
	flagVersion        = flag.Bool("version", false, `show verison info`)
	flagDataWay        = flag.String("dataway", ``, `dataway IP:Port`)
	flagCheckConfigDir = flag.Bool("check-config-dir", false, `check datakit conf.d, list configired and mis-configured collectors`)
	flagInputFilters   = flag.String("input-filter", "", "filter the inputs to enable, separator is :")
	flagListCollectors = flag.Bool("tree", false, `list vailable collectors`)
)

var (
	stopCh     chan struct{}
	stopFalgCh chan struct{}

	inputFilters = []string{}
)

func main() {

	flag.Parse()

	applyFlags()

	loadConfig()

	svcConfig := &service.Config{
		Name: config.ServiceName,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal("E! " + err.Error())
		return
	}

	log.Printf("I! starting datakit service")

	if err = s.Run(); err != nil {
		log.Fatalln(err.Error())
	}
}

func applyFlags() {

	if *flagVersion {
		fmt.Printf(`Version:        %s
Sha1:           %s
Build At:       %s
Golang Version: %s
Uploader:         %s
`, git.Version, git.Sha1, git.BuildAt, git.Golang, git.Uploader)
		os.Exit(0)
	}

	if *flagListCollectors {
		collectors := map[string][]string{}

		for k, v := range inputs.Inputs {
			cat := v().Catalog()
			collectors[cat] = append(collectors[cat], k)
		}

		for k, vs := range collectors {
			fmt.Println(k)
			for _, v := range vs {
				//fmt.Printf("  └── %s\n", v)
				fmt.Printf("  |--[d] %s\n", v)
			}
		}

		collectors = map[string][]string{}
		for k, v := range config.SupportsTelegrafMetricNames {
			collectors[v.Catalog] = append(collectors[v.Catalog], k)
		}

		for k, vs := range collectors {
			fmt.Println(k)
			for _, v := range vs {
				fmt.Printf("  |--[t] %s\n", v)
			}
		}

		os.Exit(0)
	}

	if *flagCheckConfigDir {
		config.CheckConfd()
		os.Exit(0)
	}

	if *flagInputFilters != "" {
		inputFilters = strings.Split(":"+strings.TrimSpace(*flagInputFilters)+":", ":")
	}
}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run(s)
	return nil
}

func (p *program) run(s service.Service) {
	stopCh = make(chan struct{})
	stopFalgCh = make(chan struct{})
	reloadLoop(stopCh)
}

func (p *program) Stop(s service.Service) error {
	close(stopCh)
	<-stopFalgCh //等待完整退出
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

		if err := runTelegraf(ctx); err != nil {
			log.Fatalf("E! fail to start sub service, %s", err)
		}

		if err := runDatakit(ctx); err != nil && err != context.Canceled {
			log.Fatalf("E! datakit abort: %s", err)
		}

		telegrafwrap.Svr.StopAgent()

		close(stopFalgCh)
	}
}

func loadConfig() {

	if err := config.LoadCfg(); err != nil {
		log.Fatalf("[error] load config failed: %s", err)
	}

	config.Cfg.InputFilters = inputFilters
	log.Printf("I! input fileters %v", inputFilters)

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
