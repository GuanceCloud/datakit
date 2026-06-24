// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type LocationZoneMeasurement struct{}

func (m *LocationZoneMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementLocationZone,
		Cat:    point.Metric,
		Desc:   "NGINX Plus location-zone request, response, and traffic metrics.",
		DescZh: "通过 NGINX Plus API 采集的 location zone 请求、响应和流量指标。",
		Fields: map[string]interface{}{
			"requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of requests processed by this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"response": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of responses sent by this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"discarded": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of requests completed without sending a response for this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.SizeByte,
				Desc:     "Total bytes received by this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.SizeByte,
				Desc:     "Total bytes sent by this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"response_1xx": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of responses with 1xx status codes for this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"response_2xx": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of responses with 2xx status codes for this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"response_3xx": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of responses with 3xx status codes for this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"response_4xx": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of responses with 4xx status codes for this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"response_5xx": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of responses with 5xx status codes for this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"code_200": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of responses with status code 200 for this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"code_301": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of responses with status code 301 for this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"code_404": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of responses with status code 404 for this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
			"code_503": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Total number of responses with status code 503 for this location zone.",
				Taggedby: []string{"nginx_server", "nginx_port", "location_zone"},
			},
		},
		Tags: map[string]interface{}{
			"nginx_server":  inputs.NewTagInfo("nginx server host"),
			"nginx_port":    inputs.NewTagInfo("nginx server port"),
			"location_zone": inputs.NewTagInfo("NGINX Plus location zone name"),
			"host":          inputs.NewTagInfo("host name which installed nginx"),
			"nginx_version": inputs.NewTagInfo("nginx version"),
		},
	}
}
