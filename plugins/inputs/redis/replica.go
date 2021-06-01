package redis

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type replicaMeasurement struct {
	client  *redis.Client
	name    string
	tags    map[string]string
	fields  map[string]interface{}
	ts      time.Time
	resData map[string]interface{}
}

func (m *replicaMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *replicaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_replica",
		Fields: map[string]interface{}{
			"repl_delay": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "replica delay",
			},
			"master_link_down_since_seconds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of seconds since the link is down",
			},
		},
	}
}

func (i *Input) collectReplicaMeasurement() ([]inputs.Measurement, error) {
	m := &replicaMeasurement{
		client:  i.client,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.name = "redis_replica"

	if err := m.getData(); err != nil {
		return nil, err
	}

	m.submit()

	return []inputs.Measurement{m}, nil
}

// 数据源获取数据
func (m *replicaMeasurement) getData() error {
	ctx := context.Background()
	list, err := m.client.Info(ctx, "commandstats").Result()
	if err != nil {
		l.Error("redis exec `commandstats`, happen error,", err)
		return err
	}

	m.parseInfoData(list)

	return nil
}

// 解析返回
func (m *replicaMeasurement) parseInfoData(list string) error {
	var masterDownSeconds, masterOffset, slaveOffset float64
	var masterStatus, slaveID, ip, port string
	var err error

	rdr := strings.NewReader(list)
	scanner := bufio.NewScanner(rdr)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		record := strings.SplitN(line, ":", 2)
		if len(record) < 2 {
			continue
		}

		//cmdstat_get:calls=2,usec=16,usec_per_call=8.00
		key, value := record[0], record[1]

		if key == "master_repl_offset" {
			masterOffset, _ = strconv.ParseFloat(value, 64)
		}

		if key == "master_link_down_since_seconds" {
			masterDownSeconds, _ = strconv.ParseFloat(value, 64)
		}

		if key == "master_link_status" {
			masterStatus = value
		}

		if re, _ := regexp.MatchString(`^slave\d+`, key); re {
			slaveID = strings.TrimPrefix(key, "slave")
			kv := strings.SplitN(value, ",", 5)
			if len(kv) != 5 {
				continue
			}

			split := strings.Split(kv[0], "=")
			if len(split) != 2 {
				l.Warnf("Failed to parse slave ip. %s", err)
				continue
			}
			ip = split[1]

			split = strings.Split(kv[1], "=")
			if err != nil {
				l.Warnf("Failed to parse slave port. %s", err)
				continue
			}
			port = split[1]

			split = strings.Split(kv[3], "=")
			if err != nil {
				l.Warnf("Failed to parse slave offset. %s", err)
				continue
			}
			slaveOffset, _ = strconv.ParseFloat(split[1], 64)
		}

		delay := masterOffset - slaveOffset
		addr := fmt.Sprintf("%s:%s", ip, port)
		if addr != ":" {
			m.tags["slave_addr"] = fmt.Sprintf("%s:%s", ip, port)
		}

		m.tags["slave_id"] = slaveID

		if delay >= 0 {
			m.resData["repl_delay"] = delay
		}

		if masterStatus != "" {
			m.resData["master_link_down_since_seconds"] = masterDownSeconds
		}
	}

	return nil
}

// 提交数据
func (m *replicaMeasurement) submit() error {
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
