package worker

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

type TaskOpt struct {
	// 忽略这些status，如果数据的status在此列表中，数据将不再上传
	// ex: "info"
	//     "debug"

	Category string

	Version string

	IgnoreStatus []string

	// 是否关闭添加默认status字段列，包括status字段的固定转换行为，例如'd'->'debug'
	DisableAddStatusField bool
}

type TaskData interface {
	GetContent() string
	Handler(*Result) error
}

type Task struct {
	TaskName string

	Source     string // measurement name
	ScriptName string // 为空则根据 source 匹配对应的脚本

	Data []TaskData
	Opt  *TaskOpt
	TS   time.Time

	MaxMessageLen int
	// 保留字段
	Namespace string
}

func (task *Task) GetScriptName() string {
	if task.ScriptName != "" {
		return task.ScriptName
	} else {
		return task.Source + ".p"
	}
}

func FeedPipelineTask(task *Task) error {
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

func FeedPipelineTaskBlock(task *Task) error {
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

type Result struct {
	output      *parser.Output
	measurement string
	ts          time.Time

	err string
}

func (r *Result) String() string {
	return fmt.Sprintf("%+#v", r.output)
}

func NewResult() *Result {
	return &Result{
		output: &parser.Output{
			Tags: make(map[string]string),
			Data: make(map[string]interface{}),
			Cost: make(map[string]string),
		},
	}
}

func (r *Result) GetTag(k string) (string, error) {
	if v, ok := r.output.Tags[k]; ok {
		return v, nil
	} else {
		return "", fmt.Errorf("tag not found")
	}
}

func (r *Result) GetField(k string) (interface{}, error) {
	if v, ok := r.output.Data[k]; ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("field not found")
	}
}

func (r *Result) GetMeasurement() string {
	return r.measurement
}

func (r *Result) GetTime() (time.Time, error) {
	var ts time.Time
	if v, err := r.GetField(PipelineTimeField); err == nil {
		if nanots, ok := v.(int64); ok {
			return time.Unix(nanots/int64(time.Second),
				nanots%int64(time.Second)), nil
		}
	}
	if !r.ts.IsZero() {
		return r.ts, nil
	}
	return ts, fmt.Errorf("no time")
}

func (r *Result) GetTags() map[string]string {
	return r.output.Tags
}

func (r *Result) GetFields() map[string]interface{} {
	return r.output.Data
}

func (r *Result) GetLastErr() string {
	return r.err
}

func (r *Result) SetTime(t time.Time) {
	r.DeleteField(PipelineTimeField)
	r.ts = t
}

func (r *Result) SetTag(k, v string) {
	r.output.Tags[k] = v
}

func (r *Result) SetField(k string, v interface{}) {
	r.output.Data[k] = v
}

func (r *Result) DeleteField(k string) {
	delete(r.output.Data, k)
}

func (r *Result) DeleteTag(k string) {
	delete(r.output.Tags, k)
}

func (r *Result) IsDropped() bool {
	return r.output.Dropped
}

func (r *Result) MarkAsDropped() {
	r.output.Dropped = true
}

func (r *Result) checkFieldValLen(massageLen int) {
	spiltLen := massageLen
	if spiltLen == 0 { // 初始化task时候 没有注入最大长度 则使用默认值
		spiltLen = maxFieldsLength
	}

	if i, err := r.GetField(PipelineMessageField); err == nil {
		if mass, isString := i.(string); isString {
			if len(mass) > spiltLen {
				mass = mass[:spiltLen]
				r.SetField(PipelineMessageField, mass)
			}
		}
	}
}

type TaskDataTempl struct {
	Content string
	Tags    map[string]string
	Fields  map[string]interface{}
	TS      time.Time
}

func (data *TaskDataTempl) GetContent() string {
	return data.Content
}

func (data *TaskDataTempl) Handler(result *Result) error {
	for k, v := range data.Tags {
		if _, err := result.GetTag(k); err != nil {
			result.SetTag(k, v)
		}
	}

	for k, v := range data.Fields {
		if _, err := result.GetField(k); err != nil {
			result.SetField(k, v)
		}
	}

	if _, err := result.GetTime(); err != nil &&
		!data.TS.IsZero() {
		result.SetTime(data.TS)
	}

	return nil
}

func ParsePlScript(plScript string) (*parser.Engine, error) {
	if ng, err := parser.NewEngine(plScript, funcs.FuncsMap, funcs.FuncsCheckMap, false); err != nil {
		return nil, err
	} else {
		return ng, nil
	}
}

func RunAsPlTask(category string, source string, service string, content []string, ng *parser.Engine) []*Result {
	tags := map[string]string{}
	if service != "" {
		tags["service"] = service
	}

	task := Task{
		Source: source,
		Data:   make([]TaskData, 0),
		Opt: &TaskOpt{
			Category: category,
		},
	}
	for i := 0; i < len(content); i++ {
		task.Data = append(task.Data, &TaskDataTempl{
			Tags:    tags,
			Content: content[i],
		})
	}

	return RunPlTask(&task, ng)
}

func RunPlTask(task *Task, ng *parser.Engine) []*Result {
	ts := task.TS
	if ts.IsZero() {
		ts = time.Now()
	}
	taskResult := []*Result{}

	// 优先处理最新的数据，
	// 对无时间的数据根据 task 时间递减
	for di := len(task.Data) - 1; di >= 0; di-- {
		content := task.Data[di].GetContent()
		/*
			// 在pl之后再最切割
			if len(content) >= maxFieldsLength {
				content = content[:maxFieldsLength]
			}
		*/
		result := &Result{
			output: nil,
		}
		if ng != nil {
			if err := ng.Run(content); err != nil {
				l.Debug(err)
				result.err = err.Error()
			}
			result.output = ng.Result()
		} else {
			result.output = &parser.Output{
				Cost: map[string]string{},
				Tags: map[string]string{},
				Data: map[string]interface{}{
					PipelineMessageField: content,
				},
			}
		}

		if err := task.Data[di].Handler(result); err != nil {
			result.err = err.Error()
			result.MarkAsDropped()
		}

		ts = resultTailHandler(task, ts, result)
		taskResult = append(taskResult, result)
	}

	for r, l := 0, len(taskResult)-1; r < l; r, l = r+1, l-1 {
		taskResult[r], taskResult[l] = taskResult[l], taskResult[r]
	}
	return taskResult
}

func resultTailHandler(task *Task, ts time.Time, result *Result) time.Time {
	if rTS, err := result.GetTime(); err == nil {
		ts = rTS
	} else if result.ts.IsZero() {
		ts = ts.Add(-time.Nanosecond)
	}
	result.SetTime(ts)
	result.checkFieldValLen(task.MaxMessageLen)

	result.measurement = task.Source

	if task.Opt != nil {
		switch task.Opt.Category {
		case datakit.Logging, "":
			return processLoggingResult(task, ts, result)
		default:
		}
	} else {
		return processLoggingResult(task, ts, result)
	}
	return ts
}

func processLoggingResult(task *Task, ts time.Time, result *Result) time.Time {
	result.ts = ts

	if v, err := result.GetTag(PipelineMSource); err == nil {
		result.measurement = v
		result.DeleteTag(PipelineMSource)
	} else if v, err := result.GetField(PipelineMSource); err == nil {
		if v, ok := v.(string); ok {
			result.measurement = v
			result.DeleteField(PipelineMSource)
		}
	}

	// add status if disable == true;
	// ignore logs of a specific status.
	taskOpt := TaskOpt{}
	if task.Opt != nil {
		taskOpt = *task.Opt
	}
	status := PPAddSatus(result, taskOpt.DisableAddStatusField)
	if PPIgnoreStatus(status, taskOpt.IgnoreStatus) {
		result.MarkAsDropped()
	}
	return ts
}
