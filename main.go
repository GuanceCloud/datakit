package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/siddontang/go-log/log"
	"gitlab.jiagouyun.com/cloudcare-tools/ftcollector/binlog"
	"gitlab.jiagouyun.com/cloudcare-tools/ftcollector/config"
	"gitlab.jiagouyun.com/cloudcare-tools/ftcollector/git"
)

var (
	flagVersion = flag.Bool("version", false, `show verison info`)
	flagInit    = flag.Bool(`init`, false, `init agent`)

	flagInstallDir = flag.String("install-dir", "/usr/local/cloudcare/forethought/binlog", "install directory")
	flagCfgFile    = flag.String("cfg", ``, `configure file`)
	flagLogFile    = flag.String(`log`, ``, `log file`)
	flagLogLevel   = flag.String(`log-level`, `info`, `log level`)
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

	if *flagLogFile == "" {
		*flagLogFile = filepath.Join(*flagInstallDir, "ft.log")
	}

	h, err := log.NewRotatingFileHandler(*flagLogFile, 10<<10<<10, 1)
	if err != nil {
		fmt.Printf("[error] %s\n", err)
		return
	}
	log.SetDefaultLogger(log.NewDefault(h))
	setLogLevel(*flagLogLevel)

	if *flagCfgFile == "" {
		*flagCfgFile = *flagInstallDir + "/cfg.yml"
	}

	if *flagInit {

		config.Cfg.Log = *flagLogFile
		config.Cfg.LogLevel = *flagLogLevel
		config.Cfg.InstallDir = *flagInstallDir

		config.Cfg.GenInitCfg(*flagCfgFile)

		return

	}

	if err := config.LoadConfig(*flagCfgFile); err != nil {
		log.Fatalf("check config fail: %s", err.Error())
	}

	binlog.Start(config.Cfg.Binlog)

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
