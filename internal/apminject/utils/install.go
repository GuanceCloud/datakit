// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build (linux && amd64) || (linux && arm64)
// +build linux,amd64 linux,arm64

package utils

import (
	"bufio"
	"crypto/tls"
	"debug/elf"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/GuanceCloud/cliutils/logger"
	pr "github.com/shirou/gopsutil/v3/process"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
)

const (
	preloadConfigFilePath = "/etc/ld.so.preload"

	dirInject          = "apm_inject"
	dirInjectSubInject = "inject"
	dirInjectSubLib    = "lib"

	dockerDaemonJSONPath      = "/etc/docker/daemon.json"
	dockerFieldDefaultRuntime = "default-runtime"
	dockerFieldRuntimes       = "runtimes"

	launcherName = "apm_launcher"

	glibc = "glibc"
	muslc = "musl"

	dkruncBinName = "dkrunc"
)

const dockerCtrPath = "/var/lib/docker/containers"

var (
	dirDkInstall = datakit.InstallDir

	py3Regexp        = regexp.MustCompile(`^Python 3.(\d+)`)
	reGLBC           = regexp.MustCompile(`ldd \(.*\) ([0-9\.]+)`)
	reMusl           = regexp.MustCompile("musl libc \\(.*\\)\nVersion ([0-9\\.]+)")
	soGLibcVerRegexp = regexp.MustCompile(`^GLIBC_([0-9\.]+)$`)
)

func dkRuncPath() string {
	return filepath.Join(dirDkInstall, dirInject, dirInjectSubInject, dkruncBinName)
}

func Download(log *logger.Logger, opt ...Opt) error {
	var c config
	for _, fn := range opt {
		fn(&c)
	}

	if c.installDir == "" {
		c.installDir = dirDkInstall
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

	cp.Printf("\n")
	dl.CurDownloading = "apm-inject"
	injTo := filepath.Join(c.installDir, dirInject, dirInjectSubInject)
	cp.Infof("Downloading %s => %s\n", c.launcherURL, injTo)
	if err := dl.Download(c.cli, c.launcherURL, injTo,
		true, false); err != nil {
		return err
	}

	if !(c.enableHostInject || c.enableDockerInject) {
		return nil
	}

	if c.ddJavaLibURL != "" {
		cp.Printf("\n")
		dl.CurDownloading = "apm-lib-java"
		injTo = filepath.Join(c.installDir, dirInject, dirInjectSubLib,
			"java", "dd-java-agent.jar")
		cp.Infof("Downloading %s => %s\n", c.ddJavaLibURL, injTo)
		if err := dl.Download(c.cli, c.ddJavaLibURL, injTo,
			true, true); err != nil {
			log.Warn(err)
		}
	}

	if c.pyLib {
		cp.Printf("\n")
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
	}
	return nil
}

func Install(log *logger.Logger, opt ...Opt) error {
	var c config
	for _, fn := range opt {
		fn(&c)
	}

	if c.installDir == "" {
		c.installDir = dirDkInstall
	}

	if !c.enableHostInject && !c.enableDockerInject {
		if err := unsetPreload(c.installDir); err != nil {
			log.Error(err)
		}
		if err := unsetDockerRunc(dockerDaemonJSONPath); err != nil {
			log.Error(err)
		}
		return nil
	}

	// TODO: check docker inject

	if c.enableHostInject {
		libc, hostVersion, err := lddInfo()
		if err != nil {
			log.Error(err)
			if err := unsetPreload(c.installDir); err != nil {
				log.Error(err)
			}
			goto skipHost
		}
		launcherSoPath, err := laucnherSoPath(libc, c.installDir)
		if err != nil {
			log.Error(err)
			if err := unsetPreload(c.installDir); err != nil {
				log.Error(err)
			}
			goto skipHost
		}
		elfFile, err := elf.Open(launcherSoPath)
		if err != nil {
			log.Error(err)
			if err := unsetPreload(c.installDir); err != nil {
				log.Error(err)
			}
			goto skipHost
		}
		dynSyms, err := elfFile.DynamicSymbols()
		if err != nil {
			log.Error(err)
			if err := unsetPreload(c.installDir); err != nil {
				log.Error(err)
			}
			goto skipHost
		}
		required, err := requiredGLIBCVersion(dynSyms)
		if err != nil && libc == glibc {
			log.Error(err)
			if err := unsetPreload(c.installDir); err != nil {
				log.Error(err)
			}
			goto skipHost
		}
		if hostVersion.LessThan(required) {
			log.Error(fmt.Errorf("host libc version %s is less than required %s",
				hostVersion, required))
			if err := unsetPreload(c.installDir); err != nil {
				log.Error(err)
			}
			goto skipHost
		}
		if err := setPreload(c.installDir, launcherSoPath); err != nil {
			if err := unsetPreload(c.installDir); err != nil {
				log.Error(err)
			}
			goto skipHost
		}
	} else if err := unsetPreload(c.installDir); err != nil {
		log.Error(err)
	}

skipHost:
	if c.enableDockerInject {
		if err := setDockerRunc(dockerDaemonJSONPath, dkRuncPath()); err != nil {
			log.Error(err)
		}
	}

	return nil
}

func Uninstall(opt ...Opt) error {
	var c config
	for _, fn := range opt {
		fn(&c)
	}
	if c.installDir == "" {
		c.installDir = dirDkInstall
	}

	if err := unsetDockerRunc(dockerDaemonJSONPath); err != nil {
		cp.Errorf("unset docker:%s", err)
	}

	if err := unsetPreload(c.installDir); err != nil {
		cp.Errorf("unset preload:%s", err)
	}

	return nil
}

func reloadDockerConfig() error {
	processes, _ := pr.Processes()
	var pidLi []int
	for _, proc := range processes {
		if name, err := proc.Name(); err == nil && name == "dockerd" {
			pidLi = append(pidLi, int(proc.Pid))
		}
	}

	for _, pid := range pidLi {
		err := syscall.Kill(pid, syscall.SIGHUP)
		if err != nil {
			return err
		}
	}
	return nil
}

func readPreloadWithoutLanucher(preloadPath, installDir string) (string, error) {
	if _, err := os.Stat(preloadPath); err != nil {
		return "", err
	}

	f, err := os.Open(preloadConfigFilePath)
	if err != nil {
		err = fmt.Errorf("failed to read %s: %w",
			preloadConfigFilePath, err)
		return "", err
	}

	var lns []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		t := s.Text()
		lns = append(lns, t)
	}

	soPath := filepath.Join(installDir, dirInject, dirInjectSubInject, launcherName)

	var outLns []string
	for _, ln := range lns {
		if !strings.HasPrefix(ln, soPath) {
			if ln == "" && len(outLns) == 0 {
				continue
			}
			outLns = append(outLns, ln)
		}
	}

	if len(outLns) == 0 {
		return "", nil
	}

	return strings.Join(outLns, "\n") + "\n", nil
}

