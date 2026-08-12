// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"group_overview":        "概览",
			"group_points":          "拨测数据发送情况",
			"group_tasks":           "拨测任务",
			"worker_job_total":      "最大同时发送数据数",
			"worker_job_chan_total": "待发送数据通道容量",
			"worker_job_chan_used":  "待发送数据通道已使用",
			"worker_cached_points":  "内存中缓存的数据数",
			"task_running_number":   "拨测任务运行数",
			"task_invalid":          "无效任务数",
			"task_pulled":           "已同步任务数",
			"task_pull_cost":        "同步任务耗时",
			"dataway_sent_failed":   "Dataway 发送失败",
			"points_sending":        "发送中的数据",
			"points_sent_cost":      "数据发送耗时",
			"points_sent_ok":        "发送成功的数据",
			"points_sent_failed":    "发送失败的数据",
		}
	case inputs.I18nEn:
		return map[string]string{
			"group_overview":        "Overview",
			"group_points":          "Points",
			"group_tasks":           "Tasks",
			"worker_job_total":      "Worker job total",
			"worker_job_chan_total": "Worker channel total",
			"worker_job_chan_used":  "Worker channel used",
			"worker_cached_points":  "Worker cache points",
			"task_pull_cost":        "Task pull cost",
			"task_running_number":   "Task running",
			"task_invalid":          "Invalid task",
			"task_pulled":           "Task pulled",
			"dataway_sent_failed":   "Dataway sent failed",
			"points_sending":        "Points sending",
			"points_sent_cost":      "Points sent cost",
			"points_sent_ok":        "Points sent ok",
			"points_sent_failed":    "Points sent failed",
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
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
		}
	default:
		return nil
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	infos := []*inputs.ENVInfo{
		{
			ENVName:   "DISABLE_INTERNAL_NETWORK_TASK",
			ConfField: "disable_internal_network_task",
			Type:      doc.Boolean,
			Example:   "`true`",
			Default:   "`true`",
			Desc:      "Enable or disable internal IP/service testing",
			DescZh:    "是否允许内网地址/服务的拨测。默认不允许",
		},

		{
			ENVName:   "DISABLED_INTERNAL_NETWORK_CIDR_LIST",
			ConfField: "disabled_internal_network_cidr_list",
			Type:      doc.List,
			Example:   "`[\"192.168.0.0/16\"]`",
			Default:   doc.NoDefaultSet,
			Desc:      "Disable testing on specific internal CIDR IP ranges",
			DescZh:    "禁止拨测的 CIDR 地址列表",
		},

		{
			ENVName:   "ENABLE_DEBUG_API",
			ConfField: doc.NoField,
			Type:      doc.Boolean,
			Example:   "`false`",
			Default:   "`false`",
			Desc:      "Enable the dialtesting debug API (disabled by default)",
			DescZh:    "启用拨测调试接口（默认关闭）",
		},

		{
			ENVName: "ELECTION",
			Type:    doc.Boolean,
			Example: "`false`",
			Default: "`false`",
			Desc:    "Enable election(Default disabled)",
			DescZh:  "开启选举功能（默认禁止）",
		},

		{
			ENVName:   "BROWSER_ENABLED",
			ConfField: "browser.enabled",
			Type:      doc.Boolean,
			Example:   "`false`",
			Default:   "`true`",
			Desc:      "Enable or disable browser dial testing",
			DescZh:    "是否开启浏览器拨测",
		},

		{
			ENVName:   "BROWSER_ENGINE",
			ConfField: "browser.engine",
			Type:      doc.String,
			Example:   "`lightpanda`",
			Default:   "`lightpanda`",
			Desc:      "Browser engine for browser dial testing. Supported value: lightpanda",
			DescZh:    "浏览器拨测使用的引擎，支持 lightpanda",
		},

		{
			ENVName:   "BROWSER_ENGINE_PATH",
			ConfField: "browser.engine_path",
			Type:      doc.String,
			Example:   "`/usr/local/bin/lightpanda`",
			Default:   doc.NoDefaultSet,
			Desc:      "Browser engine executable path for browser dial testing",
			DescZh:    "浏览器拨测使用的浏览器引擎可执行文件路径",
		},

		{
			ENVName:   "BROWSER_CA_CERT_FILE",
			ConfField: "browser.ca_cert_file",
			Type:      doc.String,
			Example:   "`/etc/datakit/certs/internal-ca.pem`",
			Default:   doc.NoDefaultSet,
			Desc:      "CA certificate file trusted by Lightpanda browser dial testing; use PEM format",
			DescZh:    "Lightpanda 浏览器拨测信任的 CA 证书文件，使用 PEM 格式",
		},

		{
			ENVName:   "BROWSER_CA_CERT_DIR",
			ConfField: "browser.ca_cert_dir",
			Type:      doc.String,
			Example:   "`/etc/datakit/certs`",
			Default:   doc.NoDefaultSet,
			Desc:      "Directory containing CA certificates trusted by Lightpanda browser dial testing",
			DescZh:    "Lightpanda 浏览器拨测信任的 CA 证书目录",
		},

		{
			ENVName:   "BROWSER_PROXY_URL",
			ConfField: "browser.proxy_url",
			Type:      doc.String,
			Example:   "`http://proxy.example.com:8080`",
			Default:   doc.NoDefaultSet,
			Desc:      "Default HTTP proxy URL for Lightpanda browser dial testing tasks",
			DescZh:    "Lightpanda 浏览器拨测任务默认 HTTP 代理地址",
		},

		{
			ENVName:   "BROWSER_MAX_CONCURRENCY",
			ConfField: "browser.max_concurrency",
			Type:      doc.Int,
			Example:   "`1`",
			Default:   "`0`",
			Desc:      "Maximum number of browser dial testing tasks running at the same time. 0 means no limit",
			DescZh:    "同一时间最多执行的浏览器拨测任务数，0 表示不限制",
		},
	}

	debugInfos := []*inputs.ENVInfo{
		{
			ENVName:   "MAX_CONCURRENT_RUNS",
			ConfField: doc.NoField,
			Type:      doc.Int,
			Example:   "`100`",
			Default:   "`100`",
			Desc:      "Maximum number of concurrently executing asynchronous debug runs",
			DescZh:    "异步调试任务最大并发执行数",
		},
		{
			ENVName:   "MAX_QUEUED_RUNS",
			ConfField: doc.NoField,
			Type:      doc.Int,
			Example:   "`1000`",
			Default:   "`1000`",
			Desc:      "Maximum total number of preparing and pending asynchronous debug runs",
			DescZh:    "准备中与排队中的异步调试任务总数上限",
		},
		{
			ENVName:   "MAX_QUEUE_WAIT",
			ConfField: doc.NoField,
			Type:      doc.TimeDuration,
			Example:   "`3m`",
			Default:   "`3m`",
			Desc:      "Maximum time an asynchronous debug run may wait in the queue",
			DescZh:    "异步调试任务最大排队等待时间",
		},
		{
			ENVName:   "RESULT_TTL",
			ConfField: doc.NoField,
			Type:      doc.TimeDuration,
			Example:   "`10m`",
			Default:   "`10m`",
			Desc:      "Maximum retention time for terminal asynchronous debug results",
			DescZh:    "异步调试终态结果最长保留时间",
		},
		{
			ENVName:   "MAX_LONG_POLL_WAIT",
			ConfField: doc.NoField,
			Type:      doc.TimeDuration,
			Example:   "`10s`",
			Default:   "`10s`",
			Desc:      "Maximum long-poll wait time; values above 10 seconds are capped",
			DescZh:    "异步调试查询最大长轮询等待时间，超过 10 秒时截断",
		},
		{
			ENVName:   "MAX_RETAINED_TERMINAL_RUNS",
			ConfField: doc.NoField,
			Type:      doc.Int,
			Example:   "`2000`",
			Default:   "`2000`",
			Desc:      "Maximum number of retained terminal asynchronous debug runs; the oldest result is evicted first",
			DescZh:    "异步调试终态任务最大保留数量，超过上限时优先淘汰最早结果",
		},
	}

	infos = doc.SetENVDoc("ENV_INPUT_DIALTESTING_", infos)
	return append(infos, doc.SetENVDoc("ENV_DIALTESTING_DEBUG_", debugInfos)...)
}
