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
		Name: measurementServerZone,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"requests":     newCountFieldInfo("The total number of client requests received from clients."),
			"received":     newByteFieldInfo("The total amount of data received from clients."),
			"send":         newByteFieldInfo("The total amount of data sent to clients."),
			"response_1xx": newCountFieldInfo("The number of responses with status codes 1xx"),
			"response_2xx": newCountFieldInfo("The number of responses with status codes 2xx"),
			"response_3xx": newCountFieldInfo("The number of responses with status codes 3xx"),
			"response_4xx": newCountFieldInfo("The number of responses with status codes 4xx"),
			"response_5xx": newCountFieldInfo("The number of responses with status codes 5xx"),
			// nginx plus
			"processing": newCountFieldInfo("The number of requests being processed (only for Nginx plus)"),
			"responses":  newCountFieldInfo("The total number of responses (only for Nginx plus)"),
			"discarded":  newCountFieldInfo("The number of requests being discarded (only for Nginx plus)"),
			"code_200":   newCountFieldInfo("The number of responses with status code 200 (only for Nginx plus)"),
			"code_301":   newCountFieldInfo("The number of responses with status code 301 (only for Nginx plus)"),
			"code_404":   newCountFieldInfo("The number of responses with status code 404 (only for Nginx plus)"),
			"code_503":   newCountFieldInfo("The number of responses with status code 503 (only for Nginx plus)"),
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
		Name: measurementUpstreamZone,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"request_count": newCountFieldInfo("The total number of client requests received from server."),
			"received":      newByteFieldInfo("The total number of bytes received from this server."),
			"send":          newByteFieldInfo("The total number of bytes sent to clients."),
			"response_1xx":  newCountFieldInfo("The number of responses with status codes 1xx"),
			"response_2xx":  newCountFieldInfo("The number of responses with status codes 2xx"),
			"response_3xx":  newCountFieldInfo("The number of responses with status codes 3xx"),
			"response_4xx":  newCountFieldInfo("The number of responses with status codes 4xx"),
			"response_5xx":  newCountFieldInfo("The number of responses with status codes 5xx"),
			// nginx plus
			"backup":  newCountFieldInfo("Whether it is configured as a backup server (only for Nginx plus)"),
			"weight":  newCountFieldInfo("Weights used when load balancing (only for Nginx plus)"),
			"state":   newCountFieldInfo("The current state of the server (only for Nginx plus)"),
			"active":  newCountFieldInfo("The number of active connections (only for Nginx plus)"),
			"fails":   newCountFieldInfo("The number of failed requests (only for Nginx plus)"),
			"unavail": newCountFieldInfo("The number of unavailable server (only for Nginx plus)"),
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
		Name: measurementCacheZone,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"max_size":              newByteFieldInfo("The limit on the maximum size of the cache specified in the configuration"),
			"used_size":             newByteFieldInfo("The current size of the cache."),
			"received":              newByteFieldInfo("The total number of bytes received from the cache."),
			"send":                  newByteFieldInfo("The total number of bytes sent from the cache."),
			"responses_miss":        newCountFieldInfo("The number of cache miss"),
			"responses_bypass":      newCountFieldInfo("The number of cache bypass"),
			"responses_expired":     newCountFieldInfo("The number of cache expired"),
			"responses_stale":       newCountFieldInfo("The number of cache stale"),
			"responses_updating":    newCountFieldInfo("The number of cache updating"),
			"responses_revalidated": newCountFieldInfo("The number of cache revalidated"),
			"responses_hit":         newCountFieldInfo("The number of cache hit"),
			"responses_scarce":      newCountFieldInfo("The number of cache scarce"),
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
