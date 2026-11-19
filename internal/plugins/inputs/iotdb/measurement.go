// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package iotdb

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct{}

func (*Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Cat:    point.Metric,
		Desc:   "IoTDB server metrics collected from its Prometheus reporter endpoint.",
		DescZh: "从 IoTDB Prometheus Reporter 端点采集的服务端指标。",
		Fields: map[string]interface{}{
			"value": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Float,
				Unit:     inputs.NoUnit,
				Desc:     "Prometheus metric sample value.",
			},
		},
		Tags: map[string]interface{}{
			"metric_name": &inputs.TagInfo{Desc: "Original Prometheus metric name, including histogram and summary suffixes."},
			"metric_type": &inputs.TagInfo{Desc: "Prometheus metric family type."},
			"host":        &inputs.TagInfo{Desc: "Host name."},
		},
	}
}
