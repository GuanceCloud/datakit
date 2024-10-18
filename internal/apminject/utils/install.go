// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build (linux && amd64) || (linux && arm64)
// +build linux,amd64 linux,arm64

package utils

import (
	"crypto/tls"
	"debug/elf"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
)

const (
	preloadConfigFilePath = "/etc/ld.so.preload"

	injectInstallDir = "/usr/local/datakit"
	injectDir        = "apm_inject"
	subDirLib        = "lib"
	subDirInject     = "inject"

	launcherName = "apm_launcher"

	jarLibURL = "https://static.guance.com/dd-image/dd-java-agent.jar"

	glibc = "glibc"
	muslc = "musl"
)

var (
	py3Regexp        = regexp.MustCompile(`^Python 3.(\d+)`)
	reGLBC           = regexp.MustCompile(`ldd \(.*\) ([0-9\.]+)`)
	reMusl           = regexp.MustCompile("musl libc \\(.*\\)\nVersion ([0-9\\.]+)")
	soGLibcVerRegexp = regexp.MustCompile(`^GLIBC_([0-9\.]+)$`)
)

func Download(log *logger.Logger, opt ...Opt) error {
	var c config
	for _, fn := range opt {
		fn(&c)
	}

	if !(c.enableHostInject || c.enableDockerInject) {
		return nil
	}

	if c.installDir == "" {
		c.installDir = injectInstallDir
	}

	if c.launcherURL == "" {
		return fmt.Errorf("apm inject url is empty")
	}

	if c.cli == nil {
		c.cli = httpcli.Cli(&httpcli.Options{
			// ignore SSL error
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint
		})
	}
	fmt.Printf("\n")

	dl.CurDownloading = "apm-inject"
	injTo := filepath.Join(c.installDir, injectDir, subDirInject)
	cp.Infof("Downloading %s => %s\n", c.launcherURL, injTo)
	if err := dl.Download(c.cli, c.launcherURL, injTo,
		true, false); err != nil {
		return err
	}

	fmt.Printf("\n")

	dl.CurDownloading = "apm-lib-java"
	injTo = filepath.Join(c.installDir, injectDir, subDirLib,
		"java", "dd-java-agent.jar")
	cp.Infof("Downloading %s => %s\n", jarLibURL, injTo)
	if err := dl.Download(c.cli, jarLibURL, injTo,
		true, true); err != nil {
		log.Warn(err)
	}

	fmt.Printf("\n")

	cp.Infof("Installing ddtrace python library\n")
	py, err := exec.LookPath("python3")
	if err != nil {
		py, err = exec.LookPath("python")
		if err != nil {
			log.Warn("python not found")
		}
	}
	if py != "" {
		//nolint:gosec
		ver, err := exec.Command(py, "-V").CombinedOutput()
		if err != nil {
			log.Warnf("%s -V error: %s", py, err.Error())
			goto skip_python_lib
		}
		v := py3Regexp.FindStringSubmatch(string(ver))
		var py3Ver int
		if len(v) == 2 {
			py3Ver, _ = strconv.Atoi(v[1])
		} else {
			log.Warnf("parse python version error: %s", string(ver))
			goto skip_python_lib
		}
		if py3Ver >= 7 {
			// set env: PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple
			//nolint:gosec
			if s, err := exec.Command(py, "-m",
				"pip", "install", "ddtrace").CombinedOutput(); err != nil {
				log.Warn(string(s))
				log.Warnf("pip install ddtrace error: %s", err.Error())
			} else {
				log.Info(string(s))
			}
		}
	}
skip_python_lib:

	return nil
}

func Install(opt ...Opt) error {
	var c config
	for _, fn := range opt {
		fn(&c)
	}

	if c.installDir == "" {
		c.installDir = injectInstallDir
	}

	if !c.enableHostInject && !c.enableDockerInject {
		unsetPreload(c.installDir)
		return nil
	}

	// TODO: check docker inject

	if c.enableHostInject {
		libc, hostVersion, err := lddInfo()
		if err != nil {
			unsetPreload(c.installDir)
			return err
		}
		launcherSoPath := laucnherSoPath(libc, c.installDir)
		elfFile, err := elf.Open(launcherSoPath)
		if err != nil {
			unsetPreload(c.installDir)
			return err
		}
		dynSyms, err := elfFile.DynamicSymbols()
		if err != nil {
			unsetPreload(c.installDir)
			return err
		}
		required, err := requiredGLIBCVersion(dynSyms)
		if err != nil && libc == glibc {
			unsetPreload(c.installDir)
			return err
		}
		if hostVersion.LessThan(required) {
			unsetPreload(c.installDir)
			return fmt.Errorf("host libc version %s is less than required %s",
				hostVersion, required)
		}
		if err := setPreload(c.installDir, launcherSoPath); err != nil {
			unsetPreload(c.installDir)
			return err
		}
	} else {
		unsetPreload(c.installDir)
	}

	return nil
}

