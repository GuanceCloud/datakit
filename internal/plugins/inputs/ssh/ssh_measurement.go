// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ssh

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type SSHMeasurement struct{}

//nolint:lll
func (s *SSHMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Type: "metric",
		Fields: map[string]interface{}{
			"ssh_check": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "SSH service status",
			},
			"sftp_check": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "SFTP service status",
			},
			"sftp_response_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Response time of sftp service",
			},
		},
		Tags: map[string]interface{}{
			"host": inputs.TagInfo{
				Desc: "The host of ssh",
			},
		},
	}
}
