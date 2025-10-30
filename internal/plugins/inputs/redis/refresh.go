// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// checkNodesChanged checks if the node topology has changed.
func (ipt *Input) checkNodesChanged(ctx context.Context) bool {
	if ipt.Cluster != nil && ipt.crdb != nil {
		return ipt.hasClusterNodesChanged(ctx)
	}

	if ipt.MasterSlave != nil && ipt.MasterSlave.Sentinel != nil && ipt.srdb != nil {
		return ipt.hasSentinelNodesChanged(ctx)
	}

	return false
}

// reportTopologyChange reports Redis topology change events.
func (ipt *Input) reportTopologyChange(mode, changeType, message string, cost time.Duration) {
	opts := append(point.DefaultLoggingOptions(), point.WithTime(ntp.Now()))

	var kvs point.KVs
	kvs = kvs.Set("message", message).
		AddTag("mode", mode).
		AddTag("change_type", changeType)

	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	pt := point.NewPoint(measureuemtRedisTopology, kvs, opts...)

	if err := ipt.feeder.Feed(point.Logging, []*point.Point{pt},
		dkio.WithCollectCost(cost),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(dkio.FeedSource(inputName, "topology")),
	); err != nil {
		l.Warnf("failed to report topology change: %s", err)
	}
}

// hasClusterNodesChanged checks if cluster mode nodes have changed.
func (ipt *Input) hasClusterNodesChanged(ctx context.Context) bool {
	start := time.Now()
	l.Debugf("refresh: scan cluster masters")
	newInstances, err := ipt.scanClusterMasters(ctx)
	if err != nil {
		l.Warnf("scan cluster masters failed: %s", err)
		return false
	}

	// Compare master node count
	if len(ipt.instances) != len(newInstances) {
		message := fmt.Sprintf("Cluster master count changed from %d to %d",
			len(ipt.instances), len(newInstances))
		l.Info(message)
		ipt.reportTopologyChange("cluster", "master_count_changed", message, time.Since(start))
		return true
	}

	oldMasters := make(map[string]bool)
	for _, inst := range ipt.instances {
		oldMasters[inst.addr] = true
	}

	newMasters := make(map[string]bool)
	for _, inst := range newInstances {
		newMasters[inst.addr] = true
	}

	var removedMasters, addedMasters []string
	for addr := range oldMasters {
		if !newMasters[addr] {
			removedMasters = append(removedMasters, addr)
		}
	}
	for addr := range newMasters {
		if !oldMasters[addr] {
			addedMasters = append(addedMasters, addr)
		}
	}

	if len(removedMasters) > 0 || len(addedMasters) > 0 {
		message := fmt.Sprintf("Cluster masters changed - removed: [%s], added: [%s]",
			strings.Join(removedMasters, ", "),
			strings.Join(addedMasters, ", "))
		l.Info(message)
		ipt.reportTopologyChange("cluster", "master_nodes_changed", message, time.Since(start))
		return true
	}

	oldReplicaMap := make(map[string][]string)
	for _, inst := range ipt.instances {
		replicas := make([]string, len(inst.replicas))
		for i, rep := range inst.replicas {
			replicas[i] = rep.addr
		}
		sort.Strings(replicas)
		oldReplicaMap[inst.addr] = replicas
	}

	newReplicaMap := make(map[string][]string)
	for _, inst := range newInstances {
		replicas := make([]string, len(inst.replicas))
		for i, rep := range inst.replicas {
			replicas[i] = rep.addr
		}
		sort.Strings(replicas)
		newReplicaMap[inst.addr] = replicas
	}

	for masterAddr, oldReplicas := range oldReplicaMap {
		newReplicas, ok := newReplicaMap[masterAddr]
		if !ok {
			continue
		}

		if len(oldReplicas) != len(newReplicas) {
			message := fmt.Sprintf("Cluster master %s replica count changed from %d to %d",
				masterAddr, len(oldReplicas), len(newReplicas))
			l.Info(message)
			ipt.reportTopologyChange("cluster", "replica_count_changed", message, time.Since(start))
			return true
		}

		oldSet := make(map[string]bool)
		for _, addr := range oldReplicas {
			oldSet[addr] = true
		}
		newSet := make(map[string]bool)
		for _, addr := range newReplicas {
			newSet[addr] = true
		}

		var removed, added []string
		for addr := range oldSet {
			if !newSet[addr] {
				removed = append(removed, addr)
			}
		}
		for addr := range newSet {
			if !oldSet[addr] {
				added = append(added, addr)
			}
		}

		if len(removed) > 0 || len(added) > 0 {
			message := fmt.Sprintf("Cluster master %s replicas changed - removed: [%s], added: [%s]",
				masterAddr,
				strings.Join(removed, ", "),
				strings.Join(added, ", "))
			l.Info(message)
			ipt.reportTopologyChange("cluster", "replica_nodes_changed", message, time.Since(start))
			return true
		}
	}

	return false
}

