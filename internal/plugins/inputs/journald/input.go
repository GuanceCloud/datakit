// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package journald collects systemd journal logs.
package journald

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/external"
)

const (
	inputName = "journald"

	defaultMountRootDir = "/rootfs"
)

var (
	l                  = logger.DefaultSLogger(inputName)
	_ inputs.Singleton = (*Input)(nil)

	defaultBinaryName   = "journald"
	journaldBinaryPaths = []string{
		"/usr/local/datakit/externals/journald",
		"./externals/journald",
		"journald",
	}

	defaultExternalLibDir  = filepath.Join(datakit.InstallDir, "externals", "systemd-libs")
	prepareNodeLibsFn      = prepareNodeLibs
	missingSharedObjectsFn = detectMissingSharedObjects
)

type Input struct {
	external.Input

	HTTPEndpoint string `toml:"http_endpoint"`
	LogPath      string `toml:"log_path"`
	LogLevel     string `toml:"log_level"`

	// Journal paths
	Paths []string `toml:"paths"`

	// Filter by systemd units
	Units []string `toml:"units"`

	// Filter by priority levels
	Priorities []string `toml:"priorities"`

	// Fields to exclude
	ExcludeFields []string `toml:"exclude_fields"`

	// Collection behavior
	TailOnly           bool `toml:"tail_only"`
	MaxEntriesPerBatch int  `toml:"max_entries_per_batch"`

	// Cursor management
	SaveCursor bool   `toml:"save_cursor"`
	CursorFile string `toml:"cursor_file"`

	CopyNodeLibs      bool     `toml:"copy_node_libs"`
	CopyNodeLibsFiles []string `toml:"copy_node_libs_files"`
	MountDir          string   `toml:"mount_dir"`

	semStop *cliutils.Sem
}

func (ipt *Input) Singleton() {}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	l.Info("journald input starting")

	// Runtime guard: journald only works on Linux
	// This allows code to compile on macOS/Windows for documentation export,
	// but actual data collection only runs on Linux
	if runtime.GOOS != "linux" {
		l.Warnf("journald input is only supported on Linux (current OS: %s), skipping data collection", runtime.GOOS)
		return
	}

	// Find journald binary
	execFile := ipt.findJournaldBinary()
	if execFile == "" {
		l.Errorf("journald binary not found, tried paths: %v", journaldBinaryPaths)
		return
	}

	// Update command to use found binary
	ipt.Input.Cmd = execFile

	ipt.applyKubernetesMode()

	if err := ipt.prepareLibraries(); err != nil {
		l.Warnf("prepare journald host libraries failed, journald collector will stay inactive: %v", err)
		return
	}

	// Build arguments for external binary
	ipt.buildArgs()

	// Run external binary via external.Input
	ipt.Input.Run()
}

func (ipt *Input) applyKubernetesMode() {
	if !isContainerOrK8sMode() {
		return
	}

	mountRootDir := ipt.effectiveMountRootDir()
	ipt.Paths = prefixedMountRootPaths(ipt.Paths, mountRootDir)
	ipt.CopyNodeLibs = true

	l.Infof("journald detected container/k8s mode (docker=%v kubernetes=%v)",
		datakit.Docker, config.IsKubernetes())
	l.Infof("journald auto enabled mount-root+copy libs: mount_dir=%s paths=%v copy_node_libs=%v copy_node_libs_files=%v",
		mountRootDir, ipt.Paths, ipt.CopyNodeLibs, ipt.CopyNodeLibsFiles)
}

