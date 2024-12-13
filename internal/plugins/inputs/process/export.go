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
			"sp_title_cpu_usage":  "CPU 使用率",
			"sp_x_title_mem":      "内存",
			"sp_title_mem_usage":  "内存使用率",
			"sp_title_rss":        "RSS",
			"sp_title_open_files": "打开文件数",
			"sp_desc_open_files":  "仅支持 Linux",
			"sp_title_threads":    "线程数",
		}
	case inputs.I18nEn:
		return map[string]string{
			// single process
			"sp_title_cpu_usage":  "CPU usage",
			"sp_x_title_mem":      "Memory",
			"sp_title_mem_usage":  "Memory usage",
			"sp_title_rss":        "RSS",
			"sp_title_open_files": "Open files",
			"sp_desc_open_files":  "Linux only",
			"sp_title_threads":    "Threads",
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
