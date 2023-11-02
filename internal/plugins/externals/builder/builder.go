// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	// defer time.Sleep(time.Hour)

	if err := setup(); err != nil {
		logOutput("setup() failed: %v", err)
		return
	}

	logOutput("=========================================")
	logOutput("Start building project %s", buildInfo)
	logOutput("=========================================")

	projects := []IProject{
		&BuildOceanBaseX86{},
		&BuildOceanBaseARM{},
	}

	for _, project := range projects {
		if project.Info() == buildInfo {
			if err := project.Before(); err != nil {
				logOutput("Before() failed: %v", err)
				return
			}

			if err := project.Do(); err != nil {
				logOutput("Do() failed: %v", err)
				return
			}

			if err := project.After(); err != nil {
				logOutput("After() failed: %v", err)
				return
			}

			break
		}
	}

	logOutput("Complete!")
}

////////////////////////////////////////////////////////////////////////////////

const (
	goOci8Path        = "/tmp/src/go-oci8"
	projectPathPrefix = "/tmp/gopath/src/"
	projectPath       = "gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	ldPathOrigin = os.Getenv("LD_LIBRARY_PATH")
	mEnvs        map[string]string
	buildInfo    string

	errEnvEmpty         = fmt.Errorf("env empty")
	errNotSupportedArch = fmt.Errorf("not supported arch")
)

func setup() error {
	// Enable line numbers in logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := getEnvs(); err != nil {
		return err
	}

	getBuildInfo()

	logOutput("origin LD_LIBRARY_PATH = %s", ldPathOrigin)

	return nil
}

func getEnvs() error {
	envs := []string{
		"BUILD_PROJECT",
		"BUILD_ARCH",
		"BUILD_SOURCE",
		"BUILD_DEST",
	}

	mEnvs = make(map[string]string, len(envs))

	for _, v := range envs {
		mEnvs[v] = os.Getenv(v)
	}

	for name, text := range mEnvs {
		if len(text) == 0 {
			logOutput("%s is empty!", name)
			return errEnvEmpty
		}
	}

	return nil
}

func getBuildInfo() {
	buildInfo = mEnvs["BUILD_PROJECT"] + "/" + mEnvs["BUILD_ARCH"]
}

////////////////////////////////////////////////////////////////////////////////

type IProject interface {
	Info() string
	Before() error
	Do() error
	After() error
}

////////////////////////////////////////////////////////////////////////////////

var _ IProject = (*BuildOceanBaseX86)(nil)

type BuildOceanBaseX86 struct{}

func (bd *BuildOceanBaseX86) Info() string {
	return "oceanbase/amd64"
}

func (bd *BuildOceanBaseX86) Before() error {
	setArchEnvs(mEnvs["BUILD_ARCH"])
	setOceanBaseEnvs(mEnvs["BUILD_ARCH"])

	if err := runGoOci8Install(); err != nil {
		logOutput("runGoOci8Install() failed: %v", err)
		return err
	}

	if err := runCopyOceanBaseLibs(mEnvs["BUILD_ARCH"]); err != nil {
		logOutput("runCopyOceanBaseLibs() failed: %v", err)
		return err
	}

	return nil
}

func (bd *BuildOceanBaseX86) Do() error {
	return runBuild("ob")
}

func (bd *BuildOceanBaseX86) After() error {
	delOceanBaseEnvs(mEnvs["BUILD_ARCH"])
	delArchEnvs()

	return nil
}

////////////////////////////////////////////////////////////////////////////////

var _ IProject = (*BuildOceanBaseARM)(nil)

type BuildOceanBaseARM struct{}

func (bd *BuildOceanBaseARM) Info() string {
	return "oceanbase/arm64"
}

func (bd *BuildOceanBaseARM) Before() error {
	setCrossCompilationEnvs()
	setArchEnvs(mEnvs["BUILD_ARCH"])
	setOceanBaseEnvs(mEnvs["BUILD_ARCH"])

	if err := runGoOci8Install(); err != nil {
		logOutput("runGoOci8Install() failed: %v", err)
		return err
	}

	if err := runCopyOceanBaseLibs(mEnvs["BUILD_ARCH"]); err != nil {
		logOutput("runCopyOceanBaseLibs() failed: %v", err)
		return err
	}

	return nil
}

func (bd *BuildOceanBaseARM) Do() error {
	return runBuild("ob")
}

