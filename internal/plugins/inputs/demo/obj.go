// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package demo

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type demoObj struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	TS     time.Time
}

//nolint:lll
func (m *demoObj) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "demo-obj",
		Type: "object",
		Desc: "这是一个对象的 demo(**务必加上每个指标集的描述**)",
		Tags: map[string]interface{}{
			"tag_a": &inputs.TagInfo{Desc: "示例 tag A"},
			"tag_b": &inputs.TagInfo{Desc: "示例 tag B"},
		},
		Fields: map[string]interface{}{
			"usage": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "this is CPU usage",
			},
			"disk_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "this is disk size",
			},
			"mem_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "this is memory size",
			},
			"some_string": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "some string field",
			},
			"ok": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "some boolean field",
			},
		},
	}
}