func (ipt *Input) findJournaldBinary() string {
	// Try configured binary path first
	if ipt.Input.Cmd != "" {
		if _, err := os.Stat(ipt.Input.Cmd); err == nil {
			return ipt.Input.Cmd
		}
	}

	// Try default paths
	for _, path := range journaldBinaryPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func (ipt *Input) prepareLibraries() error {
	if !ipt.CopyNodeLibs {
		l.Debug("journald host library copy prepare disabled")
		return nil
	}

	inContainerMode := isContainerOrK8sMode()
	files := ipt.CopyNodeLibsFiles
	mountRootDir := defaultMountRootDir
	if inContainerMode {
		mountRootDir = ipt.effectiveMountRootDir()
	}

	if len(files) == 0 {
		if inContainerMode {
			l.Infof("journald host library prepare enabled: source=%s mount_root=%s target=%s",
				"kubernetes-auto", mountRootDir, defaultExternalLibDir)
			if err := ipt.prepareLibrariesInKubernetes(mountRootDir); err != nil {
				return err
			}
			ipt.prependLDLibraryPaths([]string{defaultExternalLibDir})
			l.Infof("journald host libraries prepared, prepend LD_LIBRARY_PATH with %s", defaultExternalLibDir)
			return nil
		}

		return fmt.Errorf("copy_node_libs_files is required when copy_node_libs=true outside container/kubernetes mode")
	}

	l.Infof("journald host library prepare enabled: source=%s mount_root=%s target=%s files=%v",
		"configured", mountRootDir, defaultExternalLibDir, files)

	if err := prepareNodeLibsFn(mountRootDir, defaultExternalLibDir, files); err != nil {
		return err
	}

	ipt.prependLDLibraryPaths([]string{defaultExternalLibDir})
	l.Infof("journald host libraries prepared, prepend LD_LIBRARY_PATH with %s", defaultExternalLibDir)
	return nil
}

func (ipt *Input) prepareLibrariesInKubernetes(mountRootDir string) error {
	seed := []string{"libsystemd.so*"}
	l.Infof("journald kubernetes auto library prepare started: mount_root=%s target=%s seed=%v",
		mountRootDir, defaultExternalLibDir, seed)

	if err := prepareNodeLibsFn(mountRootDir, defaultExternalLibDir, seed); err != nil {
		return fmt.Errorf("copy libsystemd in kubernetes: %w", err)
	}

	copied := map[string]struct{}{}

	for round := 0; round < 5; round++ {
		missing, err := missingSharedObjectsFn(defaultExternalLibDir, "libsystemd.so.0")
		if err != nil {
			return fmt.Errorf("detect missing shared objects: %w", err)
		}

		if len(missing) == 0 {
			var copiedList []string
			for so := range copied {
				copiedList = append(copiedList, so)
			}
			sort.Strings(copiedList)
			l.Infof("journald kubernetes auto library prepare completed: rounds=%d copied=%v", round+1, copiedList)
			return nil
		}

		for _, soName := range missing {
			copied[soName] = struct{}{}
			if err := prepareNodeLibsFn(mountRootDir, defaultExternalLibDir, []string{soName}); err != nil {
				return fmt.Errorf("copy missing shared object %s: %w", soName, err)
			}
		}
	}

	missing, err := missingSharedObjectsFn(defaultExternalLibDir, "libsystemd.so.0")
	if err != nil {
		return fmt.Errorf("detect missing shared objects after retry: %w", err)
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing shared objects still unresolved after retries: %v", missing)
	}

	var copiedList []string
	for so := range copied {
		copiedList = append(copiedList, so)
	}
	sort.Strings(copiedList)
	l.Infof("journald kubernetes auto library prepare completed: rounds=%d copied=%v", 5, copiedList)
	return nil
}

func (ipt *Input) prependLDLibraryPaths(paths []string) {
	if len(paths) == 0 {
		return
	}

	ipt.Input.Envs = mergeEnvList(os.Environ(), ipt.Input.Envs)

	var cleaned []string
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		cleaned = append(cleaned, path)
	}
	if len(cleaned) == 0 {
		return
	}

	for idx, env := range ipt.Input.Envs {
		if strings.HasPrefix(env, "LD_LIBRARY_PATH=") {
			current := strings.TrimPrefix(env, "LD_LIBRARY_PATH=")
			segments := cleaned
			if current != "" {
				segments = append(segments, current)
			}
			ipt.Input.Envs[idx] = fmt.Sprintf("LD_LIBRARY_PATH=%s", strings.Join(segments, ":"))
			return
		}
	}

	ipt.Input.Envs = append(ipt.Input.Envs, fmt.Sprintf("LD_LIBRARY_PATH=%s", strings.Join(cleaned, ":")))
}

func mergeEnvList(base, overrides []string) []string {
	merged := append([]string(nil), base...)
	index := make(map[string]int, len(merged))

	for idx, env := range merged {
		key, _, found := strings.Cut(env, "=")
		if found && key != "" {
			index[key] = idx
		}
	}

	for _, env := range overrides {
		key, _, found := strings.Cut(env, "=")
		if !found || key == "" {
			continue
		}

		if idx, exists := index[key]; exists {
			merged[idx] = env
			continue
		}

		index[key] = len(merged)
		merged = append(merged, env)
	}

	return merged
}

