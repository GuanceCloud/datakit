// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package script

import (
	"strings"
)

const (
	// pipeline关键字段.
	FieldTime       = "time"
	FieldMessage    = "message"
	FieldStatus     = "status"
	PlLoggingSource = "source"

	DefaultStatus = "unknown"
)

var statusMap = map[string]string{
	"f":        "emerg",
	"emerg":    "emerg",
	"a":        "alert",
	"alert":    "alert",
	"c":        "critical",
	"critical": "critical",
	"e":        "error",
	"error":    "error",
	"w":        "warning",
	"warn":     "warning",
	"warning":  "warning",
	"n":        "notice",
	"notice":   "notice",
	"i":        "info",
	"info":     "info",
	"d":        "debug",
	"trace":    "debug",
	"verbose":  "debug",
	"debug":    "debug",
	"o":        "OK",
	"s":        "OK",
	"ok":       "OK",
}

func ProcLoggingStatus(tags map[string]string, fileds map[string]any, drop bool,
	disable bool, ignore []string,
) (map[string]string, map[string]any, bool) {
	var status string

	status, _ = getStatus(tags, fileds)

	if !disable {
		status = strings.ToLower(status)
		if s, ok := statusMap[status]; ok {
			status = s
			tags, fileds = setStatus(tags, fileds, status)
		} else {
			status = DefaultStatus
			tags, fileds = setStatus(tags, fileds, status)
		}
	}

	if len(ignore) > 0 {
		for _, ign := range ignore {
			if strings.ToLower(ign) == status {
				drop = true
				break
			}
		}
	}
	return tags, fileds, drop
}

func getStatus(tags map[string]string, fields map[string]any) (string, bool) {
	if v, ok := tags[FieldStatus]; ok {
		return v, ok
	}

	if v, ok := fields[FieldStatus]; ok {
		if s, ok := v.(string); ok {
			return s, ok
		}
	}

	return "", false
}

func setStatus(tags map[string]string, fileds map[string]any, status string) (
	map[string]string, map[string]any,
) {
	if _, ok := tags[FieldStatus]; ok {
		tags[FieldStatus] = status
	} else {
		fileds[FieldStatus] = status
	}
	return tags, fileds
}
