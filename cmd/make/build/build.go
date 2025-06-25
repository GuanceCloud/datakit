// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package build implement datakit build & release functions.
package build

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/upgrader/upgrader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	ValueNotSet = "NOT_SET"
)

var (
	// OSArches defined current supported release OS/Archs.
	// Use `go tool dist list` to get golang supported os/archs.
	OSArches = []string{ // supported os/arch list
		// Linux
		`linux/386`,
		`linux/amd64`,
		`linux/arm`,
		`linux/arm64`,

		// Darwin
		// NOTE: currently we apply amd64 arch for arm64 on Mac M1
		`darwin/amd64`,

		// Windows
		`windows/amd64`,
		`windows/386`,
	}

	// ReleaseVersion default use git describe output, you
	// can override this by set environment VERSION.
	ReleaseVersion = git.Version

	DCAVersion = ValueNotSet

	AppName = "datakit"
	AppBin  = "datakit"

	StandaloneApps = []string{
		"datakit-ebpf",
	}

	// Architectures and OS distributions, i.e,
	// darwin/amd64
	// windows/amd64
	// linux/arm64
	// ...
	Archs string

	AWSRegions      string
	EnableUploadAWS bool
	// build race-deteciton-enabled binary.
	RaceDetection string

	// File pathh of main.go.
	MainEntry string

	ReleaseType string

	Brand string

	DockerImageRepo string

	// Where to publish/download install packages.
	DownloadCDN = ValueNotSet

	OnlyExternalInputs = 0

	DistDir      = ValueNotSet
	HelmChartDir = ValueNotSet

	// InputsReleaseType defined which inputs are available
	// during current release:
	// all: release all inputs, include unchecked.
	// checked: only release checked inputs.
	InputsReleaseType string

	optionOff = "off"

	l      = logger.DefaultSLogger("build")
	ossCli *cliutils.OssCli
)

const (
	LOCAL        = "local"
	ALL          = "all"
	winBinSuffix = ".exe"

	ReleaseTesting    = "testing"
	ReleaseProduction = "production"
	ReleaseLocal      = "local"
)

func SetLog() {
	l = logger.SLogger("build")
}

// LoadENVs load all CI/CD environments.
func LoadENVs() {
	if x := os.Getenv("ROBOT_TOKEN"); x != "" {
		NotifyToken = x
	}

	var err error
	ossCli, err = getOSSInfo()
	if err != nil {
		l.Warnf("getOSSInfo: %s, ignored\n", err)
	}

	updateDownloadCDN(ReleaseType, ossCli)
}

func updateDownloadCDN(rtype string, oc *cliutils.OssCli) {
	switch rtype {
	case ReleaseTesting, ReleaseLocal:
		if oc != nil {
			// under test/local pub, do not upload dist to CDN, just upload to OSS
			DownloadCDN = fmt.Sprintf("%s.%s/%s", oc.BucketName, oc.Host, oc.WorkDir)
		}
	case ReleaseProduction: // pass
		DownloadCDN = fmt.Sprintf("%s/datakit", brand(Brand).staticURL())
		l.Infof("set DownloadCDN: %q", DownloadCDN)
	}
}

func runEnv(args, env []string) ([]byte, error) {
	l.Infof("run command %v", append(env, args...))

	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}

	return cmd.CombinedOutput()
}