func Uninstall(opt ...Opt) error {
	var c config
	for _, fn := range opt {
		fn(&c)
	}
	if c.installDir == "" {
		c.installDir = injectInstallDir
	}

	unsetPreload(c.installDir)
	return nil
}

func readPreloadWithoutLanucher(installDir string) (libs []byte) {
	f, err := os.ReadFile(preloadConfigFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %s",
			preloadConfigFilePath, err.Error())
		return
	}

	prefix := filepath.Join(installDir, injectDir, subDirInject, launcherName)
	for _, ln := range strings.Split(string(f), "\n") {
		if ln == "" {
			continue
		}
		if !strings.HasPrefix(ln, prefix) {
			libs = append(libs, []byte(ln)...)
			libs = append(libs, '\n')
		}
	}
	return libs
}

func unsetPreload(installDir string) {
	if _, err := os.Stat(preloadConfigFilePath); err != nil {
		return
	}

	lines := readPreloadWithoutLanucher(installDir)
	//nolint:gosec
	if err := os.WriteFile(preloadConfigFilePath, lines, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to clean %s: %s",
			preloadConfigFilePath, err.Error())
		return
	}
}

func setPreload(installDir, soPath string) error {
	lines := readPreloadWithoutLanucher(installDir)
	lines = append(lines, []byte(soPath)...)
	lines = append(lines, '\n')

	//nolint:gosec
	if err := os.WriteFile(preloadConfigFilePath, lines, 0o644); err != nil {
		return err
	}
	return nil
}

func laucnherSoPath(kind, installDir string) string {
	var soPath string
	bp := filepath.Join(installDir, injectDir, subDirInject)
	switch kind {
	case glibc:
		soPath = filepath.Join(bp, launcherName+".so")
	case muslc:
		soPath = filepath.Join(bp, launcherName+"_musl.so")
	}
	return soPath
}

func lddInfo() (string, Version, error) {
	fp, err := exec.LookPath("ldd")
	if err != nil {
		return "", Version{}, err
	}
	//nolint:gosec
	cmd := exec.Command(fp, "--version")
	o, err := cmd.CombinedOutput()
	if err != nil {
		return "", Version{}, err
	}

	text := string(o)
	if v1, v2, ok := libcInfo(text); !ok {
		return "", Version{}, fmt.Errorf("unknown libc")
	} else {
		var version Version
		if err := version.Parse(v2); err != nil {
			return "", Version{}, fmt.Errorf("parse version failed: %w", err)
		}
		return v1, version, nil
	}
}

func libcInfo(text string) (string, string, bool) {
	v := reGLBC.FindStringSubmatch(text)
	if len(v) != 2 {
		v = reMusl.FindStringSubmatch(text)
		if len(v) != 2 {
			return "", "", false
		}
		return muslc, v[1], true
	} else {
		return glibc, v[1], true
	}
}

type Version [3]int

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])
}

func (v Version) LessThan(other Version) bool {
	for i := 0; i < len(v); i++ {
		if v[i] < other[i] {
			return true
		}
		if v[i] > other[i] {
			return false
		}
	}
	return false
}

func (v *Version) Parse(str string) error {
	vStr := strings.Split(str, ".")
	if len(vStr) > 3 {
		return fmt.Errorf("invalid version: %s",
			str)
	}

	tmpV := Version{}
	for i := 0; i < len(vStr); i++ {
		val, err := strconv.Atoi(vStr[i])
		if err != nil {
			return fmt.Errorf("invalid version: %s",
				str)
		}
		tmpV[i] = val
	}
	*v = tmpV
	return nil
}

func requiredGLIBCVersion(dynamicSymbols []elf.Symbol) (Version, error) {
	var required Version

	for _, sym := range dynamicSymbols {
		versionMatch := soGLibcVerRegexp.FindStringSubmatch(sym.Version)
		if len(versionMatch) != 2 {
			continue
		}
		versionStr := versionMatch[1]

		var v Version
		if err := v.Parse(versionStr); err != nil {
			return Version{}, err
		} else if required.LessThan(v) {
			required = v
		}
	}
	return required, nil
}
