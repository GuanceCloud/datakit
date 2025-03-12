// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package manager

import (
	"strings"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	plast "github.com/GuanceCloud/platypus/pkg/ast"
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

func normalizeStatus(status string) string {
	status = strings.ToLower(status)

	if s, ok := statusMap[status]; ok {
		status = s
	} else if status == "" {
		status = DefaultStatus
	}

	return status
}

func filterByStatus(stats string, filterRule []string) bool {
	for _, v := range filterRule {
		if strings.ToLower(v) == stats {
			return true
		}
	}
	return false
}

func ProcLoggingStatus(plpt ptinput.PlInputPt, disable bool, ignore []string) {
	status := DefaultStatus

	if s, _, err := plpt.Get(FieldStatus); err == nil {
		if s, ok := s.(string); ok {
			status = s
		}
	}

	if !disable {
		status = normalizeStatus(status)
		_ = plpt.Set(FieldStatus, status, plast.String)
	}

	if filterByStatus(status, ignore) {
		plpt.MarkDrop(true)
	}
}
