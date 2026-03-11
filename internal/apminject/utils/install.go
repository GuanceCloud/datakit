// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package utils

import (
	"debug/elf"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
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

// DetectLibcType 检测系统的 libc 类型（musl 或 glibc）
// 通过检查动态链接器路径来判断
// glibc 使用 ld-linux-{arch}.so.*，musl 使用 ld-musl-{arch}.so.1
// 优先检查 glibc，因为大多数系统使用 glibc.
func DetectLibcType() string {
	// 根据架构确定链接器名称
	arch := "x86_64"
	if runtime.GOARCH == "arm64" {
		arch = "aarch64"
	}

	// 先检查 glibc 动态链接器（更常见）
	// glibc: /lib64/ld-linux-x86-64.so.2, /lib/ld-linux.so.2 (32-bit)
	glibcPaths := []string{
		fmt.Sprintf("/lib64/ld-linux-%s.so.2", arch),
		fmt.Sprintf("/lib/ld-linux-%s.so.2", arch),
		fmt.Sprintf("/lib/ld-linux-%s.so.1", arch),
	}
	for _, p := range glibcPaths {
		if _, err := os.Stat(p); err == nil {
			return GLibc
		}
	}

	// 检查 musl 动态链接器
	muslPaths := []string{
		fmt.Sprintf("/lib/ld-musl-%s.so.1", arch),
		fmt.Sprintf("/lib64/ld-musl-%s.so.1", arch),
	}
	for _, p := range muslPaths {
		if _, err := os.Stat(p); err == nil {
			return Muslc
		}
	}

	// 默认返回 glibc
	return GLibc
}

// DetectLibcTypeFromBinary 检测指定二进制文件链接的 libc 类型（musl 或 glibc）
// 通过 ldd 命令查看它依赖的动态库，更准确地判断该二进制需要哪个版本的 ddtrace.
func DetectLibcTypeFromBinary(binaryPath string) string {
	//nolint:gosec
	cmd := exec.Command("ldd", binaryPath)
	output, err := cmd.Output()
	if err != nil {
		// 如果 ldd 失败，回退到系统检测
		return DetectLibcType()
	}

	outputStr := string(output)

	// musl 的动态链接器路径
	if strings.Contains(outputStr, "ld-musl") {
		return Muslc
	}

	// glibc 的动态链接器路径
	if strings.Contains(outputStr, "ld-linux") {
		return GLibc
	}

	// 回退到系统检测
	return DetectLibcType()
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
