// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func Test_exportMetaInfo(t *testing.T) {
	type args struct {
		ipts map[string]inputs.Creator
	}
	tests := []struct {
		name    string
		args    args
		want    outputMetaInfo
		wantErr bool
	}{
		{
			name:    "ok",
			args:    args{ipts: mockIpts01()},
			want:    getWant01(),
			wantErr: false,
		},
		{
			name:    "error type='objectERROR'",
			args:    args{ipts: mockIpts02()},
			want:    outputMetaInfo{},
			wantErr: true,
		},
		{
			name:    "error type=' '",
			args:    args{ipts: mockIpts03()},
			want:    outputMetaInfo{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exportMetaInfo(tt.args.ipts)
			if (err != nil) != tt.wantErr {
				t.Errorf("exportMetaInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				l.Debugf("exportMetaInfo() got wanted error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var gotObj outputMetaInfo
			err = json.Unmarshal(got, &gotObj)
			if err != nil {
				t.Errorf("Unmarshal error = %v", err)
			}

			assert.Equal(t, gotObj.MetricMetaInfo, tt.want.MetricMetaInfo)
			assert.Equal(t, gotObj.ObjectMetaInfo, tt.want.ObjectMetaInfo)
		})
	}
}

func mockIpts01() map[string]inputs.Creator {
	mockInputs = map[string]inputs.Creator{}

	// add all inputs ...
	mockInitCPU()
	mockInitHostObject()
	mockInitDemo()

	return mockInputs
}

func mockIpts02() map[string]inputs.Creator {
	mockInputs = map[string]inputs.Creator{}

	// add all inputs ...
	mockInitError01()
	mockInitCPU()
	mockInitHostObject()
	mockInitDemo()

	return mockInputs
}

func mockIpts03() map[string]inputs.Creator {
	mockInputs = map[string]inputs.Creator{}

	// add all inputs ...
	mockInitError02()
	mockInitCPU()
	mockInitHostObject()
	mockInitDemo()

	return mockInputs
}

func getWant01() outputMetaInfo {
	str := `{
  "version": "1.16.1-rc1_1953-fix-point-info",
  "release_date": "2023-11-02 06:03:11",
  "doc": [
    "本文档主要用来描述 DataKit 所采集到的各种指标集数据",
    "-----------------------------",
    "version 为当前 json 对应的 datakit 版本号",
    "release_date 为当前 json 的发布日期",
    "其中 metrics.xxx 为指标集名称",
    "metrics.xxx.desc 为指标集描述",
    "metrics.xxx.type 为指标集类型，目前只有指标(M::)和对象(O::)",
    "metrics.xxx.fields 为单个指标集中的指标列表",
    "metrics.xxx.fields.xxx.data_type 为具体指标的数据类型，目前只要有 int/float/int/string 四种类型",
    "metrics.xxx.fields.xxx.unit 为具体指标的单位，可执行 cat measurements-meta.json  | grep unit | sort  | uniq 查看当前的单位列表（注意，这个列表中的单位可以再调整）",
    "metrics.xxx.fields.xxx.desc 为具体指标描述",
    "metrics.xxx.fields.xxx.yyy 其余字段暂时应该没用",
    "metrics.xxx.tags 为单个指标集中的标签列表",
    "-----------------------------",
    "objects.xxx 下为对象指标集，其余字段类似"
  ],
  "metrics": {
    "demo-metric": {
      "desc": "这是一个指标集的 demo(**务必加上每个指标集的描述**)",
      "type": "metric",
      "fields": {
        "disk_size": {
          "type": "gauge",
          "data_type": "int",
          "unit": "B",
          "desc": "this is disk size",
          "disabled": false
        },
        "mem_size": {
          "type": "gauge",
          "data_type": "int",
          "unit": "B",
          "desc": "this is memory size",
          "disabled": false
        },
        "ok": {
          "type": "gauge",
          "data_type": "bool",
          "unit": "-",
          "desc": "some boolean field",
          "disabled": false
        },
        "some_string": {
          "type": "gauge",
          "data_type": "string",
          "unit": "-",
          "desc": "some string field",
          "disabled": false
        },
        "usage": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "this is CPU usage",
          "disabled": false
        }
      },
      "tags": {
        "tag_a": {
          "Desc": "示例 tag A"
        },
        "tag_b": {
          "Desc": "示例 tag B"
        }
      },
      "from": "mockDemo"
    },
    "demo-metric-0-tag": {
      "desc": "这是一个指标集的 demo(**务必加上每个指标集的描述**)",
      "type": "metric",
      "fields": {
        "usage": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "this is CPU usage",
          "disabled": false
        }
      },
      "tags": {},
      "from": "mockDemo"
    },
    "demo-metric2": {
      "desc": "这是一个指标集的 demo(**务必加上每个指标集的描述**)",
      "type": "metric",
      "fields": {
        "disk_size": {
          "type": "gauge",
          "data_type": "int",
          "unit": "B",
          "desc": "this is disk size",
          "disabled": false
        },
        "mem_size": {
          "type": "gauge",
          "data_type": "int",
          "unit": "B",
          "desc": "this is memory size",
          "disabled": false
        },
        "ok": {
          "type": "gauge",
          "data_type": "bool",
          "unit": "-",
          "desc": "some boolean field",
          "disabled": false
        },
        "some_string": {
          "type": "gauge",
          "data_type": "string",
          "unit": "-",
          "desc": "some string field",
          "disabled": false
        },
        "usage": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "this is CPU usage",
          "disabled": false
        }
      },
      "tags": {
        "tag_a": {
          "Desc": "示例 tag A"
        },
        "tag_b": {
          "Desc": "示例 tag B"
        }
      },
      "from": "mockDemo"
    },
    "mockCPU": {
      "desc": "",
      "type": "metric",
      "fields": {
        "core_temperature": {
          "type": "gauge",
          "data_type": "float",
          "unit": "C",
          "desc": "CPU core temperature. This is collected by default. Only collect the average temperature of all cores.",
          "disabled": false
        },
        "load5s": {
          "type": "gauge",
          "data_type": "int",
          "unit": "-",
          "desc": "CPU average load in 5 seconds.",
          "disabled": false
        },
        "usage_guest": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU spent running a virtual CPU for guest operating systems.",
          "disabled": false
        },
        "usage_guest_nice": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU spent running a nice guest(virtual CPU for guest operating systems).",
          "disabled": false
        },
        "usage_idle": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU in the idle task.",
          "disabled": false
        },
        "usage_iowait": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU waiting for I/O to complete.",
          "disabled": false
        },
        "usage_irq": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU servicing hardware interrupts.",
          "disabled": false
        },
        "usage_nice": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU in user mode with low priority (nice).",
          "disabled": false
        },
        "usage_softirq": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU servicing soft interrupts.",
          "disabled": false
        },
        "usage_steal": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU spent in other operating systems when running in a virtualized environment.",
          "disabled": false
        },
        "usage_system": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU in system mode.",
          "disabled": false
        },
        "usage_total": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU in total active usage, as well as (100 - usage_idle).",
          "disabled": false
        },
        "usage_user": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "% CPU in user mode.",
          "disabled": false
        }
      },
      "tags": {
        "cpu": {
          "Desc": "CPU core ID. For ` + "`cpu-total`" + `, it means *all-CPUs-in-one-tag*. If you want every CPU's metric, please enable ` + "`percpu`" + ` option in *cpu.conf* or set ` + "`ENV_INPUT_CPU_PERCPU`" + ` under K8s"
        },
        "host": {
          "Desc": "System hostname."
        }
      },
      "from": "CPU"
    }
  },
  "objects": {
    "HostObject": {
      "desc": "Host object metrics",
      "type": "object",
      "fields": {
        "cpu_usage": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "CPU usage",
          "disabled": false
        },
        "datakit_ver": {
          "type": "",
          "data_type": "string",
          "unit": "-",
          "desc": "collector version",
          "disabled": false
        },
        "disk_used_percent": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "disk usage",
          "disabled": false
        },
        "diskio_read_bytes_per_sec": {
          "type": "gauge",
          "data_type": "int",
          "unit": "B/S",
          "desc": "disk read rate",
          "disabled": false
        },
        "diskio_write_bytes_per_sec": {
          "type": "gauge",
          "data_type": "int",
          "unit": "B/S",
          "desc": "disk write rate",
          "disabled": false
        },
        "load": {
          "type": "gauge",
          "data_type": "float",
          "unit": "-",
          "desc": "system load",
          "disabled": false
        },
        "logging_level": {
          "type": "",
          "data_type": "string",
          "unit": "-",
          "desc": "log level",
          "disabled": false
        },
        "mem_used_percent": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "memory usage",
          "disabled": false
        },
        "message": {
          "type": "",
          "data_type": "string",
          "unit": "-",
          "desc": "Summary of all host information",
          "disabled": false
        },
        "net_recv_bytes_per_sec": {
          "type": "gauge",
          "data_type": "int",
          "unit": "B/S",
          "desc": "network receive rate",
          "disabled": false
        },
        "net_send_bytes_per_sec": {
          "type": "gauge",
          "data_type": "int",
          "unit": "B/S",
          "desc": "network send rate",
          "disabled": false
        },
        "start_time": {
          "type": "",
          "data_type": "int",
          "unit": "ms",
          "desc": "Host startup time (Unix timestamp)",
          "disabled": false
        }
      },
      "tags": {
        "host": {
          "Desc": "Hostname. Required."
        },
        "name": {
          "Desc": "Hostname"
        },
        "os": {
          "Desc": "Host OS type"
        }
      },
      "from": "mockHostObject"
    },
    "demo-obj": {
      "desc": "这是一个对象的 demo(**务必加上每个指标集的描述**)",
      "type": "object",
      "fields": {
        "disk_size": {
          "type": "gauge",
          "data_type": "int",
          "unit": "B",
          "desc": "this is disk size",
          "disabled": false
        },
        "mem_size": {
          "type": "gauge",
          "data_type": "int",
          "unit": "B",
          "desc": "this is memory size",
          "disabled": false
        },
        "ok": {
          "type": "gauge",
          "data_type": "bool",
          "unit": "-",
          "desc": "some boolean field",
          "disabled": false
        },
        "some_string": {
          "type": "gauge",
          "data_type": "string",
          "unit": "-",
          "desc": "some string field",
          "disabled": false
        },
        "usage": {
          "type": "gauge",
          "data_type": "float",
          "unit": "percent",
          "desc": "this is CPU usage",
          "disabled": false
        }
      },
      "tags": {
        "tag_a": {
          "Desc": "示例 tag A"
        },
        "tag_b": {
          "Desc": "示例 tag B"
        }
      },
      "from": "mockDemo"
    }
  }
}
`
	var wantObj outputMetaInfo
	err := json.Unmarshal([]byte(str), &wantObj)
	if err != nil {
		l.Errorf("Unmarshal error = %v", err)
	}

	return wantObj
}

// common codes

var mockInputs map[string]inputs.Creator

func mockAdd(name string, creator inputs.Creator) {
	if _, ok := mockInputs[name]; ok {
		l.Fatalf("inputs %s exist(from datakit)", name)
	}

	mockInputs[name] = creator
}

// mock cpu input

func mockInitCPU() { //nolint:gochecknoinits
	mockAdd("CPU", func() inputs.Input {
		return &mockInputCPU{}
	})
}

type mockInputCPU struct{}

func (*mockInputCPU) Run()                     {}
func (*mockInputCPU) Terminate()               {}
func (*mockInputCPU) Catalog() string          { return "host" }
func (*mockInputCPU) SampleConfig() string     { return "" }
func (*mockInputCPU) AvailableArchs() []string { return []string{} }
func (*mockInputCPU) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&mockCPUMeasurement{},
	}
}

