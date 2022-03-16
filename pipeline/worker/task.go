package worker

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
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

const (
	ContentString = "string"
	ContentByte   = "byte"
)

type TaskData interface {
	ContentType() string // TaskDataString or TaskDataByte

	GetContentStr() []string
	GetContentByte() [][]byte
	ContentEncode() string

	// feed 给 pipeline 时 pl worker 会调用此方法
	Callback(*Task, []*pipeline.Result) error
}

func TaskDataContentType(data TaskData) string {
	defer func() {
		if err := recover(); err != nil {
			l.Error(err)
		}
	}()
	return data.ContentType()
}

func TaskDataGetContentStr(data TaskData) []string {
	defer func() {
		if err := recover(); err != nil {
			l.Error(err)
		}
	}()
	return data.GetContentStr()
}

func TaskDataGetContentByte(data TaskData) [][]byte {
	defer func() {
		if err := recover(); err != nil {
			l.Error(err)
		}
	}()
	return data.GetContentByte()
}

func TaskDataContentEncode(data TaskData) string {
	defer func() {
		if err := recover(); err != nil {
			l.Error(err)
		}
	}()
	return data.ContentEncode()
}

type TaskDataTemplate struct {
	ContentDataType string
	Encode          string
	ContentStr      []string
	ContentByte     [][]byte

	Tags   map[string]string
	Fields map[string]interface{}
}

func (data *TaskDataTemplate) ContentType() string {
	return data.ContentDataType
}

func (data *TaskDataTemplate) GetContentStr() []string {
	return data.ContentStr
}

func (data *TaskDataTemplate) GetContentByte() [][]byte {
	return data.ContentByte
}

func (data *TaskDataTemplate) ContentEncode() string {
	return data.Encode
}

func (data *TaskDataTemplate) Callback(task *Task, result []*pipeline.Result) error {
	result = ResultUtilsLoggingProcessor(task, result, data.Tags, data.Fields)
	ts := task.TS
	if ts.IsZero() {
		ts = time.Now()
	}
	result = ResultUtilsAutoFillTime(result, ts)
	return ResultUtilsFeedIO(task, result)
}

type Task struct {
	TaskName string

	Source     string // measurement name
	ScriptName string // 为空则根据 source 匹配对应的脚本

	Opt *TaskOpt

	Data TaskData

	TS time.Time

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

func RunAsPlTask(category string, source string, service, dataType string,
	contentStr []string, contentByte [][]byte, encode string, ng *parser.Engine) []*pipeline.Result {
	tags := map[string]string{}
	if service != "" {
		tags["service"] = service
	}

	task := Task{
		Source: source,
		Data: &TaskDataTemplate{
			ContentDataType: dataType,
			ContentStr:      contentStr,
			ContentByte:     contentByte,
			Encode:          encode,
		},
		Opt: &TaskOpt{
			Category: category,
		},
	}

	result, _ := RunPlTask(&task, ng)
	ResultUtilsLoggingProcessor(&task, result, tags, nil)
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

func RunPlTask(task *Task, ng *parser.Engine) ([]*pipeline.Result, error) {
	taskResult := []*pipeline.Result{}

	encode := TaskDataContentEncode(task.Data)
	cntType := TaskDataContentType(task.Data)
	switch cntType {
	case ContentByte:
		cntByte := TaskDataGetContentByte(task.Data)
		for _, cnt := range cntByte {
			if r, err := pipeline.RunPlByte(cnt, encode, task.Source, task.MaxMessageLen, ng); err != nil {
				l.Debug(err)
				continue
			} else {
				taskResult = append(taskResult, r)
			}
		}
	case ContentString, "":
		cntStr := TaskDataGetContentStr(task.Data)
		for _, cnt := range cntStr {
			if r, err := pipeline.RunPlStr(cnt, task.Source, task.MaxMessageLen, ng); err != nil {
				l.Debug(err)
				continue
			} else {
				taskResult = append(taskResult, r)
			}
		}
	default:
		return nil, ErrContentType
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

func ResultUtilsLoggingProcessor(task *Task, result []*pipeline.Result, tags map[string]string, fields map[string]interface{}) []*pipeline.Result {
	opt := task.Opt
	if opt == nil {
		opt = &TaskOpt{}
	}
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
		status := PPAddSatus(res, opt.DisableAddStatusField)
		if PPIgnoreStatus(status, opt.IgnoreStatus) {
			res.MarkAsDropped()
		}
	}

	return result
}

func ResultUtilsFeedIO(task *Task, result []*pipeline.Result) error {
	if len(result) == 0 {
		return nil
	}
	category := datakit.Logging
	version := ""

	if task.Opt != nil {
		if task.Opt.Category != "" {
			category = task.Opt.Category
		}
		if task.Opt.Version != "" {
			version = task.Opt.Version
		}
	}

	pts := []*io.Point{}
	for _, v := range result {
		if v.IsDropped() {
			continue
		}
		if pt, err := v.MakePoint(task.Source, task.MaxMessageLen, category); err != nil {
			l.Debug(err)
		} else {
			pts = append(pts, pt)
		}
	}
	return io.Feed(task.TaskName, category, pts,
		&io.Option{
			HighFreq: disableHighFreqIODdata,
			Version:  version,
		})
}
