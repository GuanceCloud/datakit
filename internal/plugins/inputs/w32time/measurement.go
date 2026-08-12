// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package w32time

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

//nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricName,
		Cat:    point.Metric,
		Desc:   "Windows Time Service status and time synchronization metrics collected through the Windows Performance Data Helper API.",
		DescZh: "通过 Windows Performance Data Helper API 采集的 Windows Time Service 运行状态和时间同步指标。",
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "Host name."},
		},
		Fields: map[string]interface{}{
			fieldServiceRunning:       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Whether the Windows Time Service is running. A value of 1 means running and 0 means stopped."},
			fieldComputedTimeOffset:   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Absolute offset between the system clock and the time source selected by W32Time, in seconds."},
			fieldNTPRoundtripDelay:    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Most recent NTP client round-trip delay reported by W32Time, in seconds."},
			fieldNTPClientSourceCount: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of active NTP time source addresses responding to the Windows NTP client."},
		},
	}
}
