// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collectors

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/objects"
)

func (c *Client) bucketStatusCollect() error {
	defer func() { c.config = nil }()
	if c.config == nil {
		c.config = objects.GetBucketStatsCollectorDefaultConfig()
	}

	var buckets []objects.BucketInfo
	err := c.get(c.url("pools/default/buckets"), &buckets)
	if err != nil {
		return err
	}

	for _, bucket := range buckets {
		c.Ctx.BucketName = bucket.Name

		var stats objects.BucketStats
		path := fmt.Sprintf("pools/default/buckets/%s/stats", bucket.Name)
		err := c.get(c.url(path), &stats)
		if err != nil {
			return err
		}

		for _, value := range c.config.Metrics {
			if value.Enabled {
				c.setBucketStatusMetric(value, stats.Op.Samples)
			}
		}
	}

	return nil
}

func (c *Client) setBucketStatusMetric(metric objects.MetricInfo, samples map[string][]float64) {
	if !metric.Enabled {
		return
	}

	switch metric.Name {
	case "avg_bg_wait_time":
		// comes across as microseconds.  Convert
		c.addPoint(c.config.Namespace, getFieldName(metric), last(samples[metric.Name])/1000000, metric.Labels)
	case "ep_cache_miss_rate":
		c.addPoint(c.config.Namespace, getFieldName(metric), min(last(samples[metric.Name]), 100), metric.Labels)
	default:
		c.addPoint(c.config.Namespace, getFieldName(metric), last(samples[metric.Name]), metric.Labels)
	}
}
