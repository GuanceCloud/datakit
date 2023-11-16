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
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type clusterMeasurement struct{}

// see also: https://redis.io/commands/cluster-info/
//
//nolint:lll
func (m *clusterMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisCluster,
		Type: "metric",
		Fields: map[string]interface{}{
			"cluster_state":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "State is ok if the node is able to receive queries. fail if there is at least one hash slot which is unbound (no node associated), in error state (node serving it is flagged with FAIL flag), or if the majority of masters can't be reached by this node."},
			"cluster_slots_assigned":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: " Number of slots which are associated to some node (not unbound). This number should be 16384 for the node to work properly, which means that each hash slot should be mapped to a node."},
			"cluster_slots_ok":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of hash slots mapping to a node not in `FAIL` or `PFAIL` state."},
			"cluster_slots_pfail":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of hash slots mapping to a node in `PFAIL` state. Note that those hash slots still work correctly, as long as the `PFAIL` state is not promoted to FAIL by the failure detection algorithm. `PFAIL` only means that we are currently not able to talk with the node, but may be just a transient error."},
			"cluster_slots_fail":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of hash slots mapping to a node in FAIL state. If this number is not zero the node is not able to serve queries unless cluster-require-full-coverage is set to no in the configuration."},
			"cluster_known_nodes":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of known nodes in the cluster, including nodes in HANDSHAKE state that may not currently be proper members of the cluster."},
			"cluster_size":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of master nodes serving at least one hash slot in the cluster."},
			"cluster_current_epoch":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The local Current Epoch variable. This is used in order to create unique increasing version numbers during fail overs."},
			"cluster_my_epoch":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The Config Epoch of the node we are talking with. This is the current configuration version assigned to this node."},
			"cluster_stats_messages_sent":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of messages sent via the cluster node-to-node binary bus."},
			"cluster_stats_messages_received":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of messages received via the cluster node-to-node binary bus."},
			"total_cluster_links_buffer_limit_exceeded":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Accumulated count of cluster links freed due to exceeding the `cluster-link-sendbuf-limit` configuration."},
			"cluster_stats_messages_ping_sent":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Cluster bus send PING (not to be confused with the client command PING)."},
			"cluster_stats_messages_ping_received":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Cluster bus received PING (not to be confused with the client command PING)."},
			"cluster_stats_messages_pong_sent":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "PONG send (reply to PING)."},
			"cluster_stats_messages_pong_received":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "PONG received (reply to PING)."},
			"cluster_stats_messages_meet_sent":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Handshake message sent to a new node, either through gossip or CLUSTER MEET."},
			"cluster_stats_messages_meet_received":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Handshake message received from a new node, either through gossip or CLUSTER MEET."},
			"cluster_stats_messages_fail_sent":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Mark node xxx as failing send."},
			"cluster_stats_messages_fail_received":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Mark node xxx as failing received."},
			"cluster_stats_messages_publish_sent":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Pub/Sub Publish propagation send."},
			"cluster_stats_messages_publish_received":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Pub/Sub Publish propagation received."},
			"cluster_stats_messages_auth_req_sent":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Replica initiated leader election to replace its master."},
			"cluster_stats_messages_auth_req_received":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Replica initiated leader election to replace its master."},
			"cluster_stats_messages_auth_ack_sent":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Message indicating a vote during leader election."},
			"cluster_stats_messages_auth_ack_received":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Message indicating a vote during leader election."},
			"cluster_stats_messages_update_sent":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Another node slots configuration."},
			"cluster_stats_messages_update_received":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Another node slots configuration."},
			"cluster_stats_messages_mfstart_sent":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Pause clients for manual failover."},
			"cluster_stats_messages_mfstart_received":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Pause clients for manual failover."},
			"cluster_stats_messages_module_sent":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Module cluster API message."},
			"cluster_stats_messages_module_received":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Module cluster API message."},
			"cluster_stats_messages_publishshard_sent":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Pub/Sub Publish shard propagation, see Sharded Pubsub."},
			"cluster_stats_messages_publishshard_received": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Pub/Sub Publish shard propagation, see Sharded Pubsub."},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname"},
			"server_addr":  &inputs.TagInfo{Desc: "Server addr"},
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

	return ipt.parseClusterData(list)
}

func (ipt *Input) parseClusterData(list string) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Now()))

	var kvs point.KVs
	kvs = kvs.AddTag("server_addr", ipt.Addr)

	rdr := strings.NewReader(list)
	scanner := bufio.NewScanner(rdr)

	for scanner.Scan() {
		line := scanner.Text()
		// parts:[cluster_known_nodes 1]
		parts := strings.SplitN(line, ":", 2)

		if len(parts) < 2 {
			continue
		}

		if parts[0] == "cluster_state" {
			if parts[1] == "ok" {
				parts[1] = "1"
			} else {
				parts[1] = "0"
			}
		}

		parts[0] = strings.ReplaceAll(parts[0], "-", "_")

		f, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		if _, has := clusterFieldMap[parts[0]]; has {
			// key in the MeasurementInfo
			kvs = kvs.Add(parts[0], f, false, false)
		}
	}

	if kvs.FieldCount() > 0 {
		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}
		collectCache = append(collectCache, point.NewPointV2(redisCluster, kvs, opts...))
	}
	return collectCache, nil
}

var clusterFieldMap = map[string]struct{}{}

func getClusterFieldMap() {
	m := clusterMeasurement{}
	for k := range m.Info().Fields {
		clusterFieldMap[k] = struct{}{}
	}
}