func unsetPreload(installDir string) error {
	lines, err := readPreloadWithoutLanucher(
		preloadConfigFilePath, installDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	//nolint:gosec
	if err := os.WriteFile(preloadConfigFilePath, []byte(lines), 0o644); err != nil {
		return fmt.Errorf("failed to clean %s: %w",
			preloadConfigFilePath, err)
	}
	return nil
}

func setPreload(installDir, soPath string) error {
	lines, err := readPreloadWithoutLanucher(
		preloadConfigFilePath, installDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	lines += soPath + "\n"

	//nolint:gosec
	if err := os.WriteFile(preloadConfigFilePath, []byte(lines), 0o644); err != nil {
		return err
	}
	return nil
}

func laucnherSoPath(kind, installDir string) (string, error) {
	var soPath string
	bp := filepath.Join(installDir, dirInject, dirInjectSubInject)
	switch kind {
	case glibc:
		soPath = filepath.Join(bp, launcherName+".so")
	case muslc:
		soPath = filepath.Join(bp, launcherName+"_musl.so")
	}
	if _, err := os.Stat(soPath); err != nil {
		return "", err
	}
	return soPath, nil
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

type dockerRuntime struct {
	name string

	// for runc
	path string
	// for shim
	shimRuntimeType string
}

func getDockerRuncFromSysInfo() (string, error) {
	bp, err := exec.LookPath("docker")
	if err != nil {
		return "", fmt.Errorf("cannot find docker: %w", err)
	}
	cmd := exec.Command(bp, "system", "info", "--format", "{{.DefaultRuntime}}") //nolint:gocritic,gosec
	o, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("run cmd `%s` failed: %w", cmd.String(), err)
	}
	runcInfo := strings.TrimSpace(string(o))
	if runcInfo != RuntimeRunc && runcInfo != RuntimeDkRunc {
		return "", fmt.Errorf("unknown runc: %s", runcInfo)
	}

	return runcInfo, nil
}

func getDockerRuntimeInfoFromConfig(cfg map[string]any) (defaultRuntime string, runtimes []dockerRuntime) {
	if val, ok := cfg[dockerFieldDefaultRuntime]; ok {
		if v, ok := val.(string); ok {
			defaultRuntime = v
		}
	}
	if val, ok := cfg[dockerFieldRuntimes]; ok {
		if runtimesVal, ok := val.(map[string]any); ok {
			for name, rinf := range runtimesVal {
				var rt dockerRuntime
				rt.name = name
				if rinf, ok := rinf.(map[string]any); ok {
					if path, ok := rinf["path"]; ok {
						if path, ok := path.(string); ok {
							rt.path = path
						}
					}
					if runtimeType, ok := rinf["runtimeType"]; ok {
						if runtimeType, ok := runtimeType.(string); ok {
							rt.shimRuntimeType = runtimeType
						}
					}
				}
				runtimes = append(runtimes, rt)
			}
		}
	}

	return
}

func loadDockerDaemonConfig(path string) (map[string]any, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	f, err := os.Open(path) //nolint:gocritic,gosec
	if err != nil {
		return nil, fmt.Errorf("open %s failed: %w", dockerDaemonJSONPath, err)
	}

	defer f.Close() //nolint:gosec,errcheck

	var r map[string]any
	if err := json.NewDecoder(f).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode %s failed: %w", dockerDaemonJSONPath, err)
	}

	return r, nil
}

func dumpDockerDaemonConfig(path string, config map[string]any) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o755) //nolint:gosec
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec
	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	return enc.Encode(config)
}