func (bd *BuildOceanBaseARM) After() error {
	delCrossCompilationEnvs()
	delOceanBaseEnvs(mEnvs["BUILD_ARCH"])
	delArchEnvs()

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func logOutput(format string, v ...any) {
	log.Printf(format+"\n", v...)
}

func runGoOci8Install() error {
	// cd /tmp/src/go-oci8
	// go install
	if err := os.Chdir(goOci8Path); err != nil {
		logOutput("os.Chdir() failed: %v", err)
		return err
	}

	args := []string{
		"go",
		"install",
	}
	if err := runEnv(args, os.Environ()); err != nil {
		logOutput("runEnv() failed: %v", err)
		return err
	}

	return nil
}

func runCopyOceanBaseLibs(arch string) error {
	// unalias cp
	// {
	// 	args := []string{
	// 		"unalias",
	// 		"cp",
	// 	}
	// 	_ = runEnv(args, os.Environ())
	// }

	// cp -rf /tmp/oceanbase_go/x86/u01/* /u01
	{
		var newArch string

		switch arch {
		case "amd64":
			newArch = "x86"
		case "arm64":
			newArch = "arm"
		default:
			return errNotSupportedArch
		}

		args := []string{
			"cp",
			"-rf",
			"/tmp/oceanbase_go/" + newArch + "/u01",
			"/",
		}
		if err := runEnv(args, os.Environ()); err != nil {
			logOutput("runEnv() failed: %v", err)
			return err
		}
	}

	return nil
}

func runBuild(tags string) error {
	// cd /tmp/src/go-oci8
	if err := os.Chdir(projectPathPrefix + projectPath); err != nil {
		logOutput("os.Chdir() failed: %v", err)
		return err
	}

	// Remove old.
	if err := os.RemoveAll(mEnvs["BUILD_DEST"]); err != nil {
		logOutput("os.RemoveAll() failed: %v", err)
		return err
	}
	logOutput("Remove old executable succeeded!")

	// go build xxxx
	args := []string{"go", "build"}
	if len(tags) > 0 {
		args = append(args, "-tags")
		args = append(args, tags)
	}

	moreBuild := []string{
		"-o", mEnvs["BUILD_DEST"],
		"-ldflags",
		"-w -s",
		mEnvs["BUILD_SOURCE"],
	}
	args = append(args, moreBuild...)

	if err := runEnv(args, os.Environ()); err != nil {
		logOutput("runEnv() failed: %v", err)
		return err
	}

	return nil
}

func setOceanBaseEnvs(arch string) {
	setEnv("CGO_ENABLED", "1")
	setEnv("PKG_CONFIG_PATH", goOci8Path)
	setEnv("NLS_LANG", "AMERICAN_AMERICA.UTF8")

	newLDPath := "/u01/obclient/lib"
	if len(ldPathOrigin) > 0 {
		newLDPath += ":"
	}
	newLDPath += ldPathOrigin

	logOutput("LD_LIBRARY_PATH = %s", newLDPath)
	setEnv("LD_LIBRARY_PATH", newLDPath)

	if arch == "arm64" {
		setEnv("CGO_LDFLAGS", "-g -O2 -L/u01/obclient/lib -lobclnt")
	}
}

func delOceanBaseEnvs(arch string) {
	unSetEnv("PKG_CONFIG_PATH")
	unSetEnv("NLS_LANG")
	setEnv("LD_LIBRARY_PATH", ldPathOrigin)

	if arch == "arm64" {
		setEnv("CGO_LDFLAGS", "-g -O2")
	}
}

func setArchEnvs(arch string) {
	setEnv("GOOS", "linux")
	setEnv("GOARCH", arch)
}

func delArchEnvs() {
	unSetEnv("GOOS")
	unSetEnv("GOARCH")
}

func setCrossCompilationEnvs() {
	setEnv("CC", "/opt/linaro/aarch64-linux-gnu/gcc-linaro-4.9-2016.02-x86_64_aarch64-linux-gnu/bin/aarch64-linux-gnu-gcc")
	setEnv("CXX", "/opt/linaro/aarch64-linux-gnu/gcc-linaro-4.9-2016.02-x86_64_aarch64-linux-gnu/bin/aarch64-linux-gnu-g++")
}

func delCrossCompilationEnvs() {
	unSetEnv("CC")
	unSetEnv("CXX")
}

func setEnv(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		logOutput("set env %s failed, err = %v, value = %s", key, err, value)
	}
}

func unSetEnv(key string) {
	if err := os.Unsetenv(key); err != nil {
		logOutput("unset env %s failed, err = %v", key, err)
	}
}

func runEnv(args, envs []string) error {
	log.Printf("args = %#v\n", args)
	log.Printf("envs = %#v\n", envs)

	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	if envs != nil {
		cmd.Env = append(os.Environ(), envs...)
	}

	out, err := cmd.CombinedOutput()

	log.Println(string(out))

	return err
}
