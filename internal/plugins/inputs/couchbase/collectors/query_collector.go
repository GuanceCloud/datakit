// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collectors

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/objects"

func (c *Client) queryCollect() error {
	defer func() { c.config = nil }()
	if c.config == nil {
		c.config = objects.GetQueryCollectorDefaultConfig()
	}

	var queryStats objects.Query
	err := c.get(c.url("pools/default/buckets/@query/stats"), &queryStats)
	if err != nil {
		return err
	}

	for _, value := range c.config.Metrics {
		if value.Enabled {
			c.addPoint(c.config.Namespace, getFieldName(value), last(queryStats.Op.Samples[objects.QueryMetricPrefix+value.Name]), value.Labels)
		}
	}

	return nil
}
