// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collectors

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/objects"

func (c *Client) bucketInfoCollect() error {
	defer func() { c.config = nil }()
	if c.config == nil {
		c.config = objects.GetBucketInfoCollectorDefaultConfig()
	}

	var buckets []objects.BucketInfo
	err := c.get(c.url("pools/default/buckets"), &buckets)
	if err != nil {
		return err
	}

	for _, bucket := range buckets {
		c.Ctx.BucketName = bucket.Name

		for key, value := range c.config.Metrics {
			if value.Enabled {
				c.addPoint(c.config.Namespace, getFieldName(value), bucket.BucketBasicStats[key], value.Labels)
			}
		}
	}
	return nil
}
