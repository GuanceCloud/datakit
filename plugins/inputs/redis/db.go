package redis

import (
	"bufio"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type dbMeasurement struct {
	client  *redis.Client
	name    string
	tags    map[string]string
	fields  map[string]interface{}
	ts      time.Time
	resData map[string]interface{}
}

func (m *dbMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *dbMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_client",
		Fields: map[string]interface{}{
			"keys": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
			"expires": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
			"avg_ttl": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
		},
	}
}

func CollectDBMeasurement(cli *redis.Client) *dbMeasurement {
	m := &dbMeasurement{
		client:  cli,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.getData()
	m.submit()

	return m
}

// 数据源获取数据
func (m *dbMeasurement) getData() error {
	list, err := m.client.Info("Keyspace").Result()
	if err != nil {
		return err
	}
	m.parseInfoData(list)

	return nil
}

// 解析返回结果
func (m *dbMeasurement) parseInfoData(list string) error {
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

		//cmdstat_get:calls=2,usec=16,usec_per_call=8.00
		db := parts[0]

		m.tags["db_name"] = db

		itemStrs := strings.Split(parts[1], ",")
		for _, itemStr := range itemStrs {
			item := strings.Split(itemStr, "=")

			key := item[0]
			val := strings.TrimSpace(item[1])

			m.resData[key] = val
		}
	}

	return nil
}

// 提交数据
func (m *dbMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.(*inputs.FieldInfo).DataType)
			if err != nil {
				l.Errorf("infoMeasurement metric %v value %v parse error %v", key, value, err)
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
