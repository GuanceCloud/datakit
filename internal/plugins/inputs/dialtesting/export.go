// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package dialtesting

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

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
