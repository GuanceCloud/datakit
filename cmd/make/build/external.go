// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type dkexternal struct {
	name       string
	out        string
	standalone bool

	lang string // go/others

	entry     string
	buildArgs []string

	osarchs map[string]bool
	envs    []string

	buildCmd string

	tags string
}

const externalEBPFName = "ebpf"

var externals = []*dkexternal{
	{
		// requirement: apt-get install gcc-multilib
		name: "oracle",
		lang: "go",

		entry: "internal/plugins/externals/oracle/main.go",
		tags:  "netgo",
		osarchs: map[string]bool{
			"linux/amd64": true,
			"linux/arm64": true,
		},

		envs: []string{
			"CGO_ENABLED=1",
		},
	},
	{
		// requirement: apt-get install gcc-multilib
		name: "db2",
		lang: "go",

		entry: "internal/plugins/externals/db2/main.go",
		tags:  "netgo",
		osarchs: map[string]bool{
			"linux/amd64": true,
		},

		envs: []string{
			"CGO_ENABLED=1",
		},
	},
	{
		// requirement: apt install clang llvm linux-headers-$(uname -r)
		name:       externalEBPFName,
		out:        "datakit-ebpf",
		standalone: false,
		lang:       "makefile",

		entry: "internal/plugins/externals/ebpf/Makefile",
		osarchs: map[string]bool{
			"linux/amd64": true,
			"linux/arm64": true,
		},

		buildArgs: nil,
	},
	{
		// requirement: libsystemd-dev for journald CGO bindings
		name: "journald",
		lang: "go",

		entry: "internal/plugins/externals/journald/main.go",
		tags:  "netgo",
		osarchs: map[string]bool{
			"linux/amd64": true,
			"linux/arm64": true,
		},

		buildArgs: nil,
		envs: []string{
			"CGO_ENABLED=1",
		},
	},
	{
		name: "logfwd",
		lang: "go",

		entry: "internal/plugins/externals/logfwd/cmd/main.go",
		osarchs: map[string]bool{
			"linux/amd64": true,
			"linux/arm64": true,
		},

		buildArgs: nil,
		envs: []string{
			"CGO_ENABLED=0",
		},
	},
}

func doBuildExternal(ex *dkexternal, dir, goos, goarch string, standalone bool) error {
	var (
		// NOTE: never using ex.envs for appending, it would be modified and poisoned in the future use.
		envs      = make([]string, len(ex.envs))
		buildArgs = make([]string, len(ex.buildArgs))
		curOSArch = runtime.GOOS + "/" + runtime.GOARCH
		osarch    = goos + "/" + goarch
	)

	copy(envs, ex.envs)
	copy(buildArgs, ex.buildArgs)

	if ex.standalone != standalone {
		return nil
	}

	l.Infof("building %s-%s/%s", goos, goarch, ex.name)

	if _, ok := ex.osarchs[curOSArch]; !ok {
		l.Warnf("skip build %s under %s", ex.name, curOSArch)
		return nil
	}

	if _, ok := ex.osarchs[osarch]; !ok {
		l.Warnf("skip build %s under %s", ex.name, osarch)
		return nil
	}

	// switch ex.name {
	// case "db2", "oracle", "journald":
	//	if env := os.Getenv("ENABLE_DOCKER_BUILD_INPUTS"); len(env) == 0 {
	//		l.Warnf("WARNING: skip build %s because env not specified!", ex.name)
	//		return nil
	//	}

	//	str, err := exec.LookPath("docker")
	//	if err != nil {
	//		l.Warnf("WARNING: skip build %s because docker is NOT exist!", ex.name)
	//		return nil
	//	}

	//	l.Infof("Found docker in %s", str)
	//}

	out := ex.name
	if ex.out != "" {
		out = ex.out
	}

	var outdir string
	if ex.standalone {
		outdir = filepath.Join(dir, "standalone", fmt.Sprintf("%s-%s-%s", out, goos, goarch))
	} else {
		outdir = filepath.Join(dir, "externals")
	}

	l.Info("lang = ", ex.lang)
	switch {
	case needContainerBuild(ex, envs):
		if err := buildExternalInContainer(ex, outdir, out, goos, goarch); err != nil {
			return err
		}
		if ex.name == externalEBPFName {
			if err := buildEBPFCollectorLocal(outdir, out, goos, goarch); err != nil {
				return err
			}
		}

	case strings.EqualFold(ex.lang, "go") || strings.EqualFold(ex.lang, "golang"):
		if err := buildExternalWithGo(ex, outdir, out, goos, goarch, envs); err != nil {
			return err
		}

	default: // for python, just copy source code into build dir
		buildArgs = append(buildArgs, filepath.Join(outdir, "externals")) //nolint:makezero
		cmd := exec.Command(ex.buildCmd, buildArgs...)                    //nolint:gosec
		if len(envs) > 0 {
			cmd.Env = envs
		}

		res, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to build python(%s %s): %s, err: %w",
				ex.buildCmd, strings.Join(buildArgs, " "), res, err)
		}
	}

	return nil
}

