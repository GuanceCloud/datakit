// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package self

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.InputV2 = &Input{}

type datakitMeasurement struct {
	inputs.CommonMeasurement
}

func (m *datakitMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.CommonMeasurement.Name,
		m.CommonMeasurement.Tags,
		m.CommonMeasurement.Fields, point.MOpt())
}

func (m *datakitMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementName,
		Type: "metric",
		Tags: map[string]interface{}{
			"arch":              &inputs.TagInfo{Desc: "Architecture of the DataKit"},
			"os":                &inputs.TagInfo{Desc: "Operation System of the DataKit, such as linux/mac/windows"},
			"os_version_detail": &inputs.TagInfo{Desc: "Operation System release of the DataKit, such as Ubuntu 20.04.2 LTS, macOS 10.15 Catalina"},
			"host":              &inputs.TagInfo{Desc: "Hostname of the DataKit"},
			"uuid":              &inputs.TagInfo{Desc: "**Deprecated**, currently use `hostname` as DataKit's UUID"},
			"namespace":         &inputs.TagInfo{Desc: "Election namespace(datakit.conf/namespace) of DataKit, may be not set"},
			"version":           &inputs.TagInfo{Desc: "DataKit version"},
		},

		Fields: map[string]interface{}{
			"cpu_usage": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Float,
				Unit:     inputs.Percent,
				Desc:     "CPU usage of the datakit",
			},

			"cpu_usage_top": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Float,
				Unit:     inputs.Percent,
				Desc:     "CPU usage(command `top`) of the datakit",
			},

			"incumbency": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationSecond,
				Desc:     "**Deprecated**. same as `elected`",
			},

			"elected": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationSecond,
				Desc:     "Elected duration, if not elected, the value is 0",
			},

			"dropped_points": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Current dropped points due to cache clean",
			},
			"dropped_point_total": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total dropped points due to cache clean",
			},

			"heap_alloc": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Bytes of allocated heap objects",
			},
			"max_heap_alloc": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Max bytes of allocated heap objects since DataKit start",
			},
			"min_heap_alloc": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Minimal bytes of allocated heap objects since DataKit start",
			},

			"heap_objects": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of allocated heap objects",
			},
			"max_heap_objects": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Max number of allocated heap objects since DataKit start",
			},
			"min_heap_objects": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Minimal number of allocated heap objects since DataKit start",
			},

			"heap_sys": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Bytes of heap memory obtained from OS(Estimates the largest size of the heap has had)",
			},
			"max_heap_sys": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Max bytes of heap memory obtained from OS since DataKit start",
			},
			"min_heap_sys": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Minimal bytes of heap memory obtained from OS since DataKit start",
			},

			"num_goroutines": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of goroutines that currently exitst",
			},
			"max_num_goroutines": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Max number of goroutines since DataKit start",
			},
			"min_num_goroutines": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Minimal number of goroutines since DataKit start",
			},

			"pid": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.UnknownUnit,
				Desc:     "DataKit process ID",
			},

			"uptime": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationSecond,
				Desc:     "Uptime of DataKit",
			},

			"open_files": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "open files of DataKit(Only Linux support, others are -1)",
			},
		},
	}
}