func CompileDCA() error {
	curArchs = ParseArchs(Archs)
	l.Debugf("curArchs = %v", curArchs)

	for _, arch := range curArchs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
		}

		goos, goarch := parts[0], parts[1]

		// build dca binary
		dcaDir := fmt.Sprintf("%s/%s-%s-%s", DistDir, "dca", goos, goarch)
		if err := compileArch("dca", goos, goarch, dcaDir, "cmd/dca/main.go", ValueNotSet); err != nil {
			return fmt.Errorf("unable to build DCA : %w", err)
		}
	}

	// generate install scripts and yaml during compiling.
	if err := buidlDCATemplates(); err != nil {
		return err
	}

	// prepare DCA version
	vd := &versionDesc{
		Version:  strings.TrimSpace(DCAVersion),
		Date:     git.BuildAt,
		Uploader: git.Uploader,
		Branch:   git.Branch,
		Commit:   git.Commit,
		Go:       git.Golang,
	}

	versionInfo, err := json.MarshalIndent(vd, "", "    ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(DistDir, "version.dca"), versionInfo, os.ModePerm); err != nil {
		return err
	}

	// upload DCA pod yaml to OSS.
	// NOTE: we build DCA and upload yaml to oss directly.
	// DCA currently only upload docker images to pubrepo, and no binary files upload to OSS.
	if err := pubDCA(); err != nil {
		return err
	}

	return nil
}

func prepare() error {
	if err := os.RemoveAll(DistDir); err != nil {
		l.Warnf("os.RemoveAll: %s, ignored", err.Error())
	}

	if err := os.MkdirAll(DistDir, os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(DistDir, ReleaseType), os.ModePerm); err != nil {
		return err
	}

	// create version info
	vd := &versionDesc{
		Version:  strings.TrimSpace(ReleaseVersion),
		Date:     git.BuildAt,
		Uploader: git.Uploader,
		Branch:   git.Branch,
		Commit:   git.Commit,
		Go:       git.Golang,
	}

	versionInfo, err := json.MarshalIndent(vd, "", "    ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(DistDir, ReleaseType, "version"),
		versionInfo,
		os.ModePerm); err != nil {
		return err
	}

	return nil
}

const archSep = ","

func ParseArchs(s string) (archs []string) {
	switch s {
	case ALL:

		// read cmd-line env
		if x := os.Getenv("ALL_ARCHS"); x != "" {
			archs = strings.Split(x, archSep)
		} else {
			archs = OSArches
		}

	case LOCAL:
		if x := os.Getenv("LOCAL"); x != "" {
			if x == "all" { // 指定 local 为 all，便于测试全平台编译/发布
				archs = OSArches
			} else {
				archs = strings.Split(x, archSep)
			}
		} else {
			archs = []string{runtime.GOOS + "/" + runtime.GOARCH}
		}
	default:
		archs = strings.Split(s, archSep)
	}

	return
}

var curArchs []string

var curEBpfArchs []string

