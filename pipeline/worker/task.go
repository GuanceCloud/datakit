// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package worker

import (
	"fmt"
	"reflect"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

const (
	ContentString = "string"
	ContentByte   = "byte"
)

type Task interface {
	GetSource() string
	GetScriptName() string // 待调用的 pipeline 脚本
	GetMaxMessageLen() int
	ContentType() string // TaskDataString or TaskDataByte
	ContentEncode() string
	GetContent() interface{} // []string or [][]byte

	// feed 给 pipeline 时 pl worker 会调用此方法
	Callback([]*pipeline.Result) error
}

func TaskDataContentType(task Task) string {
	defer func() {
		if err := recover(); err != nil {
			l.Error(err)
		}
	}()
	return task.ContentType()
}

func TaskDataGetContentStr(data Task) (value []string, err error) {
	defer func() {
		if errR := recover(); errR != nil {
			err = fmt.Errorf("%w", errR)
		}
	}()
	cnt := data.GetContent()
	if d, ok := cnt.([]string); ok {
		value = d
		return
	} else {
		return nil, fmt.Errorf("unsupported type '%v'", reflect.TypeOf(cnt))
	}
}

func TaskDataGetContentByte(data Task) (value [][]byte, err error) {
	defer func() {
		if errR := recover(); errR != nil {
			err = fmt.Errorf("%w", errR)
		}
	}()
	cnt := data.GetContent()
	if d, ok := cnt.([][]byte); ok {
		value = d
		return
	} else {
		return nil, fmt.Errorf("unsupported type '%v'", reflect.TypeOf(cnt))
	}
}

func TaskDataContentEncode(data Task) string {
	defer func() {
		if err := recover(); err != nil {
			l.Error(err)
		}
	}()
	return data.ContentEncode()
}

type TaskTemplate struct {
	Category string

	Version string

	Source     string // measurement name
	ScriptName string // 为空则根据 source 匹配对应的脚本

	TaskName string

	TS time.Time

	MaxMessageLen int

	IgnoreStatus []string

	// 是否关闭添加默认status字段列，包括status字段的固定转换行为，例如'd'->'debug'
	DisableAddStatusField bool

	ContentDataType string
	Encode          string
	Content         interface{}

	Tags   map[string]string
	Fields map[string]interface{}
}

func (data *TaskTemplate) GetSource() string {
	return data.Source
}

func (data *TaskTemplate) GetScriptName() string {
	if data.ScriptName != "" {
		return data.ScriptName
	} else {
		return data.Source + ".p"
	}
}

func (data *TaskTemplate) GetMaxMessageLen() int {
	return data.MaxMessageLen
}

func (data *TaskTemplate) ContentType() string {
	if data.ContentDataType == "" {
		return ContentString
	}
	return data.ContentDataType
}

func (data *TaskTemplate) GetContent() interface{} {
	return data.Content
}

func (data *TaskTemplate) ContentEncode() string {
	return data.Encode
}

func (data *TaskTemplate) Callback(result []*pipeline.Result) error {
	result = ResultUtilsLoggingProcessor(result, data.Tags, data.Fields,
		data.DisableAddStatusField, data.IgnoreStatus)
	ts := data.TS
	if ts.IsZero() {
		ts = time.Now()
	}
	result = ResultUtilsAutoFillTime(result, ts)
	return ResultUtilsFeedIO(result, data.Category, data.Version, data.Source, data.TaskName, data.MaxMessageLen)
}

func FeedPipelineTask(task Task) error {
	if task == nil {
		return nil
	}
	if wkrManager == nil || taskCh == nil || stopCh == nil {
		return ErrTaskChNotReady
	} else {
		taskChFeedNumIncrease()
		defer taskChFeedNumDecrease()

		select {
		case <-stopCh:
			return ErrTaskChClosed
		case taskCh <- task:
			return nil
		default:
			return ErrTaskBusy
		}
	}
}

func FeedPipelineTaskBlock(task Task) error {
	if task == nil {
		return nil
	}
	if wkrManager == nil || taskCh == nil || stopCh == nil {
		return ErrTaskChNotReady
	} else {
		taskChFeedNumIncrease()
		defer taskChFeedNumDecrease()

		select {
		case <-stopCh:
			return ErrTaskChClosed
		case taskCh <- task:
			return nil
		}
	}
}

func RunAsPlTask(category string, source string, service, dataType string,
	content interface{}, encode string, ng *parser.Engine) []*pipeline.Result {
	tags := map[string]string{}
	if service != "" {
		tags["service"] = service
	}

	task := &TaskTemplate{
		Source:          source,
		Encode:          encode,
		ContentDataType: dataType,
		Content:         content,
		Category:        category,
	}

	result, _ := RunPlTask(task, ng)
	ResultUtilsLoggingProcessor(result, tags, nil, task.DisableAddStatusField, task.IgnoreStatus)
	ts := task.TS
	if ts.IsZero() {
		ts = time.Now()
	}
	result = ResultUtilsAutoFillTime(result, ts)
	return result
}

func ParsePlScript(plScript string) (*parser.Engine, error) {
	if ng, err := parser.NewEngine(plScript, funcs.FuncsMap, funcs.FuncsCheckMap, false); err != nil {
		return nil, err
	} else {
		return ng, nil
	}
}

func RunPlTask(task Task, ng *parser.Engine) ([]*pipeline.Result, error) {
	taskResult := []*pipeline.Result{}

	cntType := TaskDataContentType(task)
	switch cntType {
	case ContentByte:
		cntByte, err := TaskDataGetContentByte(task)
		if err != nil {
			return nil, err
		}
		encode := TaskDataContentType(task)
		for _, cnt := range cntByte {
			if r, err := pipeline.RunPlByte(cnt, encode, task.GetSource(), task.GetMaxMessageLen(), ng); err != nil {
				l.Debug(err)
				continue
			} else {
				taskResult = append(taskResult, r)
			}
		}
	default:
		cntStr, err := TaskDataGetContentStr(task)
		if err != nil {
			return nil, err
		}
		for _, cnt := range cntStr {
			if r, err := pipeline.RunPlStr(cnt, task.GetSource(),
				task.GetMaxMessageLen(), ng); err != nil {
				l.Debug(err)
				continue
			} else {
				taskResult = append(taskResult, r)
			}
		}
	}

	return taskResult, nil
}

func ResultUtilsAutoFillTime(result []*pipeline.Result, lastTime time.Time) []*pipeline.Result {
	for di := len(result) - 1; di >= 0; di-- {
		rTS, err := result[di].GetTime()
		if err != nil {
			lastTime = lastTime.Add(-time.Nanosecond)
		} else {
			lastTime = rTS
		}
		result[di].SetTime(lastTime)
	}
	return result
}

func ResultUtilsLoggingProcessor(result []*pipeline.Result, tags map[string]string, fields map[string]interface{},
	disableAddStatusField bool, ignoreStatus []string) []*pipeline.Result {
	for _, res := range result {
		for k, v := range tags {
			if _, err := res.GetTag(k); err != nil {
				res.SetTag(k, v)
			}
		}
		for k, v := range fields {
			if _, err := res.GetField(k); err != nil {
				res.SetField(k, v)
			}
		}
		status := PPAddSatus(res, disableAddStatusField)
		if PPIgnoreStatus(status, ignoreStatus) {
			res.MarkAsDropped()
		}
	}

	return result
}

func ResultUtilsFeedIO(result []*pipeline.Result, category, version, source, feedName string, maxMessageLen int) error {
	if len(result) == 0 {
		return nil
	}
	if category == "" {
		category = datakit.Logging
	}
	if source == "" {
		source = "default"
	}

	pts := []*io.Point{}
	for _, v := range result {
		if v.IsDropped() {
			continue
		}
		if pt, err := v.MakePoint(source, maxMessageLen, category); err != nil {
			l.Error(err)
		} else {
			pts = append(pts, pt)
		}
	}
	return io.Feed(feedName, category, pts,
		&io.Option{
			HighFreq: disableHighFreqIODdata,
			Version:  version,
		})
}
