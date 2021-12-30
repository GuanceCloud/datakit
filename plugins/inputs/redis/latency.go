package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type latencyMeasurement struct {
	name    string
	tags    map[string]string
	fields  map[string]interface{}
	ts      time.Time
	resData map[string]interface{}
}

func (m *latencyMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *latencyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_latency",
		Fields: map[string]interface{}{
			"event_name": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Event name.",
			},
			"occur_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Unix timestamp of the latest latency spike for the event.",
			},
			"cost_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Latest event latency in millisecond.",
			},
			"max_cost_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "All-time maximum latency for this event.",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
		},
	}
}

func (i *Input) CollectLatencyMeasurement() ([]inputs.Measurement, error) {
	ctx := context.Background()
	list := i.client.Do(ctx, "latency", "latest").String()
	info, err := i.ParseLatencyData(list)
	if err != nil {
		l.Errorf("paserlatencydata error %v", err)
		return nil, err
	}
	return info, nil
}

// ParseLatencyData 解析数据并返回指定的数据.
func (i *Input) ParseLatencyData(list string) ([]inputs.Measurement, error) {
	var collectCache []inputs.Measurement

	// [latency latest:  command 1640151523 324 1000] ]]
	part := strings.Split(list, "[[")

	// redis没有最新延迟事件
	if len(part) != 2 {
		l.Info("have no delayed event")
		return nil, nil
	}
	// "command 1640151523 324 1000"
	part1 := strings.Split(part[1], "]]")

	// command 1640151523 324 1000
	finalPart := strings.Split(part1[0], " ")

	// 长度不足4则失败
	if len(finalPart) != 4 {
		l.Errorf("parse latency data error")
		return nil, fmt.Errorf("parse latency data error")
	}

	fieldName := []string{"event_name", "occur_time", "cost_time", "max_cost_time"}

	for index, info := range fieldName {
		m := &latencyMeasurement{
			name:    "redis_latency",
			tags:    make(map[string]string),
			fields:  make(map[string]interface{}),
			resData: make(map[string]interface{}),
		}
		m.fields[info] = finalPart[index]
		m.tags["server_addr"] = i.Addr

		err := m.submit()
		if err != nil {
			return nil, err
		}
		collectCache = append(collectCache, m)
	}

	return collectCache, nil
}

// 提交数据.
func (m *latencyMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.(*inputs.FieldInfo).DataType)
			if err != nil {
				l.Errorf("latencyMeasurement metric %v value %v parse error %v", key, value, err)
				return err
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
