// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package fileprovider wraps search files functions
package fileprovider

import (
	"fmt"
	"os"

	"github.com/bmatcuk/doublestar/v4"
)

type Scanner struct {
	patterns []string
}

func NewScanner(patterns []string) (*Scanner, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("patterns is empty")
	}
	return &Scanner{patterns}, nil
}

func (sc *Scanner) ScanFiles() ([]string, error) {
	var files []string

	for _, pattern := range sc.patterns {
		stat, err := os.Stat(pattern)
		if err == nil {
			if !stat.IsDir() {
				// The pattern is a file.
				files = append(files, pattern)
			}
		} else {
			paths, err := doublestar.FilepathGlob(pattern)
			if err != nil {
				return nil, err
			}

			files = append(files, paths...)
		}
	}

	return unique(files), nil
}

func unique(slice []string) []string {
	var res []string
	keys := make(map[string]interface{})
	for _, str := range slice {
		if _, ok := keys[str]; !ok {
			keys[str] = nil
			res = append(res, str)
		}
	}
	return res
}