type mockCPUMeasurement struct{}

//nolint:lll
func (mockCPUMeasurement) Info() *inputs.MeasurementInfo {
	// see https://man7.org/linux/man-pages/man5/proc.5.html
	return &inputs.MeasurementInfo{
		Name: "mockCPU",
		Type: "metric",
		Fields: map[string]interface{}{
			"usage_user": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU in user mode.",
			},

			"usage_nice": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU in user mode with low priority (nice).",
			},

			"usage_system": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU in system mode.",
			},

			"usage_idle": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU in the idle task.",
			},

			"usage_iowait": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU waiting for I/O to complete.",
			},

			"usage_irq": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU servicing hardware interrupts.",
			},

			"usage_softirq": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU servicing soft interrupts.",
			},

			"usage_steal": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU spent in other operating systems when running in a virtualized environment.",
			},

			"usage_guest": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU spent running a virtual CPU for guest operating systems.",
			},

			"usage_guest_nice": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU spent running a nice guest(virtual CPU for guest operating systems).",
			},

			"usage_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU in total active usage, as well as (100 - usage_idle).",
			},
			"core_temperature": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Celsius,
				Desc: "CPU core temperature. This is collected by default. Only collect the average temperature of all cores.",
			},
			"load5s": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.UnknownUnit,
				Desc: "CPU average load in 5 seconds.",
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
			"cpu":  &inputs.TagInfo{Desc: "CPU core ID. For `cpu-total`, it means *all-CPUs-in-one-tag*. If you want every CPU's metric, please enable `percpu` option in *cpu.conf* or set `ENV_INPUT_CPU_PERCPU` under K8s"},
		},
	}
}

