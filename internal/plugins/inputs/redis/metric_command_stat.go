// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type commandMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       time.Time
	resData  map[string]interface{}
	election bool
}

func (m *commandMeasurement) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElectionV2(m.election))
}

//nolint:lll
func (m *commandMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_command_stat",
		Type: "metric",
		Fields: map[string]interface{}{
			"calls": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of calls that reached command execution",
			},
			"usec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The total CPU time consumed by these commands",
			},
			"usec_per_call": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The average CPU consumed per command execution",
			},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname"},
			"method":       &inputs.TagInfo{Desc: "Command type"},
			"server":       &inputs.TagInfo{Desc: "Server addr"},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
		},
	}
}

// 解析返回结果.
func (i *Input) parseCommandData(list string) ([]*point.Point, error) {
	var collectCache []*point.Point

	rdr := strings.NewReader(list)
	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		m := &commandMeasurement{
			name:     "redis_command_stat",
			tags:     make(map[string]string),
			fields:   make(map[string]interface{}),
			resData:  make(map[string]interface{}),
			election: i.Election,
		}
		setHostTagIfNotLoopback(m.tags, i.Host)
		for key, value := range i.Tags {
			m.tags[key] = value
		}

		// cmdstat_get:calls=2,usec=16,usec_per_call=8.00
		method := parts[0]

		m.tags["method"] = method

		itemStrs := strings.Split(parts[1], ",")
		for _, itemStr := range itemStrs {
			item := strings.Split(itemStr, "=")

			key := item[0]
			val := strings.TrimSpace(item[1])

			m.resData[key] = val
		}

		if err := m.submit(); err != nil {
			return nil, err
		}

		if len(m.fields) > 0 {
			m.ts = time.Now()
			var opts []point.Option

			if m.election {
				m.tags = inputs.MergeTags(i.Tagger.ElectionTags(), m.tags, i.Host)
			} else {
				m.tags = inputs.MergeTags(i.Tagger.HostTags(), m.tags, i.Host)
			}

			pt := point.NewPointV2([]byte(m.name),
				append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
				opts...)
			collectCache = append(collectCache, pt)
		}
	}

	return collectCache, nil
}

// 提交数据.
func (m *commandMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.(*inputs.FieldInfo).DataType)
			if err != nil {
				l.Errorf("commandMeasurement metric %v value %v parse error %v", key, value, err)
				return err
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
