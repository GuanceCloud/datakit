// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collectors

import (
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/objects"
)

const (
	healthyState         = "healthy"
	uptime               = "uptime"
	clusterMembership    = "clusterMembership"
	memoryTotal          = "memoryTotal"
	memoryFree           = "memoryFree"
	mcdMemoryAllocated   = "mcdMemoryAllocated"
	mcdMemoryReserved    = "mcdMemoryReserved"
	interestingStats     = "interestingStats"
	systemStats          = "systemStats"
	interestingStatsTrim = "interestingstats_"
	systemStatsTrim      = "systemstats_"
)

// These are the metrics that we collect per node.
// Including metrics with "InterestingStats" and "SystemStats" prefixes.
// This list allows us to check
// metrics to see if we should collect them per node, or not.
var nodeSpecificStats = []string{healthyState, uptime, clusterMembership, memoryTotal, memoryFree, mcdMemoryAllocated, mcdMemoryReserved}

func (c *Client) nodeCollect() error {
	defer func() { c.config = nil }()
	if c.config == nil {
		c.config = objects.GetNodeCollectorDefaultConfig()
	}

	var nodes objects.Nodes
	err := c.get(c.url("pools/default"), &nodes)
	if err != nil {
		return err
	}

	c.Ctx.ClusterName = nodes.ClusterName

	for key, value := range c.config.Metrics {
		if contains(nodeSpecificStats, key) || strings.HasPrefix(key, interestingStats) || strings.HasPrefix(key, systemStats) {
			for _, node := range nodes.Nodes {
				c.Ctx.NodeHostname = node.Hostname

				switch key {
				case healthyState:
					c.addPoint(c.config.Namespace, getFieldName(value), boolToFloat64(node.Status == healthyState), value.Labels)
				case uptime:
					up := getUptimeValue(node.Uptime, 64)
					c.addPoint(c.config.Namespace, getFieldName(value), up, value.Labels)
				case clusterMembership:
					c.addPoint(c.config.Namespace, getFieldName(value), ifActive(node.ClusterMembership), value.Labels)
				case memoryTotal:
					c.addPoint(c.config.Namespace, getFieldName(value), node.MemoryTotal, value.Labels)
				case memoryFree:
					c.addPoint(c.config.Namespace, getFieldName(value), node.MemoryFree, value.Labels)
				case mcdMemoryAllocated:
					c.addPoint(c.config.Namespace, getFieldName(value), node.McdMemoryAllocated, value.Labels)
				case mcdMemoryReserved:
					c.addPoint(c.config.Namespace, getFieldName(value), node.McdMemoryReserved, value.Labels)
				default:
					if strings.HasPrefix(key, interestingStats) {
						// nolint:lll
						c.addPoint(c.config.Namespace, getFieldName(value), node.InterestingStats[strings.TrimPrefix(value.Name, interestingStatsTrim)], value.Labels)
					} else if strings.HasPrefix(key, systemStats) {
						c.addPoint(c.config.Namespace, getFieldName(value), node.SystemStats[strings.TrimPrefix(value.Name, systemStatsTrim)], value.Labels)
					}
				}
			}
		} else {
			c.addPoint(c.config.Namespace, getFieldName(value), nodes.Counters[value.Name], value.Labels)
		}
	}

	return nil
}

func getUptimeValue(uptime string, bitSize int) float64 {
	up, err := strconv.ParseFloat(uptime, bitSize)
	if err != nil {
		return 0
	}

	return up
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}

	return 0.0
}

func ifActive(s string) float64 {
	if s == "active" {
		return 1.0
	}

	return 0.0
}

func contains(haystack []string, needle string) bool {
	contained := false

	for _, i := range haystack {
		if i == needle {
			contained = true
			break
		}
	}

	return contained
}
