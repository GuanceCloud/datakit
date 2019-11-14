package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/siddontang/go-log/log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/aliyuncms"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/binlog"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/config/all"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/uploader"
)

var (
	flagVersion = flag.Bool("version", false, `show verison info`)
	flagInit    = flag.Bool(`init`, false, `init agent`)

	flagFtGateway = flag.String("ftgateway", ``, ``)

	flagCfgFile = flag.String("config", ``, `configure file`)
	flagCfgDir  = flag.String("sub-config-dir", ``, `sub configuration dir`)

	flagLogFile  = flag.String(`log`, ``, `log file`)
	flagLogLevel = flag.String(`log-level`, ``, `log level`)
)

func checkPid(pid int) error {
	return syscall.Kill(pid, 0)
}

func agentConfPath() string {
	return filepath.Join(config.ExecutableDir, "agent.conf")
}

func agentPidPath() string {
	return filepath.Join(config.ExecutableDir, "agent.pid")
}

func agentPath() string {
	return filepath.Join(config.ExecutableDir, "agent")
}

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

	if err = setupLog(logpath, loglevel); err != nil {
		log.Fatalf("[error] %s", err)
		return
	}

	if config.Cfg.FtGateway == "" {
		log.Errorln("ftgateway required")
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

	telcfg, err := config.GenerateTelegrafConfig()
	if err != nil {
		log.Fatalf("%s", err.Error())
		return
	}

	if err = ioutil.WriteFile(agentConfPath(), []byte(telcfg), 0664); err != nil {
		log.Fatalf("%s", err.Error())
		return
	}

	log.Infoln("starting agent...")
	startAgent()

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	go func() {
		select {
		case sig := <-signals:
			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				stopAgent()
				binlog.Stop()
				cancel()
			}
		}
	}()

	up := uploader.New(config.Cfg.FtGateway)
	up.Start()
	defer up.Stop()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = binlog.Start(up); err != nil {
			log.Errorf("start binlog fail: %s", err.Error())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(1 * time.Second)

		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pid, err := ioutil.ReadFile(agentPidPath())
				if err != nil {
					return
				}

				npid, err := strconv.Atoi(string(pid))
				if err != nil || npid <= 2 {
					return
				}

				if checkPid(npid) != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		m := aliyuncms.NewAliyunCMSManager(up)
		m.Start()
	}()

	wg.Wait()
}

func stopAgent() {

	pid, err := ioutil.ReadFile(agentPidPath())
	if err == nil {
		npid, err := strconv.Atoi(string(pid))
		if err == nil && npid > 2 {
			if checkPid(npid) == nil {
				prs, err := os.FindProcess(npid)
				if err == nil && prs != nil {
					prs.Kill()
					time.Sleep(time.Millisecond * 100)
				}
			}
		}
	}
}

func startAgent() {

	stopAgent()

	env := os.Environ()
	procAttr := &os.ProcAttr{
		Env: env,
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}

	p, err := os.StartProcess(agentPath(), []string{"agent", "--config", agentConfPath()}, procAttr)
	if err != nil {
		log.Errorf("start agent failed: %s", err.Error())
	}
	ioutil.WriteFile(agentPidPath(), []byte(fmt.Sprintf("%d", p.Pid)), 0664)

	time.Sleep(time.Millisecond * 100)
}

func setupLog(f, l string) error {

	if f != "" {
		h, err := log.NewRotatingFileHandler(f, 10<<10<<10, 1)
		if err != nil {
			return err
		}

		log.SetDefaultLogger(log.NewDefault(h))
	}

	setLogLevel(l)

	return nil
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
