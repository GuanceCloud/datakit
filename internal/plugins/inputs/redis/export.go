// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
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

func (ipt *Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
			"message":      `>等级：{{status}}  \n>主机：{{host}}  \n>内容：等待阻塞命令的客户端连接数为 {{ Result }}\n>建议：延迟或其他问题可能会阻止源列表被填充。虽然被阻止的客户端本身不会引起警报，但如果您看到此指标的值始终为非零值，则应该引起注意。`,
			"title":        `主机 {{ host }} Redis 等待阻塞命令的客户端连接数异常增加`,
			"monitor_name": "Redis 检测库",
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
			"message":      `>Level: {{status}}  \n>Host: {{host}}  \n>Content: The number of client connections waiting for blocking commands is {{ Result }}. \n>Suggest: Delays or other issues may prevent the source list from being populated. While blocked clients by themselves do not cause alarm, if you see a consistently non-zero value for this metric, it should be a cause for concern.`,
			"title":        `The number of Redis client connections waiting for blocking commands on Host {{ host }} increased abnormally.`,
			"monitor_name": "Redis check",
		}
	default:
		return nil
	}
}
