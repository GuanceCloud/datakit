// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package process

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			// single process
			"sp_title_cpu_usage":         "CPU 使用率",
			"sp_x_title_mem":             "内存",
			"sp_title_mem_usage":         "内存使用率",
			"sp_title_rss":               "RSS",
			"sp_title_open_files":        "打开文件数",
			"sp_title_threads":           "线程数",
			"sp_title_ctx_switch":        "上下文切换",
			"sp_x_title_proc_read_write": "进程读写",
			"sp_title_read_write_count":  "读写次数",
			"sp_title_read_write_bytes":  "读写字节数",
			"sp_x_title_page_fault":      "Page fault",
			"sp_title_page_fault":        "主进程 Page fault",
			"sp_title_cpage_fault":       "子进程 Page fault",
		}
	case inputs.I18nEn:
		return map[string]string{
			// single process
			"sp_title_cpu_usage":         "CPU usage",
			"sp_x_title_mem":             "Memory",
			"sp_title_mem_usage":         "Memory usage",
			"sp_title_rss":               "RSS",
			"sp_title_open_files":        "Open files",
			"sp_title_threads":           "Threads",
			"sp_title_ctx_switch":        "Context switch",
			"sp_x_title_proc_read_write": "Read/Write",
			"sp_title_read_write_count":  "Read/Write count",
			"sp_title_read_write_bytes":  "Read/Write bytes",
			"sp_x_title_page_fault":      "Page fault",
			"sp_title_page_fault":        "Page fault",
			"sp_title_cpage_fault":       "subprocess page fault",
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
