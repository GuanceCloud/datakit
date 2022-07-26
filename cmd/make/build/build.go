// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package build implement datakit build & release functions.
package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
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

	AppName = "datakit"
	AppBin  = "datakit"
	OSSPath = "datakit"

	StandaloneApps = []string{
		"datakit-ebpf",
	}

	// Architectures and OS distributions, i.e,
	// darwin/amd64
	// windows/amd64
	// linux/arm64
	// ...
	Archs string

	// build race-deteciton-enabled binary.
	RaceDetection bool

	// File pathh of main.go.
	MainEntry string

	ReleaseType string

	// Where to publish install packages.
	DownloadAddr string
	BuildDir     = "build"
	PubDir       = "pub"

	// InputsReleaseType defined which inputs are available
	// during current release:
	// all: release all inputs, include unchecked.
	// checked: only release checked inputs.
	InputsReleaseType string

	l = logger.DefaultSLogger("build")
)

const (
	LOCAL        = "local"
	ALL          = "all"
	winBinSuffix = ".exe"

	ReleaseTesting    = "testing"
	ReleaseProduction = "production"
	ReleaseLocal      = "local"
)

func runEnv(args, env []string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}

	return cmd.CombinedOutput()
}

func prepare() {
	if err := os.RemoveAll(BuildDir); err != nil {
		l.Warnf("os.RemoveAll: %s, ignored", err.Error())
	}

	_ = os.MkdirAll(BuildDir, os.ModePerm)
	_ = os.MkdirAll(filepath.Join(PubDir, ReleaseType), os.ModePerm)

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
		l.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.Join(PubDir, ReleaseType, "version"),
		versionInfo,
		os.ModePerm); err != nil {
		l.Fatal(err)
	}
}

const archSep = ","

func parseArchs(s string) (archs []string) {
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

	prepare()

	curArchs = parseArchs(Archs)

	for _, arch := range curArchs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
		}

		goos, goarch := parts[0], parts[1]

		dir := fmt.Sprintf("%s/%s-%s-%s", BuildDir, AppName, goos, goarch)

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

		if err := compileArch(AppBin, goos, goarch, dir); err != nil {
			return err
		}

		// build externals
		if err := buildExternals(dir, goos, goarch, false); err != nil {
			return err
		}

		// build standalone externals
		if err := buildExternals(BuildDir, goos, goarch, true); err != nil {
			return err
		}

		if err := buildInstaller(filepath.Join(PubDir, ReleaseType), goos, goarch); err != nil {
			return err
		}
	}

	l.Infof("Done!(elapsed %v)", time.Since(start))
	return nil
}

func compileArch(bin, goos, goarch, dir string) error {
	output := filepath.Join(dir, bin)
	if goos == datakit.OSWindows {
		output += winBinSuffix
	}

	cgoEnabled := "0"
	if goos == datakit.OSDarwin && runtime.GOOS == datakit.OSDarwin { // darwin version need CGO to build inputs CPU
		cgoEnabled = "1"
	}

	var cmdArgs []string

	// race-detection need cgo
	if RaceDetection && runtime.GOOS == goos && runtime.GOARCH == goarch {
		l.Infof("race deteciton enabled")
		cmdArgs = []string{
			"go", "build", "-race",
		}
	} else {
		cmdArgs = []string{
			"go", "build",
		}
	}

	cmdArgs = append(cmdArgs, []string{
		"-o", output,
		"-ldflags",
		fmt.Sprintf("-w -s -X main.InputsReleaseType=%s -X main.ReleaseVersion=%s -X main.DownloadAddr=%s",
			InputsReleaseType, ReleaseVersion, DownloadAddr),
		MainEntry,
	}...)

	var envs []string
	if RaceDetection && runtime.GOOS == goos && runtime.GOARCH == goarch {
		envs = []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
			`GO111MODULE=off`,
			"CGO_ENABLED=on",
		}
	} else {
		envs = []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
			`GO111MODULE=off`,
			"CGO_ENABLED=" + cgoEnabled,
		}
	}

	l.Debugf("building %s", fmt.Sprintf("%s-%s/%s", goos, goarch, bin))
	msg, err := runEnv(cmdArgs, envs)
	if err != nil {
		return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s", cmdArgs, envs, err, string(msg))
	}
	return nil
}

func buildInstaller(outdir, goos, goarch string) error {
	l.Debugf("building %s-%s/installer...", goos, goarch)

	installerExe := fmt.Sprintf("installer-%s-%s", goos, goarch)
	if goos == datakit.OSWindows {
		installerExe += winBinSuffix
	}

	var cmdArgs []string
	if RaceDetection && runtime.GOOS == goos && runtime.GOARCH == goarch {
		l.Infof("race deteciton enabled")
		cmdArgs = []string{
			"go", "build", "-race",
		}
	} else {
		cmdArgs = []string{
			"go", "build",
		}
	}

	cmdArgs = append(cmdArgs, []string{
		"-o", filepath.Join(outdir, installerExe),
		"-ldflags",
		fmt.Sprintf("-w -s -X main.DataKitBaseURL=%s -X main.DataKitVersion=%s", DownloadAddr, ReleaseVersion),
		"cmd/installer/main.go",
	}...)

	var envs []string
	if RaceDetection && runtime.GOOS == goos && runtime.GOARCH == goarch {
		envs = []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
			"CGO_ENABLED=on",
		}
	} else {
		envs = []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
			"CGO_ENABLED=off",
		}
	}

	msg, err := runEnv(cmdArgs, envs)
	if err != nil {
		return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s", cmdArgs, envs, err, string(msg))
	}
	return nil
}
