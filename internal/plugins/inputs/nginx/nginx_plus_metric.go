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
		Name: measurementLocationZone,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"requests":     newByteFieldInfo("The number of requests (only for Nginx plus)"),
			"response":     newByteFieldInfo("The number of response (only for Nginx plus)"),
			"discarded":    newByteFieldInfo("The total number of discarded request (only for Nginx plus)"),
			"received":     newByteFieldInfo("The total number of received bytes (only for Nginx plus)"),
			"sent":         newCountFieldInfo("The total number of send bytes (only for Nginx plus)"),
			"response_1xx": newCountFieldInfo("The number of 1xx response (only for Nginx plus)"),
			"response_2xx": newCountFieldInfo("The number of 2xx response (only for Nginx plus)"),
			"response_3xx": newCountFieldInfo("The number of 3xx response (only for Nginx plus)"),
			"response_4xx": newCountFieldInfo("The number of 4xx response (only for Nginx plus)"),
			"response_5xx": newCountFieldInfo("The number of 5xx response (only for Nginx plus)"),
			"code_200":     newCountFieldInfo("The number of 200 code (only for Nginx plus)"),
			"code_301":     newCountFieldInfo("The number of 301 code (only for Nginx plus)"),
			"code_404":     newCountFieldInfo("The number of 404 code (only for Nginx plus)"),
			"code_503":     newCountFieldInfo("The number of 503 code (only for Nginx plus)"),
		},
		Tags: map[string]interface{}{
			"nginx_server":  inputs.NewTagInfo("nginx server host"),
			"nginx_port":    inputs.NewTagInfo("nginx server port"),
			"location_zone": inputs.NewTagInfo("cache zone"),
			"host":          inputs.NewTagInfo("host name which installed nginx"),
			"nginx_version": inputs.NewTagInfo("nginx version"),
		},
	}
}