func Compile() error {
	start := time.Now()

	if err := prepare(); err != nil {
		return err
	}

	curArchs = ParseArchs(Archs)
	l.Debugf("curArchs = %v", curArchs)

	for _, arch := range curArchs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
		}

		goos, goarch := parts[0], parts[1]

		dir := fmt.Sprintf("%s/%s-%s-%s", DistDir, AppName, goos, goarch)
		l.Debugf("dir = %s", dir)

		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			l.Errorf("failed to mkdir: %v", err)
			return err
		}

		dir, err = filepath.Abs(dir)
		if err != nil {
			l.Errorf("filepath.Abs: %s", err)
			return err
		}

		// build externals
		if err := BuidlExternals(dir, goos, goarch, false); err != nil {
			return err
		}

		// build standalone externals
		if err := BuidlExternals(DistDir, goos, goarch, true); err != nil {
			return err
		}

		if OnlyExternalInputs == 1 {
			l.Infof("under only-input-extensions, skip compiling other dists...")
			continue // skip following build
		}

		// build lite and elinker datakit
		if isExtraLite() {
			dir := fmt.Sprintf("%s/%s_lite-%s-%s", DistDir, AppName, goos, goarch)
			if err := compileArch(AppBin, goos, goarch, dir, MainEntry, "datakit_lite && with_inputs"); err != nil {
				return err
			}
		}

		if isExtraELinker() {
			dir := fmt.Sprintf("%s/%s_elinker-%s-%s", DistDir, AppName, goos, goarch)
			if err := compileArch(AppBin, goos, goarch, dir, MainEntry, "datakit_elinker && with_inputs"); err != nil {
				return err
			}
		}

		if isExtraAWSLambda() &&
			(goarch == archAMD64 || goarch == archARM64) && /* enable build under macOS for debugging. */
			goos != "windows" { // windows not need currently
			var (
				dir       = fmt.Sprintf("%s/%s_aws_lambda-%s-%s/extensions", DistDir, AppName, goos, goarch)
				mainEntry = filepath.Join(filepath.Dir(filepath.Dir(MainEntry)), "awslambda", "main.go")
			)

			if err := compileArch(AppBin, goos, goarch, dir, mainEntry, "datakit_aws_lambda && with_inputs"); err != nil {
				return err
			}

			cmd := exec.Command("zip", []string{"-r", AppName + "_aws_extension.zip", "extensions/"}...)
			cmd.Dir = filepath.Dir(dir)

			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to run: %w, msg: %s", err, string(output))
			}
		} else {
			l.Infof("skip datakit lambda extensions under %s/%s", goos, goarch)
		}

		if err := compileArch(AppBin, goos, goarch, dir, MainEntry, "with_inputs"); err != nil {
			return err
		}

		if err := compileAPMInject(goos, goarch, DistDir); err != nil {
			l.Warnf("build APM inject failed: %s, ignored", err)
		}

		upgraderDir := fmt.Sprintf("%s/%s-%s-%s", DistDir, upgrader.BuildBinName, goos, goarch)
		l.Debugf("upgraderDir = %s", upgraderDir)
		if err := compileArch(upgrader.BuildBinName, goos, goarch, upgraderDir, upgrader.BuildEntranceFile, "not-set"); err != nil {
			return fmt.Errorf("unable to build %s : %w", upgrader.BuildBinName, err)
		}

		if err := buildInstaller(filepath.Join(DistDir, ReleaseType), goos, goarch); err != nil {
			return err
		}
	}

	if OnlyExternalInputs == 1 {
		l.Infof("under only-input-extensions, skip build templates and other dists...")
		return nil
	}

	// generate install scripts and yaml during compiling.
	if err := buildTemplates(); err != nil {
		return err
	}

	l.Infof("build helm package...")
	if err := buildDatakitHelm(); err != nil {
		l.Warnf("buildDatakitHelm: %s, ignore", err)
	}

	// export docs
	exporter := export.NewIntegration(export.WithTopDir(DistDir))
	for _, lang := range []inputs.I18n{inputs.I18nZh, inputs.I18nEn} {
		if err := exporter.ExportMiscs(lang); err != nil {
			return err
		}
	}
	if err := exporter.DumpFile(); err != nil {
		return err
	}

	l.Infof("Done!(elapsed %v)", time.Since(start))
	return nil
}

func compileArch(bin, goos, goarch, dir, mainEntranceFile, tags string) error {
	isLite := false
	isELinker := false
	if strings.Contains(tags, "datakit_lite") {
		isLite = true
	} else if strings.Contains(tags, "datakit_elinker") {
		isELinker = true
	}

	output := filepath.Join(dir, bin)
	if goos == datakit.OSWindows {
		output += winBinSuffix
	}

	cgoEnabled := "0"
	if goos == datakit.OSDarwin && runtime.GOOS == datakit.OSDarwin { // darwin version need CGO to build inputs CPU
		cgoEnabled = "1"
	}

	var cmdArgs []string

	if tags == "" {
		tags = ValueNotSet
	}

	// race-detection need cgo
	if RaceDetection != optionOff && runtime.GOOS == goos && runtime.GOARCH == goarch {
		l.Infof("race deteciton enabled")
		cmdArgs = []string{
			"go", "build",
			"-tags", tags,
			"-race",
		}
	} else {
		cmdArgs = []string{
			"go", "build",
			"-tags", tags,
		}
	}

	//nolint: lll
	ldflags := fmt.Sprintf("-w -s "+
		"-X main.Lite=%v "+
		"-X main.ELinker=%v "+
		"-X main.InputsReleaseType=%s "+
		"-X main.ReleaseVersion=%s "+
		"-X gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds.StaticCDN=%s "+
		"-X gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit.BrandDomain=%s ",
		isLite, isELinker, InputsReleaseType, ReleaseVersion, brand(Brand).staticURL(), brand(Brand).domain())

	cmdArgs = append(cmdArgs, []string{
		"-o", output,
		"-ldflags", ldflags,
		mainEntranceFile,
	}...)

	var envs []string
	if RaceDetection != optionOff && runtime.GOOS == goos && runtime.GOARCH == goarch {
		envs = []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
			`GO111MODULE=off`,
			"CGO_ENABLED=on",
			"CGO_CFLAGS=-Wno-undef-prefix",
		}
	} else {
		envs = []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
			`GO111MODULE=off`,
			"CGO_CFLAGS=-Wno-undef-prefix",
			"CGO_ENABLED=" + cgoEnabled,
		}
	}

	l.Debugf("building %q with %v", fmt.Sprintf("%s-%s/%s", goos, goarch, bin), cmdArgs)
	msg, err := runEnv(cmdArgs, envs)
	if err != nil {
		return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s", cmdArgs, envs, err, string(msg))
	}
	return nil
}

