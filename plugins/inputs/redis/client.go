package redis

import (
	"bufio"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type clientMeasurement struct {
	client  *redis.Client
	name    string
	tags    map[string]string
	fields  map[string]interface{}
	ts      time.Time
	resData map[string]interface{}
}

func (m *clientMeasurement) LineProto() (io.Point, error) {
	return io.MakeMetric(m.name, m.tags, m.fields, m.ts)
}

func (m *clientMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_client",
		Fields: map[string]*inputs.FieldInfo{
			"id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
			"addr": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
			"fd": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
			"age": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
			"idle": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
			"sub": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
			"psub": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
		},
	}
}

func CollectClientMeasurement(cli *redis.Client, tags map[string]string) *clientMeasurement {
	m := &clientMeasurement{
		client:  cli,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}
	m.name = "redis_client_list"
	m.tags = tags

	m.getData()
	m.submit()

	return m
}

// 数据源获取数据
func (m *clientMeasurement) getData() error {
	list, err := m.client.ClientList().Result()
	if err != nil {
		return err
	}
	m.parseInfoData(list)

	return nil
}

// 解析返回结果
func (m *clientMeasurement) parseInfoData(list string) error {
	rdr := strings.NewReader(list)

	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, " ", 18)
		if len(parts) < 18 {
			continue
		}

		for _, part := range parts {
			item := strings.Split(part, "=")

			key := item[0]
			val := strings.TrimSpace(item[1])

			if key == "name" {
				if val == "" {
					val = "unknown"
				} else {
					val = val
				}
				m.tags["name"] = val
			} else {
				m.resData[key] = val
			}
		}
	}

	return nil
}

// 提交数据
func (m *clientMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.DataType)
			if err != nil {
				l.Errorf("infoMeasurement metric %v value %v parse error %v", key, value, err)
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