// hasSentinelNodesChanged checks if sentinel mode nodes have changed.
func (ipt *Input) hasSentinelNodesChanged(ctx context.Context) bool {
	start := time.Now()
	l.Debugf("refresh: discover sentinel master")
	newInst, err := ipt.sentinelDiscoverMaster(ctx)
	if err != nil {
		l.Warnf("sentinel discover failed: %s", err)
		return false
	}

	if len(ipt.instances) == 0 {
		return false
	}

	oldInst := ipt.instances[0]

	if oldInst.addr != newInst.addr {
		message := fmt.Sprintf("Sentinel master failover detected: %s -> %s",
			oldInst.addr, newInst.addr)
		l.Info(message)
		ipt.reportTopologyChange("sentinel", "master_failover", message, time.Since(start))
		return true
	}

	if len(oldInst.replicas) != len(newInst.replicas) {
		message := fmt.Sprintf("Sentinel replica count changed from %d to %d",
			len(oldInst.replicas), len(newInst.replicas))
		l.Info(message)
		ipt.reportTopologyChange("sentinel", "replica_count_changed", message, time.Since(start))
		return true
	}

	oldReplicas := make([]string, len(oldInst.replicas))
	for i, rep := range oldInst.replicas {
		oldReplicas[i] = rep.addr
	}
	sort.Strings(oldReplicas)

	newReplicas := make([]string, len(newInst.replicas))
	for i, rep := range newInst.replicas {
		newReplicas[i] = rep.addr
	}
	sort.Strings(newReplicas)

	oldSet := make(map[string]bool)
	for _, addr := range oldReplicas {
		oldSet[addr] = true
	}
	newSet := make(map[string]bool)
	for _, addr := range newReplicas {
		newSet[addr] = true
	}

	var removed, added []string
	for addr := range oldSet {
		if !newSet[addr] {
			removed = append(removed, addr)
		}
	}
	for addr := range newSet {
		if !oldSet[addr] {
			added = append(added, addr)
		}
	}

	if len(removed) > 0 || len(added) > 0 {
		message := fmt.Sprintf("Sentinel replicas changed - removed: [%s], added: [%s]",
			strings.Join(removed, ", "),
			strings.Join(added, ", "))
		l.Info(message)
		ipt.reportTopologyChange("sentinel", "replica_nodes_changed", message, time.Since(start))
		return true
	}

	return false
}

// topologyChangeMeasurement defines the topology change logging measurement.
type topologyChangeMeasurement struct{}

//nolint:lll
func (m *topologyChangeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "redis_topology",
		Cat:    point.Logging,
		Desc:   "Redis topology events detected by monitoring cluster/sentinel nodes.",
		DescZh: "Redis 拓扑变化事件，通过监控集群/哨兵节点检测。",
		Tags: map[string]interface{}{
			"mode": &inputs.TagInfo{
				Desc: "Redis mode: `cluster` or `sentinel`",
			},
			"change_type": &inputs.TagInfo{
				Desc: "Change type: `master_count_changed`, `master_nodes_changed`, `master_failover`, `replica_count_changed`, `replica_nodes_changed`",
			},
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Detailed description of the topology change event",
			},
		},
	}
}
