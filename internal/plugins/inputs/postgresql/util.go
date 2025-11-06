// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"fmt"
	"regexp"
)

var setTrimPattern = regexp.MustCompile(`(?i)(?:^(?:(?:\s*/\*.*?\*/)?\s*SET\b(?:[^';]*|(?:'[^']*')*)+;)+\s*(.+?)$)`)

// var setTrimPattern = regexp.MustCompile(` \s*\bSET\b.*;(.+)$`)

func TrimLeadingSetStmts(sql string) string {
	match := setTrimPattern.FindStringSubmatch(sql)
	if len(match) > 1 {
		return match[1]
	}
	return sql
}

type FilterConfig struct {
	Include []string
	Exclude []string
}

type Filter struct {
	includeRegexps []*regexp.Regexp
	excludeRegexps []*regexp.Regexp
}

func NewFilter(config FilterConfig) (*Filter, error) {
	var includeRegexps []*regexp.Regexp
	var excludeRegexps []*regexp.Regexp

	for _, pattern := range config.Include {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid include pattern %q: %w", pattern, err)
		}
		includeRegexps = append(includeRegexps, re)
	}

	for _, pattern := range config.Exclude {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
		}
		excludeRegexps = append(excludeRegexps, re)
	}

	return &Filter{
		includeRegexps: includeRegexps,
		excludeRegexps: excludeRegexps,
	}, nil
}

func (f *Filter) Allow(target string) bool {
	// Temporary debug logging
	for _, re := range f.excludeRegexps {
		if re.MatchString(target) {
			return false
		}
	}

	if len(f.includeRegexps) == 0 {
		return true
	}

	for _, re := range f.includeRegexps {
		if re.MatchString(target) {
			return true
		}
	}

	return false
}
