package http

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	plw "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
)

type logTaskData struct {
	source   string
	version  string
	category string
	point    *io.Point
}

func (std *logTaskData) GetContent() string {
	fields, err := std.point.Fields()
	if err != nil {
		return ""
	}
	if len(fields) > 0 {
		if msg, ok := fields["message"]; ok {
			switch msg := msg.(type) {
			case string:
				return msg
			default:
				return ""
			}
		}
	}
	return ""
}

func (std *logTaskData) Handler(result *plw.Result) error {
	tags := std.point.Tags()
	for k, v := range tags {
		result.SetTag(k, v)
	}
	fields, err := std.point.Fields()
	if err == nil {
		for k, i := range fields {
			result.SetField(k, i)
		}

		// no time exist in pipeline output, use origin line proto time
		if _, err := result.GetField("time"); err != nil {
			result.SetTime(std.point.Time())
		}
	} else {
		l.Warnf("get fields err=%v", err)
	}
	return err
}

func buildLogPLTask(input, source, version, category string, pts []*io.Point) *plw.Task {
	var td []plw.TaskData
	for _, pt := range pts {
		td = append(td, &logTaskData{
			source:   source,
			version:  version,
			category: category,
			point:    pt,
		})
	}

	return &plw.Task{
		TaskName: input,
		Source:   source,
		Opt: &plw.TaskOpt{
			DisableAddStatusField: true,
		},

		//ScriptName: not set here, pipeline will detect @source.p as the default pipline

		TS:   time.Now(),
		Data: td,
	}
}