func isContainerOrK8sMode() bool {
	return datakit.Docker || config.IsKubernetes()
}

func (ipt *Input) effectiveMountRootDir() string {
	mountRootDir := strings.TrimSpace(ipt.MountDir)
	if mountRootDir == "" {
		return defaultMountRootDir
	}

	return mountRootDir
}

func prefixedMountRootPaths(paths []string, mountRootDir string) []string {
	if len(paths) == 0 {
		return nil
	}

	mountRootDir = strings.TrimSpace(mountRootDir)
	if mountRootDir == "" {
		mountRootDir = defaultMountRootDir
	}

	res := make([]string, 0, len(paths))
	for _, path := range paths {
		p := strings.TrimSpace(path)
		if p == "" {
			continue
		}

		if p == mountRootDir || strings.HasPrefix(p, mountRootDir+"/") || !strings.HasPrefix(p, "/") {
			res = append(res, p)
			continue
		}

		res = append(res, filepath.Join(mountRootDir, strings.TrimPrefix(p, "/")))
	}

	return res
}

func detectMissingSharedObjects(libDir, target string) ([]string, error) {
	l.Debugf("journald shared object probe: cwd=%s target=%s", libDir, target)
	cmd := exec.Command("ldd", target) //nolint:gosec
	cmd.Dir = libDir
	cmd.Env = mergeEnvList(os.Environ(), []string{fmt.Sprintf("LD_LIBRARY_PATH=%s", libDir)})
	out, err := cmd.CombinedOutput()
	output := string(out)

	missing := parseMissingSharedObjects(output)
	if err != nil && len(missing) == 0 {
		return nil, fmt.Errorf("run ldd in %s for %s: %w, output: %s", libDir, target, err, strings.TrimSpace(output))
	}

	l.Debugf("journald shared object probe result: target=%s missing=%v", target, missing)
	return missing, nil
}

func parseMissingSharedObjects(lddOutput string) []string {
	scanner := bufio.NewScanner(strings.NewReader(lddOutput))
	seen := make(map[string]struct{})
	var missing []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, "=> not found") {
			continue
		}

		name, _, _ := strings.Cut(line, "=>")
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		missing = append(missing, name)
	}

	sort.Strings(missing)
	return missing
}

func (ipt *Input) buildArgs() {
	args := []string{
		"--datakit-http-endpoint", ipt.HTTPEndpoint,
		"--log-path", ipt.LogPath,
		"--log-level", ipt.LogLevel,
	}

	// Add journal paths
	if len(ipt.Paths) > 0 {
		args = append(args, "--paths")
		args = append(args, strings.Join(ipt.Paths, ","))
	}

	// Add unit filters
	if len(ipt.Units) > 0 {
		args = append(args, "--units")
		args = append(args, strings.Join(ipt.Units, ","))
	}

	// Add priority filters
	if len(ipt.Priorities) > 0 {
		args = append(args, "--priorities")
		args = append(args, strings.Join(ipt.Priorities, ","))
	}

	// Add exclude fields
	if len(ipt.ExcludeFields) > 0 {
		args = append(args, "--exclude-fields")
		args = append(args, strings.Join(ipt.ExcludeFields, ","))
	}

	// Add tail only flag
	if ipt.TailOnly {
		args = append(args, "--tail-only")
	}

	// Add max entries per batch
	args = append(args, "--max-entries", strconv.Itoa(ipt.MaxEntriesPerBatch))

	// Add cursor options
	if ipt.SaveCursor && ipt.CursorFile != "" {
		args = append(args, "--save-cursor", "--cursor-file", ipt.CursorFile)
	}

	ipt.Input.Args = args
}

func prepareNodeLibs(mountRootDir, dst string, patterns []string) error {
	if err := os.MkdirAll(dst, 0o750); err != nil {
		return fmt.Errorf("mkdir %s: %w", dst, err)
	}

	var copied int
	for _, srcDir := range hostLibraryDirs(mountRootDir) {
		for _, pattern := range patterns {
			matches, err := filepath.Glob(filepath.Join(srcDir, pattern))
			if err != nil {
				return fmt.Errorf("glob %s in %s: %w", pattern, srcDir, err)
			}

			for _, match := range matches {
				ok, err := copyNodeLibFromMountRoot(mountRootDir, match, filepath.Join(dst, filepath.Base(match)))
				if err != nil {
					return err
				}
				if ok {
					copied++
				}
			}
		}
	}

	if copied == 0 {
		return fmt.Errorf("no host libraries copied from %s using patterns %v", mountRootDir, patterns)
	}

	return nil
}

