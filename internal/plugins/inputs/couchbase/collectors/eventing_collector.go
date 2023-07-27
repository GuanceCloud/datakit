// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collectors

import (
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/objects"
)

func (c *Client) eventingCollect() error {
	defer func() { c.config = nil }()
	if c.config == nil {
		c.config = objects.GetEventingCollectorDefaultConfig()
	}

	var ev objects.Eventing
	err := c.get(c.url("pools/default/buckets/@eventing/stats"), &ev)
	if err != nil {
		return err
	}

	for _, value := range c.config.Metrics {
		if value.Enabled {
			sampleName := objects.EventingMetricPrefix
			if strings.HasPrefix(value.Name, "test_") {
				sampleName += strings.ReplaceAll(value.Name, "test_", "test/")
			} else {
				sampleName += value.Name
			}
			c.addPoint(c.config.Namespace, getFieldName(value), last(ev.Op.Samples[sampleName]), value.Labels)
		}
	}

	return nil
}
