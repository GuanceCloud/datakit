// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type ChangeMeasurement struct{}

func (*ChangeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Cat:  point.KeyEvent,
		Name: "Change Event",
		Desc: "",
		Tags: map[string]any{
			"host":          &inputs.TagInfo{Desc: "System hostname."},
			"change_id":     &inputs.TagInfo{Desc: "Unique identifier for the user or group change event."},
			"df_event_id":   &inputs.TagInfo{Desc: "Event ID."},
			"df_source":     &inputs.TagInfo{Desc: "Source name."},
			"df_title":      &inputs.TagInfo{Desc: "Event title."},
			"df_message":    &inputs.TagInfo{Desc: "Event message."},
			"df_status":     &inputs.TagInfo{Desc: "Event status."},
			"df_sub_status": &inputs.TagInfo{Desc: "Event detail status."},
		},
		Fields: map[string]any{
			"change_time_us": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.TimestampUS,
				Desc:     "Timestamp of the change event in microseconds.",
			},
		},
	}
}
