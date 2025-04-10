// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package utils contains utils
package utils

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrPyLibNotFound    = errors.New("ddtrace-run not found")
	ErrParseJavaVersion = errors.New(("failed to parse java version"))
	ErrJavaLibNotFound  = errors.New("dd-java-agent.jar not found")
	ErrUnsupportedJava  = errors.New(("unsupported java version"))
	ErrAlreadyInjected  = errors.New(("already injected"))
)

func GetJavaVersion(s string) (int, error) {
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return 0, ErrParseJavaVersion
	}

	idx := strings.Index(lines[0], "\"")
	if idx == -1 {
		return 0, ErrParseJavaVersion
	}
	idxTail := strings.LastIndex(lines[0], "\"")
	if idx == -1 {
		return 0, ErrParseJavaVersion
	}

	versionStr := lines[0][idx+1 : idxTail-1]
	li := strings.Split(versionStr, ".")
	if len(li) < 2 {
		return 0, ErrParseJavaVersion
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
