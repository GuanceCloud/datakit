// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package process

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	pr "github.com/shirou/gopsutil/v3/process"
)

// Container IDs should not have a length limit, see this issue https://github.com/opencontainers/runc/issues/1429
// But there is a limit of 64 here.
var containerIDregex = regexp.MustCompile(`^[\w+-\.]{64}$`)

func getContainerID(ps *pr.Process) string {
	s, err := readCgroupFile(int(ps.Pid))
	if err != nil {
		l.Warnf("cannot open cgroup file %s", err)
		return ""
	}

	id, err := parseScopePathForCgroup(s)
	if err != nil {
		l.Warnf("cannot parse cgroup file %s", err)
		return ""
	}

	return id
}

func parseScopePathForCgroup(s string) (string, error) {
	lines := strings.Split(s, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("cannot parse cgroup content")
	}

	// like hierarchy-ID:subsystem:cgroup_paths
	items := strings.Split(strings.TrimSpace(lines[0]), ":")
	if len(items) != 3 {
		return "", fmt.Errorf("cannot parse cgroup items")
	}

	content := items[2]

	if id := parseContainerIDForUbuntuCgroup(content); id != "" && containerIDregex.MatchString(id) {
		return id, nil
	}

	if id := parseContainerIDForCentOSCgroup(content); id != "" && containerIDregex.MatchString(id) {
		return id, nil
	}

	// This process was not created by the container and does not return an error.
	return "", nil
}

func parseContainerIDForUbuntuCgroup(content string) string {
	const scopeSuffix = ".scope"

	paths := strings.Split(content, "/")
	for _, path := range paths {
		if !strings.HasSuffix(path, scopeSuffix) {
			continue
		}
		scope := strings.TrimSuffix(path, scopeSuffix)

		lastHyphenIndex := strings.LastIndex(scope, "-")
		if lastHyphenIndex == -1 || lastHyphenIndex == len(scope)-1 {
			return ""
		}

		return scope[lastHyphenIndex+1:]
	}

	return ""
}

func parseContainerIDForCentOSCgroup(content string) string {
	paths := strings.Split(content, "/")
	return paths[len(paths)-1]
}

// Get cgroup from /proc/(pid)/cgroup.
func readCgroupFile(pid int) (string, error) {
	cgroup, err := os.ReadFile(hostProc(strconv.Itoa(pid), "cgroup"))
	if err != nil {
		return "", err
	}
	return string(cgroup), nil
}

func hostProc(combineWith ...string) string {
	procPath := os.Getenv("HOST_PROC")
	if procPath == "" {
		procPath = "/proc"
	}

	switch len(combineWith) {
	case 0:
		return procPath
	case 1:
		return filepath.Join(procPath, combineWith[0])
	default:
		all := make([]string, len(combineWith)+1)
		all[0] = procPath
		copy(all[1:], combineWith)
		return filepath.Join(all...)
	}
}