// mock host object input

func mockInitHostObject() { //nolint:gochecknoinits
	mockAdd("mockHostObject", func() inputs.Input {
		return &mockInputHostObject{}
	})
}

type mockInputHostObject struct{}

func (*mockInputHostObject) Run()                     {}
func (*mockInputHostObject) Terminate()               {}
func (*mockInputHostObject) Catalog() string          { return "host" }
func (*mockInputHostObject) SampleConfig() string     { return "" }
func (*mockInputHostObject) AvailableArchs() []string { return []string{} }
func (*mockInputHostObject) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&mockHostObjectMeasurement{},
	}
}

type mockHostObjectMeasurement struct{}

//nolint:lll
func (*mockHostObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "HostObject",
		Type: "object",
		Desc: "Host object metrics",
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "Hostname. Required."},
			"name": &inputs.TagInfo{Desc: "Hostname"},
			"os":   &inputs.TagInfo{Desc: "Host OS type"},
		},
		Fields: map[string]interface{}{
			"message":                    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Summary of all host information"},
			"start_time":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationMS, Desc: "Host startup time (Unix timestamp)"},
			"datakit_ver":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "collector version"},
			"cpu_usage":                  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU usage"},
			"mem_used_percent":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "memory usage"},
			"load":                       &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "system load"},
			"disk_used_percent":          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "disk usage"},
			"diskio_read_bytes_per_sec":  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "disk read rate"},
			"diskio_write_bytes_per_sec": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "disk write rate"},
			"net_recv_bytes_per_sec":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "network receive rate"},
			"net_send_bytes_per_sec":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "network send rate"},
			"logging_level":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "log level"},
		},
	}
}

