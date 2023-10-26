// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collectors

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/objects"
)

func (c *Client) addPoint(measurementName, field string, val any, labels []string) {
	var kvs point.KVs
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(c.ts))

	kvs = kvs.Add(field, val, false, true)

	for k, v := range c.getTags(labels) {
		kvs = kvs.AddTag(k, v)
	}

	for k, v := range c.MergedTags {
		kvs = kvs.AddTag(k, v)
	}

	pt := point.NewPointV2(measurementName, kvs, opts...)
	c.Pts = append(c.Pts, pt)
}

func (c *Client) getTags(labels []string) map[string]string {
	tags := make(map[string]string)

	for _, label := range labels {
		switch label {
		case objects.ClusterLabel:
			tags[label] = c.Ctx.ClusterName
		case objects.BucketLabel:
			tags[label] = c.Ctx.BucketName
		case objects.KeyspaceLabel:
			tags[label] = c.Ctx.Keyspace
		case objects.NodeLabel:
			tags[label] = c.Ctx.NodeHostname
		case objects.TargetLabel:
			tags[label] = c.Ctx.Target
		case objects.SourceLabel:
			tags[label] = c.Ctx.Source
		default:
			if strings.Contains(label, ":") {
				splits := strings.Split(label, ":")
				tags[splits[0]] = splits[1]
			}
		}
	}

	return tags
}

func (c *Client) get(u string, v interface{}) error {
	resp, err := c.request(u)
	if err != nil {
		return fmt.Errorf("failed to Get %s : %w", u, err)
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body from %s : %w", u, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to Get 200 response status from %s : %d", u, resp.StatusCode)
	}

	if err := json.Unmarshal(bts, v); err != nil {
		return fmt.Errorf("failed to unmarshal %s : %w ", u, err)
	}

	return nil
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("%s://%s:%d/%s", c.Opt.Scheme, c.Opt.Host, c.Opt.Port, path)
}

func (c *Client) indexerURL(path string) string {
	return fmt.Sprintf("%s://%s:%d/%s", c.Opt.Scheme, c.Opt.Host, c.Opt.AdditionalPort, path)
}

func last(stats []float64) float64 {
	if len(stats) == 0 {
		return 0
	}

	return stats[len(stats)-1]
}

func min(x, y float64) float64 {
	if x > y {
		return y
	}

	return x
}

func getFieldName(v objects.MetricInfo) string {
	if v.NameOverride != "" {
		return v.NameOverride
	}
	return v.Name
}
