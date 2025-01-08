// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencontainers/runtime-spec/specs-go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/dkrunc/utils"

	injUtils "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/utils"
)

type MntDir struct {
	Path             string
	SkipCheckCtrPath bool
	ReadOnly         bool
}

func main() {
	log := NewLog("/usr/local/datakit/apm_inject/log/dkrunc.log")
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
	if addr != nil && addr.UDSAddr != "" {
		injectEnvs = append(injectEnvs, [2]string{
			injUtils.EnvDKSocketAddr, addr.UDSAddr,
		})
		dirList = append(dirList, MntDir{
			Path:             path.Dir(addr.UDSAddr),
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

	exitCode, err := utils.RunCmd("runc", args, os.Stdout, os.Stderr)
	if err != nil {
		eventRec.Errors = append(eventRec.Errors, err.Error())
	}
	eventRec.ExitCode = exitCode

	if v, err := json.Marshal(eventRec); err == nil {
		log.Info(string(v))
	}

	os.Exit(exitCode) //nolint:gocritic
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
