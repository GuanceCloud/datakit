// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import "regexp"

var setTrimPattern = regexp.MustCompile(`(?i)(?:^(?:(?:\s*/\*.*?\*/)?\s*SET\b(?:[^';]*|(?:'[^']*')*)+;)+\s*(.+?)$)`)

// var setTrimPattern = regexp.MustCompile(` \s*\bSET\b.*;(.+)$`)

func TrimLeadingSetStmts(sql string) string {
	match := setTrimPattern.FindStringSubmatch(sql)
	if len(match) > 1 {
		return match[1]
	}
	return sql
}
