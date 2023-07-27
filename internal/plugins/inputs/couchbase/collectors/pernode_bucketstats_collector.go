// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collectors

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/objects"
)

const (
	rebalanceSuccess = "rebalance_success"
)

func (c *Client) perNodeBucketStatsCollect() error {
	defer func() { c.config = nil }()
	if c.config == nil {
		c.config = objects.GetPerNodeBucketStatsCollectorDefaultConfig()
	}

	rebalanced, err := c.getClusterBalancedStatus()
	if err != nil {
		return err
	}

	if !rebalanced {
		// Not error, Waiting for Rebalance... retrying...
		return nil
	}

	var buckets []objects.BucketInfo
	err = c.get(c.url("pools/default/buckets"), &buckets)
	if err != nil {
		return err
	}

	for _, bucket := range buckets {
		c.Ctx.BucketName = bucket.Name
		samples, err := c.getPerNodeBucketStats()
		if err != nil {
			return err
		}

		for _, value := range c.config.Metrics {
			c.setPerNodeBucketStatsMetric(value, samples)
		}
	}

	return nil
}

func (c *Client) setPerNodeBucketStatsMetric(metric objects.MetricInfo, samples map[string]interface{}) {
	if !metric.Enabled {
		return
	}

	stats := strToFloatArr(fmt.Sprint(samples[metric.Name]))
	if len(stats) > 0 {
		c.addPoint(c.config.Namespace, getFieldName(metric), last(stats), metric.Labels)
	}
}

func (c *Client) getClusterBalancedStatus() (bool, error) {
	var node objects.Nodes
	err := c.get(c.url("pools/default"), &node)
	if err != nil {
		return false, fmt.Errorf("unable to retrieve nodes, %w", err)
	}

	return node.Counters[rebalanceSuccess] > 0 || (node.Balanced && node.RebalanceStatus == "none"), nil
}

func strToFloatArr(floatsStr string) []float64 {
	floatsStrArr := strings.Split(floatsStr, " ")

	var floatsArr []float64

	for _, f := range floatsStrArr {
		parse := f

		if strings.HasPrefix(parse, "[") {
			parse = strings.Replace(parse, "[", "", 1)
		}

		if strings.HasSuffix(parse, "]") {
			parse = strings.Replace(parse, "]", "", 1)
		}
		// If the key is omitted from the results (Which we know happens depending on version of CBS), this could be <nil>.
		if strings.Contains(parse, "<nil>") {
			parse = "0.0"
		}

		i, err := strconv.ParseFloat(parse, 64)
		if err == nil {
			floatsArr = append(floatsArr, i)
		} else {
			fmt.Println("")
		}
	}

	return floatsArr
}

func (c *Client) getPerNodeBucketStats() (map[string]interface{}, error) {
	url, err := c.getSpecificNodeBucketStatsURL()
	if err != nil {
		return nil, err
	}

	var bucketStats objects.PerNodeBucketStats
	err = c.get(c.url(url), &bucketStats)
	if err != nil {
		return nil, err
	}

	return bucketStats.Op.Samples, nil
}

func (c *Client) getSpecificNodeBucketStatsURL() (string, error) {
	path := fmt.Sprintf("pools/default/buckets/%s/nodes", c.Ctx.BucketName)
	var servers objects.Servers
	err := c.get(c.url(path), &servers)
	if err != nil {
		return "", err
	}

	correctURI := ""

	for _, server := range servers.Servers {
		if server.Hostname == c.Ctx.NodeHostname {
			correctURI = server.Stats["uri"]
			break
		}
	}

	correctURI = strings.TrimPrefix(correctURI, "/")
	return correctURI, nil
}
