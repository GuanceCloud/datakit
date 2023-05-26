// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/script"
)

const (
	// pipeline关键字段.
	FieldTime       = script.FieldTime
	FieldMessage    = script.FieldMessage
	FieldStatus     = script.FieldStatus
	PlLoggingSource = script.PlLoggingSource

	DefaultStatus = script.DefaultStatus
)

//nolint:structcheck,unused
type Output struct {
	Error error

	Drop bool

	Measurement string
	Time        time.Time

	Tags   map[string]string
	Fields map[string]interface{}
}

type Result struct {
	Output *Output
}

func (r *Result) String() string {
	return fmt.Sprintf("%+#v", r.Output)
}

func NewResult() *Result {
	return &Result{
		Output: &Output{
			Tags:   make(map[string]string),
			Fields: make(map[string]interface{}),
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
	return r.Output.Measurement
}

func (r *Result) SetMeasurement(m string) {
	r.Output.Measurement = m
}

func (r *Result) GetTime() (time.Time, error) {
	// var ts time.Time
	// if v, err := r.GetField(PipelineTimeField); err == nil {
	// 	if nanots, ok := v.(int64); ok {
	// 		r.Output.Time = time.Unix(nanots/int64(time.Second),
	// 			nanots%int64(time.Second))
	// 		r.DeleteField(PipelineTimeField)
	// 	}
	// }
	return r.Output.Time, nil
}

func (r *Result) GetTags() map[string]string {
	return r.Output.Tags
}

func (r *Result) GetFields() map[string]interface{} {
	return r.Output.Fields
}

func (r *Result) GetLastErr() error {
	return r.Output.Error
}

func (r *Result) SetTime(t time.Time) {
	r.Output.Time = t
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
	return r.Output.Drop
}

func (r *Result) MarkAsDropped() {
	r.Output.Drop = true
}
