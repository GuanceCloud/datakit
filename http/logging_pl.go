// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	plw "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
)

type logTask struct {
	taskName              string
	disableAddStatusField bool
	source                string
	version               string
	category              string
	point                 []*io.Point
	ts                    time.Time
}

func (std *logTask) GetScriptName() string {
	// TODO
	return std.source + ".p"
}

func (std *logTask) GetSource() string {
	return std.source
}

func (std *logTask) ContentType() string {
	return plw.ContentString
}

func (std *logTask) ContentEncode() string {
	return ""
}

func (std *logTask) GetContent() interface{} {
	cntStr := []string{}
	for _, pt := range std.point {
		fields, err := pt.Fields()
		if err != nil {
			cntStr = append(cntStr, "")
			continue
		}
		if len(fields) > 0 {
			if msg, ok := fields["message"]; ok {
				switch msg := msg.(type) {
				case string:
					cntStr = append(cntStr, msg)
					continue
				default:
					cntStr = append(cntStr, "")
					continue
				}
			}
		}
		cntStr = append(cntStr, "")
	}
	return cntStr
}

func (std *logTask) GetMaxMessageLen() int {
	return 0
}

func (std *logTask) callback(result []*pipeline.Result) ([]*pipeline.Result, error) {
	result = plw.ResultUtilsLoggingProcessor(result, nil, nil, std.disableAddStatusField, nil)
	if len(result) != len(std.point) {
		return nil, fmt.Errorf("result count is less than input")
	}
	for idx, pt := range std.point {
		tags := pt.Tags()
		for k, v := range tags {
			result[idx].SetTag(k, v)
		}
		fields, err := pt.Fields()
		if err == nil {
			for k, i := range fields {
				result[idx].SetField(k, i)
			}
		} else {
			l.Warnf("get fields err=%v", err)
		}
		// no time exist in pipeline output, use origin line proto time
		if _, err := result[idx].GetTime(); err != nil {
			result[idx].SetTime(pt.Time())
		}
	}
	return result, nil
}

func (std *logTask) Callback(result []*pipeline.Result) error {
	result, err := std.callback(result)
	if err != nil {
		return err
	}
	return plw.ResultUtilsFeedIO(result, std.category, std.version, std.source, std.taskName, 0)
}

func buildLogPLTask(input, source, version, category string, pts []*io.Point) *logTask {
	return &logTask{
		taskName:              input,
		source:                source,
		version:               version,
		category:              category,
		disableAddStatusField: true,
		point:                 pts,
		ts:                    time.Now(),
	}
}
