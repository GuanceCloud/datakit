// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collectors

import (
	"errors"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/objects"
)

func (c *Client) indexCollect() error {
	defer func() { c.config = nil }()
	if c.config == nil {
		c.config = objects.GetIndexCollectorDefaultConfig()
	}

	var indexStats objects.Index
	err := c.get(c.url("pools/default/buckets/@index/stats"), &indexStats)
	if err != nil {
		return err
	}

	currentNode, err := c.getCurrentNode()
	if err != nil {
		return err
	}

	if contains(currentNode.Services, "index") {
		var stats map[string]map[string]interface{}
		err := c.get(c.indexerURL("api/v1/stats"), &stats)
		if err != nil {
			return err
		}

		for _, value := range c.config.Metrics {
			if value.Enabled && !contains(value.Labels, objects.KeyspaceLabel) {
				c.addPoint(c.config.Namespace, getFieldName(value), last(indexStats.Op.Samples[objects.IndexMetricPrefix+value.Name]), value.Labels)
			} else {
				for key, values := range stats {
					c.Ctx.Keyspace = key
					if key == "indexer" {
						continue
					}

					val, ok := values[value.Name].(float64)

					if !ok {
						continue
					}

					c.addPoint(c.config.Namespace, getFieldName(value), val, value.Labels)
				}
			}
		}
	} else {
		for _, value := range c.config.Metrics {
			if value.Enabled && !contains(value.Labels, objects.KeyspaceLabel) {
				c.addPoint(c.config.Namespace, getFieldName(value), last(indexStats.Op.Samples[objects.IndexMetricPrefix+value.Name]), value.Labels)
			}
		}
	}

	return nil
}

// potentially deprecated.
func (c *Client) getCurrentNode() (objects.Node, error) {
	var nodes objects.Nodes
	err := c.get(c.url("pools/default"), &nodes)

	var retNode objects.Node

	if err != nil {
		return retNode, fmt.Errorf("unable to retrieve nodes, %w", err)
	}

	for _, node := range nodes.Nodes {
		if node.ThisNode {
			retNode = node
			return retNode, nil // hostname seems to work? just don't use for single node setups
		}
	}

	return retNode, errors.New("sidecar container cannot find Couchbase Hostname")
}
