// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ssh

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type SSHMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (s *SSHMeasurement) LineProto() (*point.Point, error) {
	data, err := point.NewPoint(s.name, s.tags, s.fields, point.MOpt())
	return data, err
}

//nolint:lll
func (s *SSHMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"ssh_check":          &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "ssh service status"},
			"ssh_err":            &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "fail reason of connect ssh service"},
			"sftp_check":         &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "sftp service status"},
			"sftp_err":           &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "fail reason of connect sftp service"},
			"sftp_response_time": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "response time of sftp service"},
		},
		Tags: map[string]interface{}{
			"host": inputs.TagInfo{Desc: "the host of ssh"},
		},
	}
}
