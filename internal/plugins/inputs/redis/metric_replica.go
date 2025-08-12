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

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type replicaMeasurement struct{}

//nolint:lll
func (m *replicaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisReplica,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"master_repl_offset": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "The server's current replication offset."},
			// "repl_delay":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "Replica delay"},
			"master_link_down_since_seconds": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Desc: "Number of seconds since the link is down when the link between master and replica is down, only collected for slave redis."},
			"master_link_status":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "Status of the link (up/down), `1` for up, `0` for down, only collected for slave redis."},
			"slave_offset":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "Slave offset, only collected for master redis."},
			"slave_lag":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "Slave lag, only collected for master redis."},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname."},
			"server":       &inputs.TagInfo{Desc: "Server addr."},
			"service_name": &inputs.TagInfo{Desc: "Service name."},
			"slave_id":     &inputs.TagInfo{Desc: "Slave ID, only collected for master redis."},
			"slave_addr":   &inputs.TagInfo{Desc: "Slave addr, only collected for master redis."},
			"slave_state":  &inputs.TagInfo{Desc: "Slave state, only collected for master redis."},
			"master_addr":  &inputs.TagInfo{Desc: "Master addr, only collected for slave redis."},
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
//
// role:slave
// master_host:127.0.0.1
// master_port:6380
// master_link_status:down
// master_last_io_seconds_ago:-1
// master_sync_in_progress:0
// slave_repl_offset:1
// master_link_down_since_seconds:1724739099
// slave_priority:100
// slave_read_only:1
// connected_slaves:0
// master_replid:45a37268f2359e5367c5767ce93ff303ad3f1918
// master_replid2:0000000000000000000000000000000000000000
// master_repl_offset:0
// second_repl_offset:-1
// repl_backlog_active:0
// repl_backlog_size:1048576
// repl_backlog_first_byte_offset:0
// repl_backlog_histlen:0

func (ipt *Input) parseReplicaData(list string) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	var masterIP, masterPort string

	rdr := strings.NewReader(list)
	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		var kvs point.KVs
		line := scanner.Text()

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		record := strings.Split(line, ":")
		if len(record) != 2 {
			continue
		}

		key, value := record[0], record[1]

		if key == "master_host" {
			masterIP = value
		}
		if key == "master_port" {
			masterPort = value
		}

		// key in the replicaMeasurement
		if _, has := replicaFieldMap[key]; has {
			// slave redis, collect master data
			if key == "master_link_status" {
				switch value {
				case "up":
					kvs = kvs.Add(key, 1)
				case "down":
					kvs = kvs.Add(key, 0)
				default:
					l.Warnf("parseReplicaData: unexpected value for master_link_status, got %s", value)
					continue
				}
			} else {
				float, err := strconv.ParseFloat(value, 64)
				if err != nil {
					l.Warnf("parseMasterData: %s, expect to be int, got %s", err, value)
					continue
				}
				kvs = kvs.Add(key, float)
			}
		}
		// special key for slave data
		if slaveMatch.MatchString(key) {
			// master redis, collect slave data
			slaveID, ip, port, state, offset, lag := parseConnectedSlaveString(key, value)
			kvs = kvs.AddTag("slave_id", slaveID)
			kvs = kvs.AddTag("slave_addr", fmt.Sprintf("%s:%s", ip, port))
			kvs = kvs.AddTag("slave_state", state)
			kvs = kvs.Add("slave_offset", offset)
			kvs = kvs.Add("slave_lag", lag)
		}

		if masterIP != "" && masterPort != "" {
			kvs = kvs.AddTag("master_addr", fmt.Sprintf("%s:%s", masterIP, masterPort))
		}

		if kvs.FieldCount() > 0 {
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}
			collectCache = append(collectCache, point.NewPoint(redisReplica, kvs, opts...))
		}
	}
	return collectCache, nil
}

var replicaFieldMap = map[string]struct{}{}

func getReplicaFieldMap() {
	m := replicaMeasurement{}
	for k := range m.Info().Fields {
		replicaFieldMap[k] = struct{}{}
	}
}

/*
slave0:ip=10.254.11.1,port=6379,state=online,offset=1751844676,lag=0
slave1:ip=10.254.11.2,port=6379,state=online,offset=1751844222,lag=0
*/
func parseConnectedSlaveString(slaveName string, keyValues string) (id string, ip string, port string, state string, offset float64, lag float64) {
	slaveID := strings.TrimPrefix(slaveName, "slave")
	kv := strings.SplitN(keyValues, ",", 5)
	if len(kv) != 5 {
		return
	}
	for _, v := range kv {
		k := strings.Split(v, "=")
		if len(k) != 2 {
			continue
		}
		switch k[0] {
		case "ip":
			ip = k[1]
		case "port":
			port = k[1]
		case "state":
			state = k[1]
		case "offset":
			offset, _ = strconv.ParseFloat(k[1], 64)
		case "lag":
			lag, _ = strconv.ParseFloat(k[1], 64)
		}
	}
	return slaveID, ip, port, state, offset, lag
}