func buildEBPFCollectorLocal(outdir, out, goos, goarch string) error {
	args, buildEnv := ebpfCollectorLocalBuildCommand(outdir, out, goos, goarch, time.Now().UTC())

	msg, err := runEnv(args, buildEnv)
	if err != nil {
		return fmt.Errorf("failed to build external ebpf locally without cgo: %w, msg: %s", err, string(msg))
	}

	return nil
}

func ebpfCollectorLocalBuildCommand(outdir, out, goos, goarch string, buildAt time.Time) ([]string, []string) {
	args := []string{
		"go",
		"build",
		"-tags", "ebpf netgo",
		"-buildvcs=false",
		"-o", filepath.Join(outdir, out),
		"-ldflags", fmt.Sprintf("-w -s -X 'main.Arch=%s/%s' -X 'main.Date=%s'",
			goos, goarch, buildAt.UTC().Format(time.RFC3339)),
		"internal/plugins/externals/ebpf/cmd/datakit-ebpf/datakit-ebpf.go",
	}

	buildEnv := []string{
		"GO111MODULE=on",
		fmt.Sprintf("GOOS=%s", goos),
		fmt.Sprintf("GOARCH=%s", goarch),
		"CGO_ENABLED=0",
		"GOFLAGS=-mod=" + goModuleMode(),
	}

	return args, buildEnv
}

func needContainerBuild(ex *dkexternal, envs []string) bool {
	if strings.EqualFold(ex.lang, "makefile") {
		return true
	}

	return envValue(envs, "CGO_ENABLED") == "1"
}

func buildExternalWithGo(ex *dkexternal, outdir, out, goos, goarch string, envs []string) error {
	args := []string{
		"go",
		"build",
	}

	if ex.tags != "" {
		args = append(args, "-tags", ex.tags)
	}

	args = append(args,
		"-buildvcs=false",
		"-o", filepath.Join(outdir, out),
		"-ldflags", "-w -s",
		ex.entry,
	)

	buildEnv := append([]string{
		"GO111MODULE=on",
		fmt.Sprintf("GOOS=%s", goos),
		fmt.Sprintf("GOARCH=%s", goarch),
	}, envs...)

	if envValue(buildEnv, "GOFLAGS") == "" {
		buildEnv = append(buildEnv, "GOFLAGS=-mod=vendor")
	}

	msg, err := runEnv(args, buildEnv)
	if err != nil {
		return fmt.Errorf("failed to build external %s: %w, msg: %s", ex.name, err, string(msg))
	}

	return nil
}

