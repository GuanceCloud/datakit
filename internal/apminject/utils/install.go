// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package utils

import (
	"debug/elf"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	DockerCtrPath = "/var/lib/docker/containers"
	GLibc         = "glibc"
	Muslc         = "musl"
)

var (
	reGLBC           = regexp.MustCompile(`ldd \(.*\) ([0-9\.]+)`)
	reMusl           = regexp.MustCompile("musl libc \\(.*\\)\nVersion ([0-9\\.]+)")
	soGLibcVerRegexp = regexp.MustCompile(`^GLIBC_([0-9\.]+)$`)
)

func LddInfo() (string, Version, error) {
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
		return Muslc, v[1], true
	} else {
		return GLibc, v[1], true
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

func RequiredGLIBCVersion(dynamicSymbols []elf.Symbol) (Version, error) {
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
