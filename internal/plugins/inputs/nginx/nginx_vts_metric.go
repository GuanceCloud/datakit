// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type ServerZoneMeasurement struct{}

//nolint:lll
func (m *ServerZoneMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementServerZone,
		Cat:    point.Metric,
		Desc:   "NGINX server-zone request, response, traffic, and processing metrics collected from VTS or the NGINX Plus API.",
		DescZh: "通过 VTS 或 NGINX Plus API 采集的 NGINX server zone 请求、响应、流量和处理中请求指标。",
		Fields: map[string]interface{}{
			"requests":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of client requests received by this server zone.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"received":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Total bytes received from clients by this server zone.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"send":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Total bytes sent to clients by this server zone.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"response_1xx": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with 1xx status codes for this server zone.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"response_2xx": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with 2xx status codes for this server zone.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"response_3xx": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with 3xx status codes for this server zone.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"response_4xx": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with 4xx status codes for this server zone.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"response_5xx": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with 5xx status codes for this server zone.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			// nginx plus
			"processing": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of requests being processed in this server zone, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"responses":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses for this server zone, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"discarded":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of requests completed without sending a response for this server zone, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"code_200":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with status code 200 for this server zone, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"code_301":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with status code 301 for this server zone, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"code_404":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with status code 404 for this server zone, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
			"code_503":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with status code 503 for this server zone, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "server_zone"}},
		},
		Tags: map[string]interface{}{
			"nginx_server":  inputs.NewTagInfo("nginx server host"),
			"nginx_port":    inputs.NewTagInfo("nginx server port"),
			"server_zone":   inputs.NewTagInfo("server zone"),
			"host":          inputs.NewTagInfo("host name which installed nginx"),
			"nginx_version": inputs.NewTagInfo("nginx version"),
		},
	}
}

type UpstreamZoneMeasurement struct{}

//nolint:lll
func (m *UpstreamZoneMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementUpstreamZone,
		Cat:    point.Metric,
		Desc:   "NGINX upstream peer request, response, traffic, and peer-state metrics collected from VTS or the NGINX Plus API.",
		DescZh: "通过 VTS 或 NGINX Plus API 采集的 NGINX upstream peer 请求、响应、流量和上游节点状态指标。",
		Fields: map[string]interface{}{
			"request_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of requests sent to this upstream peer.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"received":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Total bytes received from this upstream peer.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"send":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Total bytes sent to this upstream peer.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"response_1xx":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with 1xx status codes from this upstream peer.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"response_2xx":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with 2xx status codes from this upstream peer.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"response_3xx":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with 3xx status codes from this upstream peer.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"response_4xx":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with 4xx status codes from this upstream peer.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"response_5xx":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of responses with 5xx status codes from this upstream peer.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			// nginx plus
			"backup":  &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Whether this upstream peer is configured as a backup server, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"weight":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Configured load-balancing weight for this upstream peer, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"state":   &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Current NGINX Plus state string for this upstream peer.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"active":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of active connections to this upstream peer, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"fails":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of failed attempts to communicate with this upstream peer, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
			"unavail": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of times this upstream peer became unavailable for client requests, reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port", "upstream_zone", "upstream_server"}},
		},
		Tags: map[string]interface{}{
			"nginx_server":    inputs.NewTagInfo("nginx server host"),
			"nginx_port":      inputs.NewTagInfo("nginx server port"),
			"upstream_zone":   inputs.NewTagInfo("upstream zone"),
			"upstream_server": inputs.NewTagInfo("upstream server"),
			"host":            inputs.NewTagInfo("host name which installed nginx"),
			"nginx_version":   inputs.NewTagInfo("nginx version"),
		},
	}
}

type CacheZoneMeasurement struct{}

//nolint:lll
func (m *CacheZoneMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementCacheZone,
		Cat:    point.Metric,
		Desc:   "NGINX cache-zone size, traffic, and cache response classification metrics collected from VTS or the NGINX Plus API.",
		DescZh: "通过 VTS 或 NGINX Plus API 采集的 NGINX cache zone 容量、流量和缓存响应分类指标。",
		Fields: map[string]interface{}{
			"max_size":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Configured maximum size of this cache zone.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"used_size":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Current used size of this cache zone.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"received":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Total bytes received by this cache zone, reported by VTS.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"send":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Total bytes sent by this cache zone, reported by VTS.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"responses_miss":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of cache miss responses for this cache zone.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"responses_bypass":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of cache bypass responses for this cache zone.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"responses_expired":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of cache expired responses for this cache zone.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"responses_stale":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of cache stale responses for this cache zone.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"responses_updating":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of cache updating responses for this cache zone.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"responses_revalidated": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of cache revalidated responses for this cache zone.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"responses_hit":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of cache hit responses for this cache zone.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
			"responses_scarce":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of cache scarce responses for this cache zone, reported by VTS.", Taggedby: []string{"nginx_server", "nginx_port", "cache_zone"}},
		},
		Tags: map[string]interface{}{
			"nginx_server":  inputs.NewTagInfo("nginx server host"),
			"nginx_port":    inputs.NewTagInfo("nginx server port"),
			"cache_zone":    inputs.NewTagInfo("cache zone"),
			"host":          inputs.NewTagInfo("host name which installed nginx"),
			"nginx_version": inputs.NewTagInfo("nginx version"),
		},
	}
}
