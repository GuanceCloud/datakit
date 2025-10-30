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

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (i *instance) collectCluster(ctx context.Context) {
	collectStart := time.Now()

	list, err := i.curCli.clusterInfo(ctx)
	if err != nil {
		l.Warnf("get clusterinfo error: %s ignored", err)
		return
	}

	pts, err := i.parseClusterData(list)
	if err != nil {
		l.Warnf("parseClusterData: %s, ignored", err.Error())
		return
	}

	if err := i.ipt.feeder.Feed(point.Metric, pts,
		dkio.WithCollectCost(time.Since(collectStart)),
		dkio.WithElection(i.ipt.Election),
		dkio.WithSource(dkio.FeedSource(inputName, "cluster")),
		dkio.WithMeasurement(inputs.GetOverrideMeasurement(i.ipt.MeasurementVersion, measureuemtRedis)),
	); err != nil {
		l.Warnf("feed measurement: %s, ignored", err)
	}
}

func (i *instance) parseClusterData(list string) ([]*point.Point, error) {
	var (
		pts     = []*point.Point{}
		opts    = append(point.DefaultMetricOptions(), point.WithTime(i.ipt.ptsTime))
		kvs     point.KVs
		rdr     = strings.NewReader(list)
		scanner = bufio.NewScanner(rdr)
	)

	for scanner.Scan() {
		line := scanner.Text()
		arr := strings.SplitN(line, ":", 2)

		arr[0] = strings.ReplaceAll(arr[0], "-", "_")

		if len(arr) < 2 {
			l.Warnf("ignore line %q", line)
			continue
		}

		switch arr[1] {
		case "ok":
			kvs = kvs.Add(arr[0], 1.0)
		case "fail":
			kvs = kvs.Add(arr[0], 0.0)
		default:
			num, err := strconv.ParseFloat(arr[1], 64)
			if err == nil {
				kvs = kvs.Add(arr[0], num)
			} else {
				l.Warnf("ignore line %q: %s", line, err.Error())
			}
		}
	}

	if len(kvs) > 0 {
		for k, v := range i.mergedTags {
			kvs = kvs.AddTag(k, v)
		}
		pts = append(pts, point.NewPoint(measureuemtRedisCluster, kvs, opts...))
	}

	return pts, nil
}

type clusterMeasurement struct{}

// see also: https://redis.io/commands/cluster-info/
//
//nolint:lll
func (m *clusterMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measureuemtRedisCluster,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"cluster_state": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.EnumValue,
				Desc:     "State is 1(ok) if the node is able to receive queries. 0(fail) if there is at least one hash slot which is unbound (no node associated), in error state (node serving it is flagged with FAIL flag), or if the majority of masters can't be reached by this node.",
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
				Unit:     inputs.NoUnit,
				Desc:     "The local Current Epoch variable. This is used in order to create unique increasing version numbers during fail overs.",
			},
			"cluster_my_epoch": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The Config Epoch of the node we are talking with. This is the current configuration version assigned to this node.",
			},
			"cluster_stats_messages_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Number of messages sent via the cluster node-to-node binary bus.",
			},
			"cluster_stats_messages_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Number of messages received via the cluster node-to-node binary bus.",
			},
			"total_cluster_links_buffer_limit_exceeded": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Accumulated count of cluster links freed due to exceeding the `cluster-link-sendbuf-limit` configuration.",
			},
			"cluster_stats_messages_ping_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Cluster bus send PING (not to be confused with the client command PING).",
			},
			"cluster_stats_messages_ping_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Cluster bus received PING (not to be confused with the client command PING).",
			},
			"cluster_stats_messages_pong_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "PONG send (reply to PING).",
			},
			"cluster_stats_messages_pong_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "PONG received (reply to PING).",
			},
			"cluster_stats_messages_meet_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Handshake message sent to a new node, either through gossip or CLUSTER MEET.",
			},
			"cluster_stats_messages_meet_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Handshake message received from a new node, either through gossip or CLUSTER MEET.",
			},
			"cluster_stats_messages_fail_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Mark node xxx as failing send.",
			},
			"cluster_stats_messages_fail_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Mark node xxx as failing received.",
			},
			"cluster_stats_messages_publish_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Pub/Sub Publish propagation send.",
			},
			"cluster_stats_messages_publish_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Pub/Sub Publish propagation received.",
			},
			"cluster_stats_messages_auth_req_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Replica initiated leader election to replace its master.",
			},
			"cluster_stats_messages_auth_req_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Replica initiated leader election to replace its master.",
			},
			"cluster_stats_messages_auth_ack_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Message indicating a vote during leader election.",
			},
			"cluster_stats_messages_auth_ack_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Message indicating a vote during leader election.",
			},
			"cluster_stats_messages_update_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Another node slots configuration.",
			},
			"cluster_stats_messages_update_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Another node slots configuration.",
			},
			"cluster_stats_messages_mfstart_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Pause clients for manual failover.",
			},
			"cluster_stats_messages_mfstart_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Pause clients for manual failover.",
			},
			"cluster_stats_messages_module_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Module cluster API message.",
			},
			"cluster_stats_messages_module_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Module cluster API message.",
			},
			"cluster_stats_messages_publishshard_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Pub/Sub Publish shard propagation, see Sharded Pubsub.",
			},
			"cluster_stats_messages_publishshard_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Pub/Sub Publish shard propagation, see Sharded Pubsub.",
			},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname"},
			"server_addr":  &inputs.TagInfo{Desc: "Server addr"},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
		},
	}
}
