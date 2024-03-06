// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ssh

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
			"ssh_check":          "SSH 服务状态",
			"sftp_check":         "SFTP 服务状态",
			"sftp_response_time": "SFTP 服务响应时间",
			"levels_normal":      "正常",
			"levels_abnormal":    "异常",
			"host_name":          "主机名",
			"title":              "SSH 监控视图",
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
			"ssh_check":          "SSH service status",
			"sftp_check":         "SFTP service status",
			"sftp_response_time": "Response time of sftp service",
			"levels_normal":      "Normal",
			"levels_abnormal":    "Abnormal",
			"host_name":          "Host name",
			"title":              "SSH",
		}
	default:
		return nil
	}
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
		}
	default:
		return nil
	}
}