func setDockerRunc(configPath, runcPath string) error {
	if _, err := os.Stat(runcPath); err != nil {
		return err
	}

	injLdPreld := filepath.Join(dirDkInstall, dirInject, dirInjectSubInject, "ld.so.preload")
	if _, err := os.Stat(injLdPreld); err != nil {
		soPath := filepath.Join(dirDkInstall, dirInject, dirInjectSubInject, launcherName+".so") + "\n"
		if err := os.WriteFile(injLdPreld, []byte(soPath), 0o644); err != nil { //nolint:gosec
			return err
		}
	}

	config, err := loadDockerDaemonConfig(configPath)
	if err != nil {
		return err
	}

	runcName, err := getDockerRuncFromSysInfo()
	if err != nil {
		return err
	}

	if runcName == RuntimeDkRunc {
		return nil
	}

	if runcName != RuntimeRunc {
		return fmt.Errorf("docker default runtime is not runc, but: %s", runcName)
	}

	if cfgVal, _ := getDockerRuntimeInfoFromConfig(config); cfgVal != "" {
		if runcName != cfgVal {
			return fmt.Errorf("config not match the actual information: system info: %s, config: %s",
				runcName, cfgVal)
		} else {
			return nil
		}
	}
	if config == nil {
		config = map[string]any{}
	}

	var runtimes map[string]any
	if v, ok := config[dockerFieldRuntimes]; ok {
		if v, ok := v.(map[string]any); ok {
			runtimes = v
		} else {
			return fmt.Errorf("docker config `runtimes` not map")
		}
	} else {
		runtimes = map[string]any{}
	}
	config[dockerFieldDefaultRuntime] = RuntimeDkRunc
	runtimes[RuntimeDkRunc] = map[string]any{
		"path": runcPath,
	}
	config[dockerFieldRuntimes] = runtimes
	if err := dumpDockerDaemonConfig(configPath, config); err != nil {
		return err
	}

	return reloadDockerConfig()
}

func unsetDockerRunc(configPath string) error {
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	config, err := loadDockerDaemonConfig(configPath)
	if err != nil {
		return err
	}

	if len(config) == 0 {
		return nil
	}
	if v, ok := config[dockerFieldDefaultRuntime]; !ok {
		return nil
	} else {
		if v, ok := v.(string); !ok {
			return fmt.Errorf("docker config `default-runtime` not string")
		} else if v != RuntimeDkRunc {
			return nil
		}
	}
	delete(config, dockerFieldDefaultRuntime)

	if err := dumpDockerDaemonConfig(configPath, config); err != nil {
		return err
	}

	return reloadDockerConfig()
}

func ChangeDockerHostConfigRunc(from, to, ctrPath string) error {
	if ctrPath == "" {
		ctrPath = dockerCtrPath
	}

	elems, err := os.ReadDir(ctrPath)
	if err != nil {
		return err
	}

	var ctrs []string
	for _, e := range elems {
		if e.IsDir() {
			ctrs = append(ctrs, e.Name())
		}
	}

	for _, c := range ctrs {
		hc := filepath.Join(ctrPath, c, "hostconfig.json")
		data, err := os.ReadFile(hc) //nolint:gosec
		if err != nil {
			continue
		}
		var hostconfig map[string]any
		if err := json.Unmarshal(data, &hostconfig); err != nil {
			return err
		}
		if v, ok := hostconfig["Runtime"]; ok {
			if v, ok := v.(string); ok {
				if v == from {
					hostconfig["Runtime"] = to
				}
			}
		}
		f, err := os.OpenFile(hc, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o644) //nolint:gosec
		if err != nil {
			return err
		}

		if err := json.NewEncoder(f).Encode(hostconfig); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()
	}

	return nil
}
