// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencontainers/runtime-spec/specs-go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/dkrunc/utils"

	reUtils "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/rewriter/utils"

	injUtils "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/utils"
)

type MntDir struct {
	Path             string
	SkipCheckCtrPath bool
	ReadOnly         bool
}

var log = &Log{}

func main() {
	log = NewLog("/usr/local/datakit/apm_inject/log/dkrunc.log")
	defer log.Close()
	eventRec := &EventRec{
		Time: time.Now().Format(time.RFC3339),
	}
	eventRec.Args = append(eventRec.Args, os.Args...)

	injectEnvs := [][2]string{
		{
			"LD_PRELOAD",
			filepath.Join(utils.DirInjectSubInject, "apm_launcher.so"),
		},
	}

	dirList := []MntDir{
		{ReadOnly: true, Path: utils.DirInject},
		{ReadOnly: true, Path: utils.DirInjectSubInject},
		{ReadOnly: true, Path: utils.DirInjectSubLib},
	}

	addr := injUtils.GetDKAddr()
	if addr != nil && addr.DkUds != "" {
		injectEnvs = append(injectEnvs, [2]string{
			injUtils.EnvDKSocketAddr, addr.DkUds,
		})
		dirList = append(dirList, MntDir{
			Path:             path.Dir(addr.DkUds),
			SkipCheckCtrPath: true,
		})
	}

	if bundle, config, spec, ok := loadSpecFromBundleDir(eventRec); ok {
		if newSpec, err := tryInjectSpec(bundle, spec, dirList, injectEnvs); err == nil {
			if err := dumpSpec(config, newSpec); err != nil {
				eventRec.Errors = append(eventRec.Errors, err.Error())
			}
		} else {
			eventRec.Errors = append(eventRec.Errors, err.Error())
		}
	}

	var args []string
	if len(os.Args) > 0 {
		args = append(args, os.Args[1:]...)
	}

	exitCode, err := runCmd("runc", args)
	if err != nil {
		eventRec.Errors = append(eventRec.Errors, err.Error())
	}
	eventRec.ExitCode = exitCode

	if v, err := json.Marshal(eventRec); err == nil {
		log.Info(string(v))
	}

	os.Exit(exitCode) //nolint:gocritic
}

func runCmd(name string, args []string) (int, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Start()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		return -1, err
	}

	var exitCode int
	err = cmd.Wait()
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return exitCode, err
}

func tryInjectSpec(bundle string, spec *Spec, dirList []MntDir, injectEnvs [][2]string) (*Spec, error) {
	rootPath := spec.Root.Path
	if !filepath.IsAbs(rootPath) {
		rootPath = filepath.Join(bundle, rootPath)
	}

	mnts := map[string]struct{}{}
	for i := range spec.Mounts {
		mnts[spec.Mounts[i].Destination] = struct{}{}
	}

	if err := checkContainerEnv(spec, injectEnvs); err != nil {
		return nil, err
	}

	mounts, err := setContainerMount(rootPath, dirList, mnts)
	if err != nil {
		return nil, err
	}

	spec.Mounts = append(spec.Mounts, mounts...)
	for i := range injectEnvs {
		spec.Process.Env = append(spec.Process.Env,
			strings.Join(injectEnvs[i][:], "="))
	}

	if s, err := tryModProcSpec(spec); err == nil {
		if s != nil {
			spec = s
		}
	}
	return spec, nil
}

func checkContainerEnv(spec *Spec, injEnvs [][2]string) error {
	envMap := map[string]struct{}{}
	if spec.Process == nil {
		return fmt.Errorf("spec.process is nil")
	}
	for _, v := range spec.Process.Env {
		if n := strings.Index(v, "="); n > 0 {
			envMap[v[:n]] = struct{}{}
		}
	}
	for i := range injEnvs {
		if _, ok := envMap[injEnvs[i][0]]; ok {
			return fmt.Errorf("env %s exist", injEnvs[i][0])
		}
	}

	return nil
}

func tryModProcSpec(spec *Spec) (*Spec, error) {
	if p := spec.Process; p == nil || len(spec.Process.Args) < 1 {
		return nil, fmt.Errorf("process parameters do not meet the conditions")
	}

	arg0 := spec.Process.Args[0]

	if arg0 != "java" && strings.HasSuffix(arg0, "/java") {
		return nil, fmt.Errorf("not recognized as a java program")
	}

	if spec.Root == nil || spec.Root.Path == "" {
		return nil, fmt.Errorf("container root path not found")
	}

	var pathEnv string
	for _, envVar := range spec.Process.Env {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 && parts[0] == "PATH" {
			pathEnv = parts[1]
		}
	}

	if ok, err := checkJavaInContainer(spec.Root.Path, pathEnv, arg0); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("unsupported java")
	}

	agentDir := filepath.Join(utils.DirInjectSubLib, "java/dd-java-agent.jar")

	if _, err := os.Stat(agentDir); err != nil {
		return nil, fmt.Errorf("stat dd-java-agent.jar: %w", err)
	}

	for i := 1; i < len(spec.Process.Args[1:]); i++ {
		p := strings.TrimSpace(spec.Process.Args[i])
		if strings.HasPrefix(p, "-javaagent:") {
			v := strings.TrimPrefix(p, "-javaagent:")
			if path.Base(strings.TrimSpace(v)) ==
				path.Base(agentDir) {
				return nil, fmt.Errorf("already injected")
			}
		}
	}

	args := []string{spec.Process.Args[0], fmt.Sprintf("-javaagent:%s", agentDir)}
	args = append(args, spec.Process.Args[1:]...)
	spec.Process.Args = args
	spec.Process.Env = append(spec.Process.Env,
		fmt.Sprintf("DD_TRACE_AGENT_URL=unix://%s", injUtils.DefaultDKUDS),
		fmt.Sprintf("DD_JMXFETCH_STATSD_HOST=unix://%s", injUtils.DefaultStatsDUDS),
		"DD_JMXFETCH_STATSD_PORT=0",
	)

	return spec, nil
}

