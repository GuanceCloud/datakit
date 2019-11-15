package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/satori/go.uuid"

	"github.com/siddontang/go-log/log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/config/all"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/uploader"
)

var (
	flagVersion = flag.Bool("version", false, `show verison info`)
	flagInit    = flag.Bool(`init`, false, `init agent`)

	flagFtGateway = flag.String("ftdataway", ``, ``)

	flagCfgFile = flag.String("config", ``, `configure file`)
	flagCfgDir  = flag.String("sub-config-dir", ``, `sub configuration dir`)

	flagLogFile  = flag.String(`log`, ``, `log file`)
	flagLogLevel = flag.String(`log-level`, ``, `log level`)
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

	exepath, err := os.Executable()
	if err != nil {
		log.Fatalln(err)
	}
	config.ExecutableDir = filepath.Dir(exepath)

	if *flagCfgFile == "" {
		*flagCfgFile = filepath.Join(config.ExecutableDir, "datakit.conf")
	}
	config.CfgPath = *flagCfgFile

	if *flagInit {

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
			log.Fatalln(err)
		}
		config.Cfg.UUID = "dkit_" + uid.String()
		config.Cfg.FtGateway = *flagFtGateway
		config.Cfg.Log = *flagLogFile
		config.Cfg.LogLevel = *flagLogLevel
		config.Cfg.ConfigDir = *flagCfgDir

		if err = config.InitializeConfigs(); err != nil {
			log.Fatalf("intialize configs error: %s", err.Error())
		}

		return
	}

	if err := config.LoadConfig(config.CfgPath); err != nil {
		log.Fatalf("[error] load config failed: %s", err.Error())
		return
	}

	logpath := config.Cfg.Log
	loglevel := config.Cfg.LogLevel
	if *flagLogFile != "" {
		logpath = *flagLogFile
	}
	if *flagLogLevel != "" {
		loglevel = *flagLogLevel
	}
	if loglevel == "" {
		loglevel = "info"
	}

	config.Cfg.Log = logpath
	config.Cfg.LogLevel = loglevel

	var gLogger *log.Logger
	if gLogger, err = setupLog(logpath, loglevel); err != nil {
		log.Fatalf("[error] %s", err)
		return
	}

	if config.Cfg.FtGateway == "" {
		log.Errorln("ftdateway required")
		return
	}

	subcfgdir := config.Cfg.ConfigDir
	if *flagCfgDir != "" {
		subcfgdir = *flagCfgDir
	}
	if subcfgdir == "" {
		subcfgdir = filepath.Join(config.ExecutableDir, "config.d")
	}
	config.Cfg.ConfigDir = subcfgdir

	if err = config.LoadSubConfigs(subcfgdir); err != nil {
		log.Fatalf("%s", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		select {
		case sig := <-signals:
			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				cancel()
			}
		}
	}()

	var wg sync.WaitGroup

	up := uploader.New(config.Cfg.FtGateway)
	up.Start()
	defer up.Stop()

	for _, svrCreator := range service.Services {
		wg.Add(1)
		svr := svrCreator(gLogger)
		go func(s service.Service) {
			defer wg.Done()
			s.Start(ctx, up)
		}(svr)
	}

	wg.Wait()

}

func setupLog(f, l string) (*log.Logger, error) {

	var dl *log.Logger

	h, err := log.NewRotatingFileHandler(f, 10<<10<<10, 1)
	if err != nil {
		return nil, err
	}

	dl = log.NewDefault(h)
	log.SetDefaultLogger(dl)

	setLogLevel(l)

	return dl, nil
}

func setLogLevel(level string) {
	switch strings.ToUpper(level) {
	case `DEBUG`:
		log.SetLevel(log.LevelDebug)
	case `INFO`:
		log.SetLevel(log.LevelInfo)
	case `WARN`:
		log.SetLevel(log.LevelWarn)
	case `ERROR`:
		log.SetLevel(log.LevelError)
	case `FATAL`:
		log.SetLevel(log.LevelFatal)
	default:
		log.SetLevel(log.LevelInfo)
	}
}
