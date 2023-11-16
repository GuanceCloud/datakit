// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

//nolint:unused
import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type replicaMeasurement struct{}

//nolint:lll
func (m *replicaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisReplica,
		Type: "metric",
		Fields: map[string]interface{}{
			"repl_delay":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "Replica delay"},
			"master_link_down_since_seconds": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "Number of seconds since the link is down"},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname"},
			"server":       &inputs.TagInfo{Desc: "Server addr"},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
			"slave_id":     &inputs.TagInfo{Desc: "Slave ID"},
		},
	}
}

func (ipt *Input) collectReplicaMeasurement() ([]*point.Point, error) {
	ctx := context.Background()
	list, err := ipt.client.Info(ctx, "replication").Result()
	if err != nil {
		l.Error("redis exec `info replication`, happen error,", err)
		return nil, err
	}

	return ipt.parseReplicaData(list)
}

var slaveMatch = regexp.MustCompile(`^slave\d+`)

// example data:
//
// # Replication
// role:master
// connected_slaves:2
// slave0:ip=127.0.0.1,port=6232,state=online,offset=966,lag=0
// slave1:ip=127.0.0.1,port=6233,state=online,offset=966,lag=0
// master_failover_state:no-failover
// master_replid:b4f438ca13f1ee505f8bf4f03ccab1f648126463
// master_replid2:0000000000000000000000000000000000000000
// master_repl_offset:966
// second_repl_offset:-1
// repl_backlog_active:1
// repl_backlog_size:1048576
// repl_backlog_first_byte_offset:1
// repl_backlog_histlen:966.
func (ipt *Input) parseReplicaData(list string) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Now()))

	masterDownSeconds := map[string]float64{}
	masterData := map[string]float64{}

	// master
	rdr := strings.NewReader(list)
	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		record := strings.Split(line, ":")
		if len(record) != 2 {
			continue
		}

		key, value := record[0], record[1]

		if key == "master_repl_offset" {
			masterData["master_repl_offset"], _ = strconv.ParseFloat(value, 64)
		}

		if key == "master_link_down_since_seconds" {
			masterDownSeconds["master_link_down_since_seconds"], _ = strconv.ParseFloat(value, 64)
		}
	}

	// slaves
	rdr = strings.NewReader(list)
	scanner = bufio.NewScanner(rdr)
	for scanner.Scan() {
		var kvs point.KVs
		slaveData := map[string]float64{}
		var slaveID, ip, port string

		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		record := strings.Split(line, ":")
		if len(record) != 2 {
			continue
		}

		key, value := record[0], record[1]

		if slaveMatch.MatchString(key) {
			slaveID = strings.TrimPrefix(key, "slave")
			kv := strings.SplitN(value, ",", 5)
			if len(kv) != 5 {
				continue
			}

			split := strings.Split(kv[0], "=")
			if len(split) != 2 {
				l.Warnf("Failed to parse slave ip, got %s", kv[0])
				continue
			}
			ip = split[1]

			split = strings.Split(kv[1], "=")
			if len(split) != 2 {
				l.Warnf("Failed to parse slave port, got %s", kv[1])
				continue
			}
			port = split[1]

			split = strings.Split(kv[3], "=")
			if len(split) != 2 {
				l.Warnf("Failed to parse slave offset, got %s", kv[3])
				continue
			}

			temp, err := strconv.ParseFloat(split[1], 64)
			if err != nil {
				l.Warnf("ParseFloat: %s, slaveOffset expect to be int, got %s", err, split[1])
				continue
			}
			slaveData["slave_offset"] = temp
		}

		var masterOffset, slaveOffset float64
		var ok bool
		if masterOffset, ok = masterData["master_repl_offset"]; !ok {
			continue
		}
		if slaveOffset, ok = slaveData["slave_offset"]; !ok {
			continue
		}

		delay := masterOffset - slaveOffset
		addr := fmt.Sprintf("%s:%s", ip, port)
		if addr != ":" {
			kvs = kvs.AddTag("slave_addr", fmt.Sprintf("%s:%s", ip, port))
		}

		kvs = kvs.AddTag("slave_id", slaveID)

		if delay >= 0 {
			kvs = kvs.Add("repl_delay", delay, false, false)
		}

		for k, v := range masterDownSeconds {
			kvs = kvs.Add(k, v, false, false)
		}

		if kvs.FieldCount() > 0 {
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}
			collectCache = append(collectCache, point.NewPointV2(redisReplica, kvs, opts...))
		}
	}

	return collectCache, nil
}
