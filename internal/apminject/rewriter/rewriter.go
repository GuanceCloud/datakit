// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/utils"
)

const (
	recSep   = "\x1E"
	groupSep = "\x1D"

	filePrefix = "/tmp/dk_inject_rewrite_"

	ddtraceRun     = "ddtrace-run"
	ddtraceLibName = "dd-java-agent.jar"

	ddtraceJarPath = "/usr/local/datakit/apm_inject/lib/java/" + ddtraceLibName
)

var (
	pyRegexp   = regexp.MustCompile(`^python(?:3(?:[\.\d]+)*)*$`)
	javaRegexp = regexp.MustCompile(`^java$`)
)

var (
	errPyLibNotFound    = errors.New("ddtrace-run not found")
	errParseJavaVersion = errors.New(("failed to parse java version"))
	errJavaLibNotFound  = errors.New("dd-java-agent.jar not found")
	errUnsupportedJava  = errors.New(("unsupported java version"))
	errAlreadyInjected  = errors.New(("already injected"))
)

func main() {
	argv := os.Args // include tmpid, path and args
	envs := os.Environ()
	if len(argv) < 3 {
		return // error
	}

	tmpid := argv[0]
	ret, err := rewrite(&reArgs{
		path: argv[1],
		argv: argv[2:],
		envp: envs,
	})
	if err != nil {
		return
	}
	content := marshal(ret.path, ret.argv, ret.envp)

	//nolint:gosec
	if err := os.WriteFile(filePrefix+tmpid, []byte(content), 0o644); err != nil {
		_ = err
		return
	}
}

type reArgs struct {
	path string
	argv []string
	envp []string
}

func pyScriptWhitelist(path string) bool {
	_, exeName := filepath.Split(path)
	return strings.EqualFold(exeName, "flask") || strings.Contains(exeName, "gunicorn")
}

func rewrite(param *reArgs) (*reArgs, error) {
	exePath := filepath.Clean(param.path)
	_, exeName := filepath.Split(exePath)

	for _, env := range param.envp {
		if !strings.Contains(env, utils.EnvDKAPMINJECT) {
			continue
		}
		if v := strings.SplitN(env, "=", 2); len(v) == 2 {
			if utils.CheckDisableInjFromEnv(v[0], v[1]) {
				return nil, utils.ErrInjectDisabled
			}
		}
	}

	var pyScript bool
	if pyScriptWhitelist(exePath) {
		if s, err := pythonScriptMagic(exePath); err == nil {
			pyScript = true
			// python only
			exePath = s
		}
	}
	switch {
	case pyScript, pyRegexp.MatchString(exeName): // skip flask
		ddrun, err := checkPython(exePath, param.argv)
		if err != nil {
			return nil, err
		}

		ret := &reArgs{
			path: ddrun,
		}

		urlEnvs, err := traceURL(langPython)
		if err != nil {
			return nil, err
		}

		ret.argv = append(ret.argv, ddtraceRun)
		ret.argv = append(ret.argv, param.argv...)
		for _, v := range urlEnvs {
			ret.envp = append(ret.envp, strings.Join(v[:], "="))
		}
		ret.envp = append(ret.envp, param.envp...)
		return ret, nil
	case javaRegexp.MatchString(exeName):
		//nolint:gosec
		cmd := exec.Command(exePath, "-version")
		o, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}

		ver, err := getJavaVersion(string(o))
		if err != nil {
			return nil, err
		}

		if ver < 8 {
			return nil, errUnsupportedJava
		}

		urlEnvs, err := traceURL(langJava)
		if err != nil {
			return nil, err
		}

		for i := 1; i < len(param.argv); i++ {
			p := strings.TrimSpace(param.argv[i])
			if strings.HasPrefix(p, "-javaagent:") {
				if strings.Contains(p, ddtraceLibName) {
					return nil, errAlreadyInjected
				}
			}
		}

		var javaOpt string
		for _, v := range param.envp {
			p := strings.SplitN(v, "=", 2)
			if len(p) != 2 {
				continue
			}
			if p[0] != "JAVA_TOOL_OPTIONS" {
				continue
			}
			javaOpt = p[1]
			if strings.Contains(javaOpt, ddtraceLibName) {
				return nil, errAlreadyInjected
			}
		}

		finf, err := os.Stat(ddtraceJarPath)
		if err != nil {
			return nil, fmt.Errorf("stat dd-java-agent.jar: %w", err)
		}

		if finf.IsDir() || (finf.Mode().Perm()&0b100) != 0b100 {
			return nil, errJavaLibNotFound
		}

		ret := &reArgs{
			path: param.path,
			argv: append([]string{}, param.argv...),
		}

		newJavaOptEnv := fmt.Sprintf("JAVA_TOOL_OPTIONS=-javaagent:%s", ddtraceJarPath)
		if len(javaOpt) > 0 {
			newJavaOptEnv += " " + javaOpt
		}
		ret.envp = append(ret.envp, newJavaOptEnv)
		for _, v := range urlEnvs {
			ret.envp = append(ret.envp, strings.Join(v[:], "="))
		}
		ret.envp = append(ret.envp, param.envp...)
		return ret, nil
	}

	return nil, fmt.Errorf("skip rewrite exec %s", exePath)
}

