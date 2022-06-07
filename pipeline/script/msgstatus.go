// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package script

import (
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
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

func ProcLoggingStatus(output *parser.Output, disable bool, ignore []string) *parser.Output {
	var status string

	status, _ = getStatus(output)

	if !disable {
		status = strings.ToLower(status)
		if s, ok := statusMap[status]; ok {
			status = s
			setStatus(output, status)
		} else {
			status = DefaultStatus
			setStatus(output, status)
		}
	}

	if len(ignore) > 0 {
		for _, ign := range ignore {
			if strings.ToLower(ign) == status {
				output.Drop = true
				break
			}
		}
	}
	return output
}

func getStatus(output *parser.Output) (string, bool) {
	if v, ok := output.Tags[FieldStatus]; ok {
		return v, ok
	}

	if v, ok := output.Fields[FieldStatus]; ok {
		if s, ok := v.(string); ok {
			return s, ok
		}
	}

	return "", false
}

func setStatus(output *parser.Output, status string) {
	if _, ok := output.Tags[FieldStatus]; ok {
		output.Tags[FieldStatus] = status
		return
	}

	output.Fields[FieldStatus] = status
}
