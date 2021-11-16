package redis

import (
	"bufio"
	"context"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type dbMeasurement struct {
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
		Name: "redis_db",
		Tags: map[string]interface{}{
			"db": &inputs.TagInfo{
				Desc: "db name",
			},
		},
		Fields: map[string]interface{}{
			"keys": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "key",
			},
			"expires": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "过期时间",
			},
			"avg_ttl": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "avg ttl",
			},
		},
	}
}

func (i *Input) collectDBMeasurement() ([]inputs.Measurement, error) {
	// 获取数据
	ctx := context.Background()
	list, err := i.client.Info(ctx, "Keyspace").Result()
	if err != nil {
		return nil, err
	}
	// 拿到处理后的数据
	info, err := i.ParseInfoData(list)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// ParseInfoData 解析数据并返回指定的数据.
func (i *Input) ParseInfoData(list string) ([]inputs.Measurement, error) {
	rdr := strings.NewReader(list)
	var collectCache []inputs.Measurement
	scanner := bufio.NewScanner(rdr)
	dbIndexSlice := i.DBS
	// 配置定义了db，加入dbIndexSlice
	if i.DB != -1 {
		dbIndexSlice = append(dbIndexSlice, i.DB)
	}

	// 遍历每一行数据
	for scanner.Scan() {
		m := &dbMeasurement{
			name:    "redis_client",
			tags:    make(map[string]string),
			fields:  make(map[string]interface{}),
			resData: make(map[string]interface{}),
		}
		line := scanner.Text()

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// parts数据格式 [db0 keys=8,expires=0,avg_ttl=0]
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		// cmdstat_get:calls=2,usec=16,usec_per_call=8.00
		db := parts[0]
		m.name = "redis_db"
		m.tags["db_name"] = db
		itemStrs := strings.Split(parts[1], ",")

		for _, itemStr := range itemStrs {
			item := strings.Split(itemStr, "=")
			key := item[0]
			val := strings.TrimSpace(item[1])
			m.fields[key] = val
		}

		if len(i.DBS) == 0 {
			if err := m.submit(); err != nil {
				return nil, err
			}
			collectCache = append(collectCache, m)
		} else {
			dbIndex, err := strconv.Atoi(db[2:])
			// 解析db出错，把没出错的信息和错误返回
			if err != nil {
				return collectCache, err
			}

			if IsSlicesHave(dbIndexSlice, dbIndex) {
				if err := m.submit(); err != nil {
					return nil, err
				}

				collectCache = append(collectCache, m)
			}
		}
	}
	return collectCache, nil
}

// 提交数据.
func (m *dbMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.(*inputs.FieldInfo).DataType)
			if err != nil {
				l.Errorf("infoMeasurement metric %v value %v parse error %v", key, value, err)
				return err
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