func buildExternalInContainer(ex *dkexternal, outdir, out, goos, goarch string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}

	outputDir, err := filepath.Abs(outdir)
	if err != nil {
		return fmt.Errorf("filepath.Abs(%q): %w", outdir, err)
	}

	var containerOutDir string
	mountArgs := []string{"-v", projectRoot + ":/work"}
	if rel, relErr := filepath.Rel(projectRoot, outputDir); relErr == nil && !strings.HasPrefix(rel, "..") {
		containerOutDir = filepath.ToSlash(filepath.Join("/work", rel))
	} else {
		mountArgs = append(mountArgs, "-v", filepath.Dir(outputDir)+":/out")
		containerOutDir = "/out"
	}

	dockerCmd, err := resolveDockerCmd()
	if err != nil {
		return err
	}

	platform := os.Getenv("DK_BUILD_DOCKER_PLATFORM")
	if platform == "" {
		platform = "linux/" + runtime.GOARCH
	}

	buildImage := os.Getenv("DK_BUILD_ENV_IMAGE")
	if buildImage == "" {
		return fmt.Errorf("DK_BUILD_ENV_IMAGE is required")
	}

	containerCmd := fmt.Sprintf(
		"make --no-print-directory -C externals build_external_local "+
			"EXTERNAL_NAME=%s EXTERNAL_GOOS=%s EXTERNAL_ARCH=%s "+
			"EXTERNAL_OUTDIR=%s EXTERNAL_OUTPUT=%s GO_MODULE_MODE=%s%s%s",
		shQuote(ex.name),
		shQuote(goos),
		shQuote(goarch),
		shQuote(containerOutDir),
		shQuote(out),
		shQuote(goModuleMode()),
		ebpfKernelArg(ex.name, goarch),
		ebpfTargetArg(ex.name),
	)

	args := append([]string{}, dockerCmd...)
	args = append(args,
		"run", "--rm",
		"--platform", platform,
	)
	args = append(args, mountArgs...)
	args = append(args,
		"-w", "/work",
		buildImage,
		"bash", "-lc", containerCmd,
	)

	msg, err := runEnv(args, nil)
	if err != nil {
		return fmt.Errorf("failed to build external %s in container: %w, msg: %s", ex.name, err, string(msg))
	}

	chownTargets := []string{containerOutDir}
	if ex.name == externalEBPFName {
		chownTargets = append(chownTargets, filepath.ToSlash(filepath.Join(
			"/work",
			"internal/plugins/externals/ebpf/internal/c/elf",
			"linux_"+goarch,
		)))
	}

	chownArgs := append([]string{}, dockerCmd...)
	chownArgs = append(chownArgs,
		"run", "--rm",
		"--platform", platform,
	)
	chownArgs = append(chownArgs, mountArgs...)
	chownArgs = append(chownArgs,
		"-w", "/work",
		buildImage,
		"chown", "-R", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	)
	chownArgs = append(chownArgs, chownTargets...)

	_, _ = runEnv(chownArgs, nil)

	return nil
}

func resolveDockerCmd() ([]string, error) {
	if cmd := os.Getenv("DK_BUILD_DOCKER_CMD"); cmd != "" {
		return strings.Fields(cmd), nil
	}

	candidates := [][]string{
		{"docker"},
		{"sudo", "-n", "docker"},
	}

	for _, candidate := range candidates {
		args := append([]string{}, candidate...)
		args = append(args, "info")
		if _, err := runEnv(args, nil); err == nil {
			return candidate, nil
		}
	}

	return nil, fmt.Errorf("docker is not available for non-interactive use")
}

func goModuleMode() string {
	if mode := os.Getenv("GO_MODULE_MODE"); mode != "" {
		return mode
	}

	return "vendor"
}

func ebpfKernelArg(name, goarch string) string {
	if name != externalEBPFName {
		return ""
	}

	return " DK_BPF_KERNEL_SRC_PATH=" + shQuote("/usr/src/linux-headers-"+goarch)
}

func ebpfTargetArg(name string) string {
	if name != externalEBPFName {
		return ""
	}

	return " EXTERNAL_EBPF_TARGET='bpfobjs'"
}

func shQuote(s string) string {
	if s == "" {
		return "''"
	}

	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func envValue(envs []string, key string) string {
	prefix := key + "="
	for _, env := range envs {
		if strings.HasPrefix(env, prefix) {
			return strings.TrimPrefix(env, prefix)
		}
	}

	return ""
}

func BuidlExternals(dir, goos, goarch string, standalone bool) error {
	for _, ex := range externals {
		if err := doBuildExternal(ex, dir, goos, goarch, standalone); err != nil {
			return err
		}
	}

	return nil
}

func getProjectPrefix(str string) string {
	nIdx := strings.Index(str, "gitlab.jiagouyun.com/")
	if nIdx == -1 {
		return ""
	}

	return str[:nIdx]
}

func getProjectSuffix(str string) string {
	projectName := "/datakit/"
	nIdx := strings.Index(str, projectName)
	if nIdx == -1 {
		return ""
	}

	return str[nIdx+len(projectName):]
}
