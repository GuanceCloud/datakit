// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ssh

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type SSHMeasurement struct{}

var sshTaggedby = []string{"host"}

//nolint:lll
func (s *SSHMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Cat:    point.Metric,
		Desc:   "SSH and optional SFTP service reachability and response-time checks for a configured remote host.",
		DescZh: "针对配置的远程主机执行 SSH 以及可选 SFTP 服务可达性和响应时间检查。",
		Fields: map[string]interface{}{
			"ssh_check": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.Bool,
				Desc:     "Whether the SSH service check succeeded.",
				Taggedby: sshTaggedby,
			},
			"sftp_check": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.Bool,
				Desc:     "Whether the SFTP service check succeeded.",
				Taggedby: sshTaggedby,
			},
			"sftp_response_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "SFTP service response time in milliseconds.",
				Taggedby: sshTaggedby,
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{
				Desc: "The host of ssh",
			},
		},
	}
}
