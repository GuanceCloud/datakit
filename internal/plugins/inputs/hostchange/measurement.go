// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

//nolint:lll,goconst,funlen // Measurement metadata contains intentionally long repeated literals.
package hostchange

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type ChangeMeasurement struct{}

func (*ChangeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Cat:    point.KeyEvent,
		Name:   "Change Event",
		Desc:   "Host change events collected from the local system, including configured file, crontab, network, service, user, and group changes.",
		DescZh: "从本机采集的主机变更事件，包括已配置的文件、crontab、网络、服务、用户和用户组变更。",
		Tags: map[string]any{
			"host":          &inputs.TagInfo{Desc: "System hostname."},
			"change_id":     &inputs.TagInfo{Desc: "Unique identifier for the host change event."},
			"df_event_id":   &inputs.TagInfo{Desc: "Event ID."},
			"df_source":     &inputs.TagInfo{Desc: "Source name."},
			"df_status":     &inputs.TagInfo{Desc: "Event status."},
			"df_sub_status": &inputs.TagInfo{Desc: "Event detail status."},
		},
		Fields: map[string]any{
			"df_title":   &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Title text summarizing the host change event."},
			"df_message": &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Message text describing the host change event."},
			"change_time_us": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.TimestampUS,
				Desc:     "Timestamp of the change event in microseconds.",
			},
		},
	}
}