const (
	langJava   = "java"
	langPython = "python"
)

func traceURL(lang string) (envs [][2]string, err error) {
	addr := utils.GetDKAddr()

	switch lang {
	case langJava:
		if addr.DkUds != "" {
			if addr.SDUds != "" {
				envs = [][2]string{
					{"DD_TRACE_AGENT_URL", "unix://" + addr.DkUds},
					{"DD_JMXFETCH_STATSD_HOST", addr.SDUds},
					{"DD_JMXFETCH_STATSD_PORT", "0"},
				}
			} else {
				envs = [][2]string{
					{"DD_TRACE_AGENT_URL", "unix://" + addr.DkUds},
				}
			}
			return
		} else {
			envs = [][2]string{
				{"DD_AGENT_HOST", addr.DkHost},
				{"DD_TRACE_AGENT_PORT", addr.DkPort},
				{"DD_JMXFETCH_STATSD_HOST", addr.SDHost},
				{"DD_JMXFETCH_STATSD_PORT", addr.SDPort},
			}
			return
		}
	case langPython:
		if addr.DkUds != "" {
			envs = [][2]string{
				{"DD_TRACE_AGENT_URL", "unix://" + addr.DkUds},
			}
			return
		} else {
			envs = [][2]string{
				{"DD_AGENT_HOST", addr.DkHost},
				{"DD_AGENT_PORT", addr.DkPort},
			}
			return
		}
	}

	return nil, fmt.Errorf("unsupported language")
}

func marshal(path string, args, envp []string) string {
	var s string
	s += "1" + recSep
	s += path + recSep
	s += groupSep + recSep

	s += strconv.Itoa(len(args)) + recSep
	for _, v := range args {
		s += v + recSep
	}
	s += groupSep + recSep

	s += strconv.Itoa(len(envp)) + recSep
	for _, v := range envp {
		s += v + recSep
	}
	s += groupSep + recSep

	return s
}

func getJavaVersion(s string) (int, error) {
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return 0, errParseJavaVersion
	}

	idx := strings.Index(lines[0], "\"")
	if idx == -1 {
		return 0, errParseJavaVersion
	}
	idxTail := strings.LastIndex(lines[0], "\"")
	if idx == -1 {
		return 0, errParseJavaVersion
	}

	versionStr := lines[0][idx+1 : idxTail-1]
	li := strings.Split(versionStr, ".")
	if len(li) < 2 {
		return 0, errParseJavaVersion
	}

	v, err := strconv.Atoi(li[0])
	if err != nil {
		return 0, err
	}

	if v == 1 {
		v, err = strconv.Atoi(li[1])
		if err != nil {
			return 0, err
		}
		return v, nil
	} else {
		return v, nil
	}
}

var regexpPythonMagic = regexp.MustCompile("^#!(/.*/python(?:3(?:[\\.\\d]+)*)?)\n$")

func pythonScriptMagic(fp string) (string, error) {
	fp, err := exec.LookPath(fp)
	if fp == "" || err != nil {
		return "", fmt.Errorf("not found %s", fp)
	}

	if fp == "ddtrace-run" {
		return "", fmt.Errorf("cannot inject ddtrace-run")
	}
	if strings.HasSuffix(fp, "/ddtrace-run") {
		return "", fmt.Errorf("cannot inject ddtrace-run")
	}

	f, err := os.Open(fp) // nolint:gosec
	if err != nil {
		return "", err
	}
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return "", err
	}
	if n <= 0 {
		return "", fmt.Errorf("read python magic fail")
	}
	buf = buf[:n]

	n = bytes.Index(buf, []byte("\n"))
	if n == -1 {
		return "", fmt.Errorf("read python magic fail")
	}

	buf = buf[:n+1]
	val := regexpPythonMagic.FindStringSubmatch(string(buf))
	if len(val) != 2 {
		return "", fmt.Errorf("read python magic fail")
	}

	return val[1], nil
}

func checkPython(pyBinPath string, argv []string) (string, error) {
	//nolint:gosec
	// require ddtrace, python3
	cmd := exec.Command(pyBinPath,
		"-c", "import ddtrace; print(ddtrace.__version__)")

	if _, err := cmd.Output(); err != nil {
		return "", fmt.Errorf("get ddtrace-run version error: %w", err)
	}

	ddrun, err := exec.LookPath(ddtraceRun)
	if err != nil {
		return "", errPyLibNotFound
	}
	finf, err := os.Stat(ddrun)
	if err != nil {
		return "", errPyLibNotFound
	}

	if finf.IsDir() || (finf.Mode().Perm()&0b1) != 0b1 {
		return "", fmt.Errorf("ddtrace-run is not executable")
	}

	for _, v := range argv {
		if strings.ToLower(strings.TrimSpace(v)) == "-m" {
			return "", fmt.Errorf("python arg -m is not supported")
		}
	}
	return ddrun, nil
}
