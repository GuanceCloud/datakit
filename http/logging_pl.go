package http

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	plw "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
)

type logTaskData struct {
	source   string
	version  string
	category string
	point    []*io.Point
}

func (std *logTaskData) ContentType() string {
	return plw.ContentString
}

func (std *logTaskData) ContentEncode() string {
	return ""
}

func (std *logTaskData) GetContentByte() [][]byte {
	return nil
}

func (std *logTaskData) GetContentStr() []string {
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

func (std *logTaskData) Callback(task *plw.Task, result []*plw.Result) error {
	result = plw.ResultUtilsLoggingProcessor(task, result, nil, nil)
	if len(result) != len(std.point) {
		return fmt.Errorf("result count is less than input")
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

			// no time exist in pipeline output, use origin line proto time
			if _, err := result[idx].GetTime(); err != nil {
				result[idx].SetTime(pt.Time())
			}
		} else {
			l.Warnf("get fields err=%v", err)
		}
	}

	return plw.ResultUtilsFeedIO(task, result)
}

func buildLogPLTask(input, source, version, category string, pts []*io.Point) *plw.Task {
	td := logTaskData{
		source:   source,
		version:  version,
		category: category,
		point:    pts,
	}

	return &plw.Task{
		TaskName: input,
		Source:   source,
		Opt: &plw.TaskOpt{
			DisableAddStatusField: true,
		},

		// ScriptName: not set here, pipeline will detect @source.p as the default pipline

		TS:   time.Now(),
		Data: &td,
	}
}
