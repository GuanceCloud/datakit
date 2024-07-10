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
	"strconv"
	"strings"

	pr "github.com/shirou/gopsutil/v3/process"
)

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

	const scopeSuffix = ".scope"

	paths := strings.Split(items[2], "/")
	for _, path := range paths {
		if !strings.HasSuffix(path, scopeSuffix) {
			continue
		}
		scope := strings.TrimSuffix(path, scopeSuffix)

		lastHyphenIndex := strings.LastIndex(scope, "-")
		if lastHyphenIndex == -1 || lastHyphenIndex == len(scope)-1 {
			return "", fmt.Errorf("unexpected scope path")
		}

		return scope[lastHyphenIndex+1:], nil
	}

	return "", nil
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
