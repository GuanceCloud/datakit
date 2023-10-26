// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"context"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type clusterMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	resData  map[string]interface{}
	election bool
}

//nolint:lll
func (m *clusterMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisCluster,
		Type: "metric",
		Fields: map[string]interface{}{
			"cluster_state": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "State is ok if the node is able to receive queries. fail if there is at least one hash slot which is unbound (no node associated), in error state (node serving it is flagged with FAIL flag), or if the majority of masters can't be reached by this node.",
			},
			"cluster_slots_assigned": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     " Number of slots which are associated to some node (not unbound). This number should be 16384 for the node to work properly, which means that each hash slot should be mapped to a node.",
			},
			"cluster_slots_ok": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of hash slots mapping to a node not in `FAIL` or `PFAIL` state.",
			},
			"cluster_slots_pfail": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of hash slots mapping to a node in `PFAIL` state. Note that those hash slots still work correctly, as long as the `PFAIL` state is not promoted to FAIL by the failure detection algorithm. `PFAIL` only means that we are currently not able to talk with the node, but may be just a transient error.",
			},
			"cluster_slots_fail": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of hash slots mapping to a node in FAIL state. If this number is not zero the node is not able to serve queries unless cluster-require-full-coverage is set to no in the configuration.",
			},
			"cluster_known_nodes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total number of known nodes in the cluster, including nodes in HANDSHAKE state that may not currently be proper members of the cluster.",
			},
			"cluster_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of master nodes serving at least one hash slot in the cluster.",
			},
			"cluster_current_epoch": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "The local Current Epoch variable. This is used in order to create unique increasing version numbers during fail overs.",
			},
			"cluster_my_epoch": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "The Config Epoch of the node we are talking with. This is the current configuration version assigned to this node.",
			},
			"cluster_stats_messages_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of messages sent via the cluster node-to-node binary bus.",
			},
			"cluster_stats_messages_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of messages received via the cluster node-to-node binary bus.",
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{
				Desc: "Hostname",
			},
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
		},
	}
}

func (ipt *Input) CollectClusterMeasurement() ([]*point.Point, error) {
	ctx := context.Background()
	list, err := ipt.client.ClusterInfo(ctx).Result()
	if err != nil {
		l.Errorf("get clusterinfo error %v", err)
		return nil, err
	}
	info, err := ipt.ParseClusterData(list)
	if err != nil {
		l.Errorf("paserclusterdata error %v", err)
		return nil, err
	}
	return info, nil
}

// ParseClusterData 解析数据并返回指定的数据.
func (ipt *Input) ParseClusterData(list string) ([]*point.Point, error) {
	rdr := strings.NewReader(list)
	var collectCache []*point.Point
	scanner := bufio.NewScanner(rdr)

	// 遍历每一行数据
	for scanner.Scan() {
		m := &clusterMeasurement{
			name:     "cluster",
			tags:     make(map[string]string),
			fields:   make(map[string]interface{}),
			resData:  make(map[string]interface{}),
			election: ipt.Election,
		}
		line := scanner.Text()
		// parts:[cluster_known_nodes 1]
		parts := strings.SplitN(line, ":", 2)

		if len(parts) < 2 {
			continue
		}

		m.name = redisCluster
		m.tags["server_addr"] = ipt.Addr
		m.fields[parts[0]] = parts[1]
		err := m.submit()
		if err != nil {
			return nil, err
		}
		var opts []point.Option

		if m.election {
			m.tags = inputs.MergeTagsWrapper(m.tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.Host)
		} else {
			m.tags = inputs.MergeTagsWrapper(m.tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.Host)
		}

		pt := point.NewPointV2(m.name,
			append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
			opts...)
		collectCache = append(collectCache, pt)
	}
	return collectCache, nil
}

// 提交数据.
func (m *clusterMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.(*inputs.FieldInfo).DataType)
			if err != nil {
				l.Errorf("clusterMeasurement metric %v value %v parse error %v", key, value, err)
				return err
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
