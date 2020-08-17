package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
)

var (
	flagVersion        = flag.Bool("version", false, `show verison info`)
	flagDataWay        = flag.String("dataway", ``, `dataway IP:Port`)
	flagCheckConfigDir = flag.Bool("check-config-dir", false, `check datakit conf.d, list configired and mis-configured collectors`)
	flagInputFilters   = flag.String("input-filter", "", "filter the inputs to enable, separator is :")

	flagListCollectors    = flag.Bool("tree", false, `list vailable collectors`)
	flagListConfigSamples = flag.Bool("config-samples", false, `list all config samples`)
)

var (
	stopCh     = make(chan interface{})
	waitExitCh = make(chan interface{})

	waithttpStopCh = make(chan interface{})

	inputFilters = []string{}
	l            *logger.Logger
)

func main() {

	logger.SetStdoutRootLogger(logger.DEBUG, logger.OPT_DEFAULT)
	l = logger.SLogger("main")

	flag.Parse()

	applyFlags()

	loadConfig()

	svcConfig := &service.Config{
		Name: datakit.ServiceName,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		l.Fatal(err)
		return
	}

	l.Info("starting datakit service")

	if err = s.Run(); err != nil {
		l.Fatal(err)
	}
}

func applyFlags() {

	if *flagVersion {
		fmt.Printf(`
       Version: %s
        Commit: %s
        Branch: %s
 Build At(UTC): %s
Golang Version: %s
      Uploader: %s
`, git.Version, git.Commit, git.Branch, git.BuildAt, git.Golang, git.Uploader)
		os.Exit(0)
	}

	if *flagListCollectors {
		showAllCollectors()
		os.Exit(0)
	}

	if *flagListConfigSamples {
		showAllConfigSamples()
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

func showAllCollectors() {
	collectors := map[string][]string{}

	for k, v := range inputs.Inputs {
		cat := v().Catalog()
		collectors[cat] = append(collectors[cat], k)
	}

	ndatakit := 0
	for k, vs := range collectors {
		for _, v := range vs {
			fmt.Printf("[d][% 12s] %s\n", k, v)
			ndatakit++
		}
	}

	nagent := 0
	collectors = map[string][]string{}
	for k, v := range config.TelegrafInputs {
		collectors[v.Catalog] = append(collectors[v.Catalog], k)
	}

	for k, vs := range collectors {
		for _, v := range vs {
			fmt.Printf("[t][% 12s] %s\n", k, v)
			nagent++
		}
	}

	fmt.Println("===================================")
	fmt.Printf("total: %d, datakit: %d, agent: %d\n", ndatakit+nagent, ndatakit, nagent)
}

func showAllConfigSamples() {
	for k, v := range inputs.Inputs {
		sample := v().SampleConfig()
		fmt.Printf("%s\n========= [D] ==========\n%s\n", k, sample)
	}

	for k, v := range config.TelegrafInputs {
		fmt.Printf("%s\n========= [T] ==========\n%s\n", k, v.Sample)
	}
}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run(s)
	return nil
}

func (p *program) run(s service.Service) {
	__run()
}

func (p *program) Stop(s service.Service) error {
	close(stopCh)

	// We must wait here:
	// On windows, we stop datakit in services.msc, if datakit process do not
	// echo to here, services.msc will complain the datakit process has been
	// exit unexpected
	<-waitExitCh

	return nil
}

func exitDatakit() {
	datakit.Exit.Close()

	l.Info("wait all goroutines exit...")
	datakit.WG.Wait()

	l.Info("closing waitExitCh...")
	close(waitExitCh)
}

func __run() {

	inputs.StartTelegraf()

	l.Info("datakit start...")
	if err := runDatakitWithHTTPServer(); err != nil && err != context.Canceled {
		l.Fatalf("datakit abort: %s", err)
	}

	l.Info("datakit start ok. Wait signal or service stop...")

	// NOTE:
	// Actually, the datakit process been managed by system service, no matter on
	// windows/UNIX, datakit should exit via `service-stop' operation, so the signal
	// branch should not reached, but for daily debugging(ctrl-c), we kept the signal
	// exit option.
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-signals:
		if sig == syscall.SIGHUP {
			// TODO: reload configures
		} else {
			l.Infof("get signal %v, wait & exit", sig)
			exitDatakit()
		}
	case <-stopCh:
		l.Infof("service stopping")
		exitDatakit()
	case <-datakit.GlobalExit.Wait():
		l.Debug("datakit exit on sem")
	}

	l.Info("datakit exit.")
}

func loadConfig() {
	config.Cfg.InputFilters = inputFilters

	for {
		if err := config.LoadCfg(); err != nil {
			l.Errorf("load config failed: %s", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	l = logger.SLogger("main")
}

func runDatakitWithHTTPServer() error {

	l = logger.SLogger("datakit")
	io.Start()

	go func() {
		http.Start(config.Cfg.MainCfg.HTTPBind)
	}()

	if err := inputs.RunInputs(); err != nil {
		l.Error("error running inputs: %v", err)
	}

	return nil
}
