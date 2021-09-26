// tool to build datakit

package main

import (
	"flag"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/make/build"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
)

var (
	flagBinary       = flag.String("binary", "", "binary name to build")
	flagName         = flag.String("name", *flagBinary, "same as -binary")
	flagBuildDir     = flag.String("build-dir", "build", "output of build files")
	flagMain         = flag.String(`main`, `main.go`, `binary build entry`)
	flagDownloadAddr = flag.String("download-addr", "", "")
	flagPubDir       = flag.String("pub-dir", "pub", "")
	flagArchs        = flag.String("archs", "local", "os archs")
	flagEnv          = flag.String(`env`, ``, `build for local/test/preprod/release`)
	flagPub          = flag.Bool(`pub`, false, `publish binaries to OSS: local/test/release/preprod`)
	flagBuildISP     = flag.Bool("build-isp", false, "generate ISP data")

	l = logger.DefaultSLogger("make")
)

func applyFlags() {
	if *flagBuildISP {
		curDir, _ := os.Getwd()

		inputIPDir := filepath.Join(curDir, "china-operator-ip")
		ip2ispFile := filepath.Join(curDir, "pipeline", "ip2isp", "ip2isp.txt")
		if err := os.Remove(ip2ispFile); err != nil {
			l.Warnf("os.Remove: %s, ignored", err.Error())
		}

		if err := ip2isp.MergeIsp(inputIPDir, ip2ispFile); err != nil {
			l.Errorf("MergeIsp failed: %v", err)
		} else {
			l.Infof("merge ip2isp file in `%v`", ip2ispFile)
		}

		inputFile := filepath.Join(curDir, "IP2LOCATION-LITE-DB11.CSV")
		outputFile := filepath.Join(curDir, "pipeline", "ip2isp", "contry_city.yaml")
		if !datakit.FileExist(inputFile) {
			l.Errorf("%v not exist, you can download from `https://lite.ip2location.com/download?id=9`", inputFile)
			os.Exit(0)
		}

		if err := os.Remove(ip2ispFile); err != nil {
			l.Warnf("os.Remove: %s, ignored", err.Error())
		}

		if err := ip2isp.BuildContryCity(inputFile, outputFile); err != nil {
			l.Errorf("BuildContryCity failed: %v", err)
		} else {
			l.Infof("contry and city list in file  `%v`", outputFile)
		}

		os.Exit(0)
	}

	build.AppBin = *flagBinary
	build.BuildDir = *flagBuildDir
	build.PubDir = *flagPubDir
	build.AppName = *flagName
	build.Archs = *flagArchs

	build.Release = *flagEnv
	build.MainEntry = *flagMain
	build.DownloadAddr = *flagDownloadAddr

	// override git.Version
	if x := os.Getenv("VERSION"); x != "" {
		build.ReleaseVersion = x
	}

	l.Infof("use version %s", build.ReleaseVersion)

	switch *flagEnv {
	case "release":
		l.Debug("under release, only checked inputs released")
		build.ReleaseType = "checked"
	default:
		l.Debug("under non-release, all inputs released")
		build.ReleaseType = "all"
	}

	if *flagPub {
		build.PubDatakit()
		os.Exit(0)
	}
}

func main() {
	flag.Parse()
	applyFlags()
	build.Compile()
}
