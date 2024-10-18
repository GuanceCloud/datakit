// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package filepathtoolkit is a collection of filepath utils.
package filepathtoolkit

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	currentOsPathSep = string(os.PathSeparator)
	otherOsPathSep   = otherOsPathSeparator()
)

func otherOsPathSeparator() string {
	if os.PathSeparator == '/' {
		return "\\"
	}
	return "/"
}

func BaseName(path string) string {
	if !strings.Contains(path, currentOsPathSep) && strings.Contains(path, otherOsPathSep) {
		path = strings.ReplaceAll(path, otherOsPathSep, currentOsPathSep)
	}
	return filepath.Base(path)
}

func DirName(path string) string {
	dir := filepath.Dir(path)

	// for other platforms
	if dir == "." && !strings.Contains(path, currentOsPathSep) {
		idx := strings.LastIndex(path, otherOsPathSep)
		if idx >= 0 {
			if idx == 0 || path[idx-1] == ':' {
				return path[:idx+1]
			}
			return path[:idx]
		}
	}
	if dir == "." && len(path) >= 2 && path[0] == '<' && path[len(path)-1] == '>' {
		return path
	}
	return dir
}
