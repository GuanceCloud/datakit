// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package chrony

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

var chronyTaggedby = []string{"reference_id", "stratum", "leap_status"}

// Info for docs and integrate testing.
// nolint:lll
func (docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricName,
		Cat:    point.Metric,
		Desc:   "Chrony tracking metrics parsed from `chronyc -n tracking`, describing the local system clock state relative to its current NTP reference source.",
		DescZh: "从 `chronyc -n tracking` 输出解析出的 Chrony 跟踪指标，用于描述本地系统时钟相对当前 NTP 参考源的状态。",
		Fields: map[string]interface{}{
			"system_time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationSecond,
				Desc: "Signed current offset between the system clock and NTP time, in seconds. The collector records values reported as `slow` as negative and values reported as `fast` as positive.", Taggedby: chronyTaggedby,
			},
			"last_offset": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationSecond,
				Desc: "Estimated local clock offset applied at the last clock update, in seconds.", Taggedby: chronyTaggedby,
			},
			"rms_offset": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationSecond,
				Desc: "Long-term root-mean-square average of clock offset values, in seconds.", Taggedby: chronyTaggedby,
			},
			"frequency": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PartPerMillion,
				Desc: "Estimated system clock frequency error that chronyd is correcting, in parts per million.", Taggedby: chronyTaggedby,
			},
			"residual_freq": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PartPerMillion,
				Desc: "Residual frequency error for the currently selected reference source, in parts per million.", Taggedby: chronyTaggedby,
			},
			"skew": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PartPerMillion,
				Desc: "Estimated error bound of the frequency estimate, in parts per million.", Taggedby: chronyTaggedby,
			},
			"root_delay": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationSecond,
				Desc: "Total network path delay to the stratum-1 source through the synchronization chain, in seconds.", Taggedby: chronyTaggedby,
			},
			"root_dispersion": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationSecond,
				Desc: "Total dispersion accumulated through the synchronization chain back to the stratum-1 source, in seconds.", Taggedby: chronyTaggedby,
			},
			"update_interval": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationSecond,
				Desc: "Interval between the last two local clock updates, in seconds.", Taggedby: chronyTaggedby,
			},
		},

		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Host name associated with the collected chrony instance."},
			"reference_id": &inputs.TagInfo{Desc: "Reference ID of the server or source to which the system is currently synchronized."},
			"stratum":      &inputs.TagInfo{Desc: "Distance in NTP strata from the ultimate reference clock."},
			"leap_status":  &inputs.TagInfo{Desc: "Leap-second synchronization status reported by chronyc, such as Normal, Insert second, Delete second, or Not synchronized."},
		},
	}
}
