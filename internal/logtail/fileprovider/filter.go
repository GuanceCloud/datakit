// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package fileprovider

import (
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
)

type GlobFilter struct {
	include []string
	exclude []string
}

func NewGlobFilter(includePatterns, excludePatterns []string) (*GlobFilter, error) {
	for _, pattern := range includePatterns {
		if b := doublestar.ValidatePathPattern(pattern); !b {
			return nil, fmt.Errorf("invalid pattern: %s", pattern)
		}
	}

	for _, pattern := range excludePatterns {
		if b := doublestar.ValidatePathPattern(pattern); !b {
			return nil, fmt.Errorf("invalid pattern: %s", pattern)
		}
	}

	return &GlobFilter{
		include: includePatterns,
		exclude: excludePatterns,
	}, nil
}

func (f *GlobFilter) IncludeFilterFiles(files []string) []string {
	if len(f.include) == 0 {
		return files
	}

	var res []string

	for _, path := range files {
		for _, pattern := range f.include {
			if doublestar.MatchUnvalidated(pattern, path) {
				res = append(res, path)
				break
			}
		}
	}

	return res
}

func (f *GlobFilter) ExcludeFilterFiles(files []string) []string {
	if len(f.exclude) == 0 {
		return files
	}

	var res []string

	for _, path := range files {
		excluded := false
		for _, pattern := range f.exclude {
			if doublestar.MatchUnvalidated(pattern, path) {
				excluded = true
				break
			}
		}
		if !excluded {
			res = append(res, path)
		}
	}

	return res
}
