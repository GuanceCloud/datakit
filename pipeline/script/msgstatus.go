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
	PipelineTimeField     = "time"
	PipelineMessageField  = "message"
	PipelineStatusField   = "status"
	PipelineMSource       = "source"
	PipelineDefaultStatus = "unknown"
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

func ProcLoggingStatus(output *parser.Output, disable bool, ignore []string, spiltLen int) *parser.Output {
	var status string

	if v, ok := output.Fields[PipelineStatusField]; ok {
		if v, ok := v.(string); ok {
			status = strings.ToLower(v)
		}
	}

	if !disable {
		status = strings.ToLower(status)
		if s, ok := statusMap[status]; ok {
			status = s
		} else {
			status = PipelineDefaultStatus
			output.Fields[PipelineStatusField] = status
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

	if spiltLen <= 0 { // 当初始化 task 时没有注入最大长度则使用默认值
		spiltLen = maxFieldsLength
	}
	for key := range output.Fields {
		if i, ok := output.Fields[key]; ok {
			if mass, isString := i.(string); isString {
				if len(mass) > spiltLen {
					mass = mass[:spiltLen]
					output.Fields[key] = mass
				}
			}
		}
	}

	return output
}
