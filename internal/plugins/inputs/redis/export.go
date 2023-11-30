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
			"readme": "说明",
			//nolint:lll
			"readme_note":                 "Redis 监控视图显示 Redis 实例运行状态指标，例如客户端连接数、副本连接数、客户端最大链接时长等多种指标，帮助监控分析 Redis 各种异常情况。<br/><br/>https://docs.guance.com/datakit/redis/",
			"connected_clients":           "客户端连接数",
			"connected_slaves":            "副本连接数",
			"max_client_age":              "客户端最大链接时长",
			"command_calls":               "命令执行分布",
			"keyspace_misses":             "Key值查找失败次数",
			"rejected_connections":        "拒绝的连接数",
			"used_cpu":                    "CPU 消耗时间",
			"used_cpu_sys":                "核心态消耗的 CPU 时间",
			"used_cpu_user":               "用户态消耗的 CPU 时间",
			"keyspace_hits":               "命中",
			"blocked_clients":             "等待阻塞命令的客户端连接数",
			"mem_fragmentation_ratio":     "内存碎片率",
			"evicted_keys":                "最大内存限制收回的key数",
			"used_memory":                 "已使用内存",
			"rdb_bgsave_in_progress":      "是否在进行bgsave操作",
			"rdb_bgsave_in_progress_0":    "未运行",
			"rdb_bgsave_in_progress_1":    "正在运行",
			"rdb_changes_since_last_save": "最近转储更改数量",
			"rdb_last_bgsave_time_sec":    "最近RDB保存操作持续时间",
			"keys_sampled":                "key总数",
			"overview":                    "总览,",
			"performance":                 "性能",
			"persistence":                 "持久化",
			"host":                        "主机名",
		}
	case inputs.I18nEn:
		return map[string]string{
			"readme": "Instructions",
			//nolint:lll
			"readme_note":                 "The Redis monitoring view displays various indicators of the running status of Redis instances, such as the number of client connections, number of replica connections, maximum client link duration, etc., to help monitor and analyze various abnormal situations in Redis.<br/>https://docs.guance.com/datakit/redis/",
			"connected_clients":           "Connected clients",
			"connected_slaves":            "Connected slaves",
			"max_client_age":              "Max client age",
			"command_calls":               "Command calls",
			"keyspace_misses":             "Keyspace misses",
			"rejected_connections":        "Rejected connections",
			"used_cpu":                    "Used CPU",
			"used_cpu_sys":                "Used CPU sys",
			"used_cpu_user":               "Used CPU user",
			"keyspace_hits":               "Keyspace hits",
			"blocked_clients":             "Blocked clients",
			"mem_fragmentation_ratio":     "MEM fragmentation ratio",
			"evicted_keys":                "MEM evicted keys",
			"used_memory":                 "MEM used",
			"rdb_bgsave_in_progress":      "RDB bgsave",
			"rdb_bgsave_in_progress_0":    "Stop",
			"rdb_bgsave_in_progress_1":    "Running",
			"rdb_changes_since_last_save": "RDB changes last",
			"rdb_last_bgsave_time_sec":    "RDB last time",
			"keys_sampled":                "Keys sampled",
			"overview":                    "Overview",
			"performance":                 "Performance",
			"persistence":                 "Persistence",
			"host":                        "Host",
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
