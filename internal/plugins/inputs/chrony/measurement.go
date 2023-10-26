// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package chrony

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// Info for docs and integrate testing.
// nolint:lll
func (docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"system_time":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is the current offset between the NTP clock and system clock."},
			"last_offset":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is the estimated local offset on the last clock update."},
			"rms_offset":      &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is a long-term average of the offset value."},
			"frequency":       &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PartPerMillion, Desc: "This is the rate by which the system clock would be wrong if *chronyd* was not correcting it."},
			"residual_freq":   &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PartPerMillion, Desc: "This shows the residual frequency for the currently selected reference source."},
			"skew":            &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PartPerMillion, Desc: "This is the estimated error bound on the frequency."},
			"root_delay":      &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is the total of the network path delays to the stratum-1 computer from which the computer is ultimately synchronized."},
			"root_dispersion": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is the total dispersion accumulated through all the computers back to the stratum-1 computer from which the computer is ultimately synchronized."},
			"update_interval": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is the interval between the last two clock updates."},
		},

		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Host name"},
			"reference_id": &inputs.TagInfo{Desc: "This is the reference ID and name (or IP address) of the server to which the computer is currently synchronized."},
			"stratum":      &inputs.TagInfo{Desc: "The stratum indicates how many hops away from a computer with an attached reference clock we are."},
			"leap_status":  &inputs.TagInfo{Desc: "This is the leap status, which can be Normal, Insert second, Delete second or Not synchronized."},
		},
	}
}