func hostLibraryDirs(mountRootDir string) []string {
	return []string{
		filepath.Join(mountRootDir, "usr/lib64"),
		filepath.Join(mountRootDir, "lib64"),
		filepath.Join(mountRootDir, "usr/lib/x86_64-linux-gnu"),
		filepath.Join(mountRootDir, "lib/x86_64-linux-gnu"),
		filepath.Join(mountRootDir, "usr/lib/aarch64-linux-gnu"),
		filepath.Join(mountRootDir, "lib/aarch64-linux-gnu"),
	}
}

func copyNodeLibFromMountRoot(mountRootDir, src, dst string) (bool, error) {
	return copyNodeLibWithSeen(mountRootDir, src, dst, map[string]struct{}{})
}

func copyNodeLibWithSeen(mountRootDir, src, dst string, seen map[string]struct{}) (bool, error) {
	key := mountRootDir + ":" + src + "->" + dst
	if _, ok := seen[key]; ok {
		return false, fmt.Errorf("cyclic symlink while copying %s to %s", src, dst)
	}
	seen[key] = struct{}{}

	info, err := os.Lstat(src)
	if err != nil {
		return false, fmt.Errorf("lstat %s: %w", src, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(src)
		if err != nil {
			return false, fmt.Errorf("readlink %s: %w", src, err)
		}

		linkForDst := link
		if filepath.IsAbs(linkForDst) {
			// Host absolute symlink paths are not portable in the DataKit runtime.
			// Rewrite to local file name and copy the target file alongside it.
			linkForDst = filepath.Base(linkForDst)
		}

		_ = os.Remove(dst)
		if err := os.Symlink(linkForDst, dst); err != nil {
			return false, fmt.Errorf("symlink %s -> %s: %w", dst, linkForDst, err)
		}

		targetSrc := link
		if filepath.IsAbs(targetSrc) {
			if mountRootDir != "" {
				targetSrc = filepath.Join(mountRootDir, strings.TrimPrefix(targetSrc, "/"))
			}
		} else {
			targetSrc = filepath.Join(filepath.Dir(src), targetSrc)
		}
		targetDst := filepath.Join(filepath.Dir(dst), filepath.Base(targetSrc))
		if _, err := copyNodeLibWithSeen(mountRootDir, targetSrc, targetDst, seen); err != nil {
			return false, fmt.Errorf("copy symlink target for %s: %w", src, err)
		}
		return true, nil
	}

	if !info.Mode().IsRegular() {
		return false, nil
	}

	srcFile, err := os.Open(src) //nolint:gosec
	if err != nil {
		return false, fmt.Errorf("open %s: %w", src, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm()) //nolint:gosec
	if err != nil {
		_ = srcFile.Close()
		return false, fmt.Errorf("open %s: %w", dst, err)
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		_ = dstFile.Close()
		_ = srcFile.Close()
		return false, fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}

	if err := dstFile.Close(); err != nil {
		_ = srcFile.Close()
		return false, fmt.Errorf("close %s: %w", dst, err)
	}
	if err := srcFile.Close(); err != nil {
		return false, fmt.Errorf("close %s: %w", src, err)
	}

	return true, nil
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
	ipt.Input.Terminate()
}

func (*Input) Catalog() string { return "logging" }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&journalMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelK8s, datakit.LabelDocker}
}

func defaultInput() *Input {
	extInput := external.NewInput()

	extInput.Name = inputName
	extInput.Election = false // journald used to collect local log, do not need election
	extInput.Daemon = true
	extInput.Interval = "10s"
	extInput.Cmd = defaultBinaryName

	return &Input{
		Input:        *extInput,
		HTTPEndpoint: "http://localhost:9529",
		Paths: []string{
			"/var/log/journal",
			"/run/log/journal",
		},
		TailOnly:           true,
		MaxEntriesPerBatch: 1000,
		SaveCursor:         true,
		CursorFile:         filepath.Join(datakit.DataDir, "cache", "journald.cursor"),
		CopyNodeLibs:       false,
		MountDir:           defaultMountRootDir,
		semStop:            cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
