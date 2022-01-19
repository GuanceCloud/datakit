package worker

import (
	"strings"
)

const (
	// pipeline关键字段.
	PipelineTimeField     = "time"
	PipelineMessageField  = "message"
	PipelineStatusField   = "status"
	PipelineMSource       = "source"
	DefaultPipelineStatus = "info"
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

func PPAddSatus(result *Result, disable bool) string {
	if disable {
		if v, err := result.GetField(PipelineStatusField); err == nil {
			if v, ok := v.(string); ok {
				return v
			}
		} else {
			return ""
		}
	}

	if v, err := result.GetField(PipelineStatusField); err == nil {
		if v, ok := v.(string); ok {
			if s, ok := statusMap[strings.ToLower(v)]; ok {
				result.SetField(PipelineStatusField, s)
				return s
			}
		}
	}
	result.SetField(PipelineStatusField, DefaultPipelineStatus)
	return DefaultPipelineStatus
}

// PPIgnoreStatus 过滤指定status.
func PPIgnoreStatus(status string, ignoreStatus []string) bool {
	if len(ignoreStatus) == 0 {
		return false
	}

	s := strings.ToLower(status)

	for _, ignore := range ignoreStatus {
		if strings.ToLower(ignore) == s {
			return true
		}
	}
	return false
}
