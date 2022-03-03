package worker

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
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
	Callback(*Task, []*Result) error
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

func (data *TaskDataTemplate) Callback(task *Task, result []*Result) error {
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

type Result struct {
	Measurement string

	Output *parser.Output

	TS time.Time

	Err string
}

func (r *Result) String() string {
	return fmt.Sprintf("%+#v", r.Output)
}

func NewResult() *Result {
	return &Result{
		Output: &parser.Output{
			Tags: make(map[string]string),
			Data: make(map[string]interface{}),
			Cost: make(map[string]string),
		},
	}
}

func (r *Result) GetTag(k string) (string, error) {
	if v, ok := r.Output.Tags[k]; ok {
		return v, nil
	} else {
		return "", fmt.Errorf("tag not found")
	}
}

func (r *Result) GetField(k string) (interface{}, error) {
	if v, ok := r.Output.Data[k]; ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("field not found")
	}
}

func (r *Result) GetMeasurement() string {
	if v, err := r.GetTag(PipelineMSource); err == nil {
		r.Measurement = v
		r.DeleteTag(PipelineMSource)
	} else if v, err := r.GetField(PipelineMSource); err == nil {
		if v, ok := v.(string); ok {
			r.Measurement = v
			r.DeleteField(PipelineMSource)
		}
	}
	return r.Measurement
}

func (r *Result) SetMeasurement(m string) {
	r.DeleteField(PipelineMSource)
	r.DeleteTag(PipelineMSource)
	r.Measurement = m
}

func (r *Result) GetTime() (time.Time, error) {
	var ts time.Time
	if v, err := r.GetField(PipelineTimeField); err == nil {
		if nanots, ok := v.(int64); ok {
			r.TS = time.Unix(nanots/int64(time.Second),
				nanots%int64(time.Second))
			r.DeleteField(PipelineTimeField)
		}
	}

	if !r.TS.IsZero() {
		return r.TS, nil
	}
	return ts, fmt.Errorf("no time")
}

func (r *Result) GetTags() map[string]string {
	return r.Output.Tags
}

func (r *Result) GetFields() map[string]interface{} {
	return r.Output.Data
}

func (r *Result) GetLastErr() string {
	return r.Err
}

func (r *Result) SetTime(t time.Time) {
	r.DeleteField(PipelineTimeField)
	r.TS = t
}

func (r *Result) SetTag(k, v string) {
	r.Output.Tags[k] = v
}

func (r *Result) SetField(k string, v interface{}) {
	r.Output.Data[k] = v
}

func (r *Result) DeleteField(k string) {
	delete(r.Output.Data, k)
}

func (r *Result) DeleteTag(k string) {
	delete(r.Output.Tags, k)
}

func (r *Result) IsDropped() bool {
	return r.Output.Dropped
}

func (r *Result) MarkAsDropped() {
	r.Output.Dropped = true
}

func (r *Result) MakePointIgnoreDropped(measurement string, maxMessageLen int, category string) (*io.Point, error) {
	if r.IsDropped() {
		return nil, nil
	}
	return r.MakePoint(measurement, maxMessageLen, category)
}

func (r *Result) MakePoint(measurement string, maxMessageLen int, category string) (*io.Point, error) {
	if category == "" {
		category = datakit.Logging
	}
	if r.Measurement == "" {
		r.Measurement = measurement
	}
	return io.NewPoint(r.Measurement, r.Output.Tags, r.Output.Data,
		&io.PointOption{
			Time:              r.TS,
			Category:          category,
			DisableGlobalTags: false,
			Strict:            true,
			MaxFieldValueLen:  maxMessageLen,
		})
}

func (r *Result) CheckFieldValLen(messageLen int) {
	spiltLen := messageLen
	if spiltLen <= 0 { // 当初始化 task 时没有注入最大长度则使用默认值
		spiltLen = maxFieldsLength
	}
	for key := range r.Output.Data {
		if i, err := r.GetField(key); err == nil {
			if mass, isString := i.(string); isString {
				if len(mass) > spiltLen {
					mass = mass[:spiltLen]
					r.SetField(key, mass)
				}
			}
		}
	}
}

func (r *Result) preprocessing(source string, maxMessageLen int) {
	m := r.GetMeasurement()
	if m == "" {
		m = source
	}
	r.SetMeasurement(m)

	rTS, _ := r.GetTime()
	r.SetTime(rTS)

	r.CheckFieldValLen(maxMessageLen)
}

func ParsePlScript(plScript string) (*parser.Engine, error) {
	if ng, err := parser.NewEngine(plScript, funcs.FuncsMap, funcs.FuncsCheckMap, false); err != nil {
		return nil, err
	} else {
		return ng, nil
	}
}

func RunPlStr(cntStr, source string, maxMessageLen int, ng *parser.Engine) (*Result, error) {
	result := &Result{
		Output: nil,
	}
	if ng != nil {
		if err := ng.Run(cntStr); err != nil {
			l.Debug(err)
			result.Err = err.Error()
		}
		result.Output = ng.Result()
	} else {
		result.Output = &parser.Output{
			Cost: map[string]string{},
			Tags: map[string]string{},
			Data: map[string]interface{}{
				PipelineMessageField: cntStr,
			},
		}
	}
	result.preprocessing(source, maxMessageLen)
	return result, nil
}

func RunPlByte(cntByte []byte, encode string, source string, maxMessageLen int, ng *parser.Engine) (*Result, error) {
	cntStr, err := DecodeContent(cntByte, encode)
	if err != nil {
		return nil, err
	}
	return RunPlStr(cntStr, source, maxMessageLen, ng)
}

func RunAsPlTask(category string, source string, service, dataType string,
	contentStr []string, contentByte [][]byte, encode string, ng *parser.Engine) []*Result {
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

func RunPlTask(task *Task, ng *parser.Engine) ([]*Result, error) {
	taskResult := []*Result{}

	encode := TaskDataContentEncode(task.Data)
	cntType := TaskDataContentType(task.Data)
	switch cntType {
	case ContentByte:
		cntByte := TaskDataGetContentByte(task.Data)
		for _, cnt := range cntByte {
			if r, err := RunPlByte(cnt, encode, task.Source, task.MaxMessageLen, ng); err != nil {
				l.Debug(err)
				continue
			} else {
				taskResult = append(taskResult, r)
			}
		}
	case ContentString, "":
		cntStr := TaskDataGetContentStr(task.Data)
		for _, cnt := range cntStr {
			if r, err := RunPlStr(cnt, task.Source, task.MaxMessageLen, ng); err != nil {
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

func DecodeContent(content []byte, encode string) (string, error) {
	var err error
	if encode != "" {
		encode = strings.ToLower(encode)
	}
	switch encode {
	case "gbk", "gb18030":
		content, err = GbToUtf8(content, encode)
		if err != nil {
			return "", err
		}
	case "utf8", "utf-8":
	default:
	}
	return string(content), nil
}

// GbToUtf8 Gb to UTF-8.
// http/api_pipeline.go.
func GbToUtf8(s []byte, encoding string) ([]byte, error) {
	var t transform.Transformer
	switch encoding {
	case "gbk":
		t = simplifiedchinese.GBK.NewDecoder()
	case "gb18030":
		t = simplifiedchinese.GB18030.NewDecoder()
	}
	reader := transform.NewReader(bytes.NewReader(s), t)
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func ResultUtilsAutoFillTime(result []*Result, lastTime time.Time) []*Result {
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

func ResultUtilsLoggingProcessor(task *Task, result []*Result, tags map[string]string, fields map[string]interface{}) []*Result {
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

func ResultUtilsFeedIO(task *Task, result []*Result) error {
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