func setContainerMount(root string, dirList []MntDir, mnts map[string]struct{}) ([]specs.Mount, error) {
	var mounts []specs.Mount
	for _, dir := range dirList {
		if _, ok := mnts[dir.Path]; ok {
			return nil, fmt.Errorf("mount %s conflicts", dir.Path)
		}

		if _, err := os.Stat(dir.Path); err != nil {
			return nil, fmt.Errorf("host dir %s not exist", dir.Path)
		}
		if !dir.SkipCheckCtrPath {
			ctrDir := dir.Path
			if root != "" {
				ctrDir = filepath.Join(root, dir.Path)
			}

			if elems, err := os.ReadDir(ctrDir); err == nil && len(elems) > 0 {
				return nil, fmt.Errorf("container dir %s not empty", ctrDir)
			}
		}

		mnt := specs.Mount{
			Type:        "none",
			Destination: dir.Path,
			Source:      dir.Path,
			Options: []string{
				"bind",
			},
		}
		if dir.ReadOnly {
			mnt.Options = append(mnt.Options, "ro")
		}

		mounts = append(mounts, mnt)
	}

	return mounts, nil
}

func loadSpecFromBundleDir(eventRec *EventRec) (string, string, *Spec, bool) {
	var create, beforeBundle bool
	var bundle string
	for _, v := range os.Args {
		switch strings.TrimSpace(v) {
		case "--bundle":
			beforeBundle = true
		case "create":
			create = true
		default:
			if beforeBundle {
				bundle = v
				beforeBundle = false
			}
		}
	}

	var config string
	if create && bundle != "" {
		entries, _ := os.ReadDir(bundle)
		for _, entry := range entries {
			eventRec.BundleList = append(eventRec.BundleList,
				entry.Name())
			if entry.Name() == "config.json" && !entry.IsDir() {
				config = filepath.Join(bundle, "config.json")
			}
		}
	}

	if spec, err := loadSpec(config); err != nil {
		return "", "", nil, false
	} else {
		return bundle, config, spec, true
	}
}

type Spec = specs.Spec

func loadSpec(configPath string) (*Spec, error) {
	f, err := os.Open(configPath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:gosec,errcheck
	var spec Spec

	if err := json.NewDecoder(f).Decode(&spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

func dumpSpec(configPath string, spec *Spec) error {
	fp, err := os.OpenFile(configPath, os.O_RDWR|os.O_TRUNC, 0o666) //nolint:gosec
	if err != nil {
		return err
	}
	defer fp.Close() //nolint:gosec,errcheck
	if err := json.NewEncoder(fp).Encode(spec); err != nil {
		return err
	}
	return nil
}

type EventRec struct {
	Time       string   `json:"time"`
	Args       []string `json:"args"`
	ExitCode   int      `json:"exit_code"`
	BundleList []string `json:"bundle_list"`
	Errors     []string `json:"errors"`
}

type Log struct {
	file *os.File
}

func NewLog(fp string) *Log {
	var l Log
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o777) //nolint:gosec
	if err == nil {
		l.file = f
	}
	return &l
}

func (log *Log) Info(val string) {
	if log.file != nil {
		_, _ = log.file.WriteString(val + "\n")
	}
}

func (log *Log) Close() {
	if log.file != nil {
		log.Close() //nolint:gosec,errcheck
	}
}

func FindBinary(rootfs, envPath, binaryName string) (string, error) {
	pathDirs := strings.Split(envPath, string(os.PathListSeparator))

	for _, dir := range pathDirs {
		fullPath := filepath.Join(rootfs, dir, binaryName)
		if _, err := os.Stat(fullPath); err == nil {
			if isExecutable(fullPath) {
				return fullPath, nil
			}
		}
	}

	return "", fmt.Errorf("binary %s not found", binaryName)
}

func isExecutable(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	mode := info.Mode()
	return mode&0o111 != 0
}

func checkJavaInContainer(rootfs string, envPath string, binName string) (bool, error) {
	var binPath string
	if filepath.IsAbs(binName) {
		binPath = filepath.Join(rootfs, binName)
		if !isExecutable(binPath) {
			return false, fmt.Errorf("file %s is not executable", binPath)
		}
	} else {
		if v, err := FindBinary(rootfs, envPath, binName); err == nil {
			binPath = v
		} else {
			return false, err
		}
	}
	if binPath == "" {
		return false, fmt.Errorf("binary %s not found", binName)
	}

	cmd := exec.Command(binPath, "-version")
	o, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	ver, err := reUtils.GetJavaVersion(string(o))
	if err != nil {
		return false, err
	}

	if ver < 8 {
		return false, reUtils.ErrUnsupportedJava
	}

	return true, nil
}
