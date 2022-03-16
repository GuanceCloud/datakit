package pipeline

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

const (
	// pipeline关键字段.
	PipelineTimeField     = "time"
	PipelineMessageField  = "message"
	PipelineStatusField   = "status"
	PipelineMSource       = "source"
	DefaultPipelineStatus = "info"
	// ES value can be at most 32766 bytes long.
	maxFieldsLength = 32766
)

type Result struct {
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
			Tags:   make(map[string]string),
			Fields: make(map[string]interface{}),
			Cost:   make(map[string]string),
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
	if v, ok := r.Output.Fields[k]; ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("field not found")
	}
}

func (r *Result) GetMeasurement() string {
	return r.Output.DataMeasurement
}

func (r *Result) SetMeasurement(m string) {
	r.Output.DataMeasurement = m
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
	return r.Output.Fields
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
	r.Output.Fields[k] = v
}

func (r *Result) DeleteField(k string) {
	delete(r.Output.Fields, k)
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
	if r.Output.DataMeasurement == "" {
		if measurement == "" {
			measurement = "default"
		}
		r.Output.DataMeasurement = measurement
	}
	return io.NewPoint(r.Output.DataMeasurement, r.Output.Tags, r.Output.Fields,
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
	for key := range r.Output.Fields {
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
