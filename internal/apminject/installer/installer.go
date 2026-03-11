// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build (linux && amd64) || (linux && arm64)
// +build linux,amd64 linux,arm64

package installer

import (
	"bufio"
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
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	pr "github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/utils"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const (
	preloadConfigFilePath     = "/etc/ld.so.preload"
	dkruncBinName             = "dkrunc"
	launcherName              = "apm_launcher"
	dockerDaemonJSONPath      = "/etc/docker/daemon.json"
	dockerFieldDefaultRuntime = "default-runtime"
	dockerFieldRuntimes       = "runtimes"
	ldPreloadFileName         = "ld.so.preload"
	launcherSoFileName        = launcherName + ".so"
	launcherSoMuslFileName    = launcherName + "_musl.so"
	javaAgentFileName         = "dd-java-agent.jar"
	phpDdtraceVersion         = "1.16.0"
	phpDdtraceIniName         = "98-ddtrace.ini" // 官方命名，98 确保在大多数扩展之后加载
)

var (
	dirDkInstall = datakit.InstallDir
	py3Regexp    = regexp.MustCompile(`^Python 3.(\d+)`)
)

func dkRuncPath() string {
	return filepath.Join(dirDkInstall, DirInject, DirInjectSubInject, dkruncBinName)
}

type dockerRuntime struct {
	name string

	// for runc
	path string
	// for shim
	shimRuntimeType string
}

func launcherSoPath(kind, installDir string) (string, error) {
	var soPath string
	bp := filepath.Join(installDir, DirInject, DirInjectSubInject)
	switch kind {
	case utils.GLibc:
		soPath = filepath.Join(bp, launcherSoFileName)
	case utils.Muslc:
		soPath = filepath.Join(bp, launcherSoMuslFileName)
	}
	if _, err := os.Stat(soPath); err != nil {
		return "", err
	}
	return soPath, nil
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

// backupDockerConfig 备份 Docker 配置文件.
func backupDockerConfig(path string) error {
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	backupPath := path + ".bak." + strconv.FormatInt(time.Now().UnixNano(), 36)
	if err := os.WriteFile(backupPath, data, 0o644); err != nil { //nolint:gosec
		return fmt.Errorf("backup docker config to %s failed: %w", backupPath, err)
	}
	return nil
}

func dumpDockerDaemonConfig(path string, config map[string]any) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o644) //nolint:gosec
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

	injLdPreld := filepath.Join(dirDkInstall, DirInject, DirInjectSubInject, ldPreloadFileName)
	if _, err := os.Stat(injLdPreld); err != nil {
		soPath := filepath.Join(dirDkInstall, DirInject, DirInjectSubInject, launcherSoFileName) + "\n"
		if err := os.WriteFile(injLdPreld, []byte(soPath), 0o644); err != nil { //nolint:gosec
			return err
		}
	}

	// 备份配置文件
	if err := backupDockerConfig(configPath); err != nil {
		cp.Warnf("backup docker config failed: %v", err)
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

// cleanupPreloadOnError 在错误发生时清理 preload 配置.
func cleanupPreloadOnError(installDir string, log *logger.Logger) {
	if err := unsetPreload(installDir); err != nil {
		if log != nil {
			log.Error(err)
		}
	}
}

// installHostInject 安装主机注入功能，包含完整的错误处理.
func installHostInject(installDir string, log *logger.Logger) error {
	libc, hostVersion, err := utils.LddInfo()
	if err != nil {
		cleanupPreloadOnError(installDir, log)
		return fmt.Errorf("get ldd info failed: %w", err)
	}

	soPath, err := launcherSoPath(libc, installDir)
	if err != nil {
		cleanupPreloadOnError(installDir, log)
		return fmt.Errorf("get launcher so path failed: %w", err)
	}

	elfFile, err := elf.Open(soPath)
	if err != nil {
		cleanupPreloadOnError(installDir, log)
		return fmt.Errorf("open elf file failed: %w", err)
	}
	defer func() {
		if err := elfFile.Close(); err != nil {
			log.Warnf("close elf file failed: %v", err)
		}
	}()

	dynSyms, err := elfFile.DynamicSymbols()
	if err != nil {
		cleanupPreloadOnError(installDir, log)
		return fmt.Errorf("get dynamic symbols failed: %w", err)
	}

	required, err := utils.RequiredGLIBCVersion(dynSyms)
	if err != nil && libc == utils.GLibc {
		cleanupPreloadOnError(installDir, log)
		return fmt.Errorf("get required glibc version failed: %w", err)
	}

	if hostVersion.LessThan(required) {
		cleanupPreloadOnError(installDir, log)
		return fmt.Errorf("host libc version %s is less than required %s", hostVersion, required)
	}

	if err := setPreload(installDir, soPath); err != nil {
		cleanupPreloadOnError(installDir, log)
		return fmt.Errorf("set preload failed: %w", err)
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

	f, err := os.Open(preloadPath) //nolint:gosec
	if err != nil {
		err = fmt.Errorf("failed to read %s: %w",
			preloadPath, err)
		return "", err
	}

	var lns []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		t := s.Text()
		lns = append(lns, t)
	}

	soPath := filepath.Join(installDir, DirInject, DirInjectSubInject, launcherSoFileName)

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

func Install(log *logger.Logger, opt ...Opt) error {
	var c config
	for _, fn := range opt {
		fn(&c)
	}

	if c.installDir == "" {
		c.installDir = dirDkInstall
	}

	if err := validateInstallDir(c.installDir); err != nil {
		return err
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
		if err := installHostInject(c.installDir, log); err != nil {
			log.Error(err)
		}
	} else {
		if err := unsetPreload(c.installDir); err != nil {
			log.Error(err)
		}
	}

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

	if err := validateInstallDir(c.installDir); err != nil {
		return err
	}

	if err := unsetDockerRunc(dockerDaemonJSONPath); err != nil {
		cp.Errorf("unset docker:%s", err)
	}

	if err := unsetPreload(c.installDir); err != nil {
		cp.Errorf("unset preload:%s", err)
	}

	return nil
}

func ChangeDockerHostConfigRunc(from, to, ctrPath string) error {
	if ctrPath == "" {
		ctrPath = utils.DockerCtrPath
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
