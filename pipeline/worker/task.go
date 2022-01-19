package worker

import (
	"fmt"
	"time"

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

type Result struct {
	output *parser.Output
}

func (r *Result) String() string {
	return fmt.Sprintf("%+#v", r.output) // FIXME: need more string format
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

func (r *Result) SetTime(t time.Time) {
	r.SetField("time", t.UnixNano())
}

func (r *Result) GetTag(k string) (string, error) {
	if v, ok := r.output.Tags[k]; ok {
		return v, nil
	} else {
		return "", fmt.Errorf("tag not found")
	}
}

func (r *Result) SetTag(k, v string) {
	r.output.Tags[k] = v
}

func (r *Result) DeleteTag(k string) {
	delete(r.output.Tags, k)
}

func (r *Result) GetField(k string) (interface{}, error) {
	if v, ok := r.output.Data[k]; ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("field not found")
	}
}

func (r *Result) SetField(k string, v interface{}) {
	r.output.Data[k] = v
}

func (r *Result) DeleteField(k string) {
	delete(r.output.Data, k)
}

func (r *Result) Drop() {
	r.output.Dropped = true
}

type TaskData interface {
	GetContent() string
	Handler(*Result) error
}

type Task struct {
	TaskName   string
	ScriptName string // 为空则根据 source 匹配对应的脚本
	Source     string // measurement name
	Data       []TaskData
	Opt        *TaskOpt
	TS         time.Time

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
