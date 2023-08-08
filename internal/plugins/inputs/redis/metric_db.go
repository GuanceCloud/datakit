// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"context"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type dbMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	resData  map[string]interface{}
	election bool
}

func (m *dbMeasurement) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElectionV2(m.election))
}

func (m *dbMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisDB,
		Type: "metric",
		Tags: map[string]interface{}{
			"db": &inputs.TagInfo{
				Desc: "db name",
			},
			"host": &inputs.TagInfo{
				Desc: "Hostname",
			},
			"server":       &inputs.TagInfo{Desc: "Server addr"},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
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

func (i *Input) collectDBMeasurement() ([]*point.Point, error) {
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
func (i *Input) ParseInfoData(list string) ([]*point.Point, error) {
	rdr := strings.NewReader(list)
	var collectCache []*point.Point
	scanner := bufio.NewScanner(rdr)
	dbIndexSlice := i.DBS
	// 配置定义了db，加入dbIndexSlice
	if i.DB != -1 {
		dbIndexSlice = append(dbIndexSlice, i.DB)
	}

	// 遍历每一行数据
	for scanner.Scan() {
		m := &dbMeasurement{
			name:     redisClient,
			tags:     make(map[string]string),
			fields:   make(map[string]interface{}),
			resData:  make(map[string]interface{}),
			election: i.Election,
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
		m.name = redisDB
		setHostTagIfNotLoopback(m.tags, i.Host)
		m.tags["db_name"] = db
		itemStrs := strings.Split(parts[1], ",")

		for _, itemStr := range itemStrs {
			item := strings.Split(itemStr, "=")
			key := item[0]
			val := strings.TrimSpace(item[1])
			m.resData[key] = val
		}

		if len(i.DBS) == 0 {
			if err := m.submit(); err != nil {
				return nil, err
			}
			var opts []point.Option

			if m.election {
				m.tags = inputs.MergeTagsWrapper(m.tags, i.Tagger.ElectionTags(), i.Tags, i.Host)
			} else {
				m.tags = inputs.MergeTagsWrapper(m.tags, i.Tagger.HostTags(), i.Tags, i.Host)
			}

			pt := point.NewPointV2([]byte(m.name),
				append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
				opts...)
			collectCache = append(collectCache, pt)
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
				var opts []point.Option

				if m.election {
					m.tags = inputs.MergeTagsWrapper(m.tags, i.Tagger.ElectionTags(), i.Tags, i.Host)
				} else {
					m.tags = inputs.MergeTagsWrapper(m.tags, i.Tagger.HostTags(), i.Tags, i.Host)
				}

				pt := point.NewPointV2([]byte(m.name),
					append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
					opts...)
				collectCache = append(collectCache, pt)
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
				l.Errorf("dbMeasurement metric %v value %v parse error %v", key, value, err)
				return err
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