func compileAPMInject(goos, goarch, dir string) error {
	//nolint:gosec
	_ = os.Mkdir(filepath.Join(dir, "/datakit-apm-inject-linux-"+goarch), 0o755)

	// skip build under macOS, we'll never release production package under macOS.
	if goos != "linux" || runtime.GOOS != "linux" {
		l.Warnf("skip building apm auto-inject launcher: unsupported os %s", goos)
		return nil
	}

	if goarch != "amd64" && goarch != "arm64" {
		l.Warnf("skip building apm auto-inject launcher: unsupported arch %s", goarch)
		return nil
	}

	_, err := exec.LookPath("docker")
	if err != nil {
		l.Warnf("skip building apm auto-inject launcher: %s",
			err.Error())
		return nil
	}

	cmdArgs := []string{
		"sh", "internal/apminject/build_lib.sh",
		goarch, fmt.Sprintf("%s/datakit-apm-inject-linux-%s", dir, goarch),
	}

	l.Debugf("building %v with %v", fmt.Sprintf("%s-%s/%s",
		goos, goarch, "apm-auto-inject-launcher"), cmdArgs)

	var envs []string
	msg, err := runEnv(cmdArgs, envs)
	if err != nil {
		return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s",
			cmdArgs, envs, err, string(msg))
	}

	return nil
}

// is_extra_lite check whether to build lite datakit.
func isExtraLite() bool {
	extraLite := true
	liteDisable := os.Getenv("LITE_DISABLE")
	if len(liteDisable) > 0 {
		if v, err := strconv.ParseBool(liteDisable); err != nil {
			l.Warnf("parse LITE_DISABLE error: %s, ignore", err.Error())
		} else {
			extraLite = !v
		}
	}

	return extraLite
}

// is_extra_elinker check whether to build elinker datakit.
func isExtraELinker() bool {
	extraELinker := true
	elinkerDisable := os.Getenv("ELINKER_DISABLE")
	if len(elinkerDisable) > 0 {
		if v, err := strconv.ParseBool(elinkerDisable); err != nil {
			l.Warnf("parse ELINKER_DISABLE error: %s, ignore", err.Error())
		} else {
			extraELinker = !v
		}
	}

	return extraELinker
}

// is_extra_aws_lambda check whether to build aws lambda datakit.
func isExtraAWSLambda() bool {
	extraAWSLambda := true
	awsLambdaDisable := os.Getenv("AWS_LAMBDA_DISABLE")
	if len(awsLambdaDisable) > 0 {
		if v, err := strconv.ParseBool(awsLambdaDisable); err != nil {
			l.Warnf("parse AWS_LAMBDA_DISABLE error: %s, ignore", err.Error())
		} else {
			extraAWSLambda = !v
		}
	}

	return extraAWSLambda
}
