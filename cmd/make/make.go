package main

import (
	"flag"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/make/build"
	"os"
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
		//curDir, _ := os.Getwd()
		//inputDir := filepath.Join(curDir, "china-operator-ip")
		//outputFile := filepath.Join(curDir, "pipeline", "ip2isp", "ip2isp.go")
		//build.GenIspFile(inputDir, outputFile)

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
