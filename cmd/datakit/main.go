package main

import (
	"flag"
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
)

var (
	flagVersion        = flag.Bool("version", false, `show version info`)
	flagCheckConfigDir = flag.Bool("check-config-dir", false, `check datakit conf.d, list configired and mis-configured collectors`)
	flagInputFilters   = flag.String("input-filter", "", "filter the inputs to enable, separator is :")
	flagDocker         = flag.Bool("docker", false, "run within docker")

	flagListCollectors    = flag.Bool("tree", false, `list vailable collectors`)
	flagListConfigSamples = flag.Bool("config-samples", false, `list all config samples`)
)

var (
	inputFilters = []string{}
	l            = logger.DefaultSLogger("main")
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

	if *flagListConfigSamples {
		showAllConfigSamples()
		os.Exit(0)
	}

	if *flagCheckConfigDir {
		config.CheckConfd()
		os.Exit(0)
	}

	if *flagListCollectors {
		listCollectors()
		os.Exit(0)
	}

	if *flagInputFilters != "" {
		inputFilters = strings.Split(":"+strings.TrimSpace(*flagInputFilters)+":", ":")
	}

	if *flagDocker {
		datakit.Docker = true
	}
}

func listCollectors() {
	collectors := map[string][]string{}

	for k, v := range inputs.Inputs {
		cat := v().Catalog()
		collectors[cat] = append(collectors[cat], k)
	}

	ndk := 0
	for k, vs := range collectors {
		fmt.Println(k)
		for _, v := range vs {
			fmt.Printf("  |--[d] %s\n", v)
			ndk++
		}
	}

	collectors = map[string][]string{}
	for k, v := range tgi.TelegrafInputs {
		collectors[v.Catalog] = append(collectors[v.Catalog], k)
	}

	ntg := 0
	for k, vs := range collectors {
		fmt.Println(k)
		for _, v := range vs {
			fmt.Printf("  |--[t] %s\n", v)
			ntg++
		}
	}

	fmt.Printf("total %d, datakit: %d, telegraf: %d\n", ntg+ndk, ndk, ntg)
}

func showAllConfigSamples() {
	for k, v := range inputs.Inputs {
		fmt.Printf("#========= [D] % 32s ==========\n%s\n", k, v().SampleConfig())
	}

	for k, v := range tgi.TelegrafInputs {
		fmt.Printf("#========= [T] % 32s ==========\n%s\n", k, v.SampleConfig())
	}
}

func run() {

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

	case <-datakit.GlobalExit.Wait():
		l.Debug("datakit exit on sem")
	}

	l.Info("datakit exit.")
}

func tryLoadConfig() {
	datakit.Cfg.InputFilters = inputFilters

	for {
		if err := config.LoadCfg(datakit.Cfg); err != nil {
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
