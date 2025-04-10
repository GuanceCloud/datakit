// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"debug/elf"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/utils"
)

func TestDKAddr(t *testing.T) {
	utils.GetDKAddr()
	t.Setenv(utils.EnvDKSocketAddr, utils.DefaultDKUDS)
	traceURL(langJava)
	traceURL(langPython)
}

var (
	javaVer8 = `java version "1.8.12_xx" 2024-07-16
OpenJDK Runtime Environment (build 17.0.12+7-Ubuntu-1ubuntu222.04)
OpenJDK 64-Bit Server VM (build 17.0.12+7-Ubuntu-1ubuntu222.04, mixed mode, sharing)`

	javaVer17 = `openjdk version "17.0.12" 2024-07-16
OpenJDK Runtime Environment (build 17.0.12+7-Ubuntu-1ubuntu222.04)
OpenJDK 64-Bit Server VM (build 17.0.12+7-Ubuntu-1ubuntu222.04, mixed mode, sharing)`
)

func TestRegexp(t *testing.T) {
	assert.True(t, pyRegexp.MatchString("python3.7"))
	assert.True(t, pyRegexp.MatchString("python3"))
	assert.True(t, pyRegexp.MatchString("python"))

	assert.False(t, pyRegexp.MatchString("python2"))

	assert.True(t, javaRegexp.MatchString("java"))
	assert.False(t, javaRegexp.MatchString("java-"))

	ver, err := getJavaVersion(javaVer8)
	assert.NoError(t, err)
	assert.Equal(t, 8, ver)

	ver, err = getJavaVersion(javaVer17)
	assert.NoError(t, err)
	assert.Equal(t, 17, ver)

	assert.True(t, pyScriptWhitelist("/home/xx/flask"))
	assert.True(t, pyScriptWhitelist("/home/xx/gunicorn"))
	assert.False(t, pyScriptWhitelist("/home/xx/abcdefg"))

	v, err := DetectArch("/bin/sh")
	t.Log(v, err)
}

func DetectArch(filePath string) (string, error) {
	f, err := elf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close() //nolint:errcheck

	switch f.Machine { //nolint:exhaustive
	case elf.EM_AARCH64:
		return "arm64", nil
	case elf.EM_X86_64:
		return "amd64", nil
	default:
		return "", fmt.Errorf("unsupported arch: %v", f.Machine)
	}
}
