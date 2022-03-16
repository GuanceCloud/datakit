package worker

import (
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
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

func PPAddSatus(result *pipeline.Result, disable bool) string {
	if disable {
		if v, err := result.GetField(pipeline.PipelineStatusField); err == nil {
			if v, ok := v.(string); ok {
				return v
			}
		} else {
			return ""
		}
	}

	if v, err := result.GetField(pipeline.PipelineStatusField); err == nil {
		if v, ok := v.(string); ok {
			if s, ok := statusMap[strings.ToLower(v)]; ok {
				result.SetField(pipeline.PipelineStatusField, s)
				return s
			}
		}
	}
	result.SetField(pipeline.PipelineStatusField, pipeline.DefaultPipelineStatus)
	return pipeline.DefaultPipelineStatus
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