// mock demo input

func mockInitDemo() { //nolint:gochecknoinits
	mockAdd("mockDemo", func() inputs.Input {
		return &mockInputDemo{}
	})
}

type mockInputDemo struct{}

func (*mockInputDemo) Run()                     {}
func (*mockInputDemo) Terminate()               {}
func (*mockInputDemo) Catalog() string          { return "host" }
func (*mockInputDemo) SampleConfig() string     { return "" }
func (*mockInputDemo) AvailableArchs() []string { return []string{} }
func (*mockInputDemo) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&mockDemoMetric{},
		&mockDemoMetric2{},
		&mockDemoObj{},
		&mockDemoLog{},
		&mockDemoMetric0Field{},
		&mockDemoMetric0Tag{},
	}
}

type (
	mockDemoMetric       struct{}
	mockDemoMetric2      struct{}
	mockDemoObj          struct{}
	mockDemoLog          struct{}
	mockDemoMetric0Field struct{}
	mockDemoMetric0Tag   struct{}
)

func (m *mockDemoMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "demo-metric",
		Type: "metric",
		Desc: "这是一个指标集的 demo(**务必加上每个指标集的描述**)",
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

func (m *mockDemoMetric2) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "demo-metric2",
		Type: "metric",
		Desc: "这是一个指标集的 demo(**务必加上每个指标集的描述**)",
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

//nolint:lll
func (m *mockDemoObj) Info() *inputs.MeasurementInfo {
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

//nolint:lll
func (m *mockDemoLog) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "demo-log",
		Type: "logging",
		Desc: "这是一个日志的 demo(**务必加上每个指标集的描述**)",
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

func (m *mockDemoMetric0Field) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "demo-metric-0-field",
		Type: "metric",
		Desc: "这是一个指标集的 demo(**务必加上每个指标集的描述**)",
		Tags: map[string]interface{}{
			"tag_a": &inputs.TagInfo{Desc: "示例 tag A"},
			"tag_b": &inputs.TagInfo{Desc: "示例 tag B"},
		},
		Fields: map[string]interface{}{},
	}
}

func (m *mockDemoMetric0Tag) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "demo-metric-0-tag",
		Type: "metric",
		Desc: "这是一个指标集的 demo(**务必加上每个指标集的描述**)",
		Tags: map[string]interface{}{},
		Fields: map[string]interface{}{
			"usage": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "this is CPU usage",
			},
		},
	}
}

// mock error01 input

func mockInitError01() { //nolint:gochecknoinits
	mockAdd("mockError01", func() inputs.Input {
		return &mockInputError01{}
	})
}

type mockInputError01 struct{}

func (*mockInputError01) Run()                     {}
func (*mockInputError01) Terminate()               {}
func (*mockInputError01) Catalog() string          { return "host" }
func (*mockInputError01) SampleConfig() string     { return "" }
func (*mockInputError01) AvailableArchs() []string { return []string{} }
func (*mockInputError01) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&mockError01Measurement{},
	}
}

type mockError01Measurement struct{}

//nolint:lll
func (*mockError01Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "Error01",
		Type: "objectERROR",
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "Hostname. Required."},
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Summary of all host information"},
		},
	}
}

// mock error02 input

func mockInitError02() { //nolint:gochecknoinits
	mockAdd("mockError02", func() inputs.Input {
		return &mockInputError02{}
	})
}

type mockInputError02 struct{}

func (*mockInputError02) Run()                     {}
func (*mockInputError02) Terminate()               {}
func (*mockInputError02) Catalog() string          { return "host" }
func (*mockInputError02) SampleConfig() string     { return "" }
func (*mockInputError02) AvailableArchs() []string { return []string{} }
func (*mockInputError02) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&mockError02Measurement{},
	}
}

type mockError02Measurement struct{}

//nolint:lll
func (*mockError02Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "Error02",
		Type: " ",
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "Hostname. Required."},
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Summary of all host information"},
		},
	}
}
