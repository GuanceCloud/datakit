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
			"cpu_usage":          "CPU 使用占比",
			"mem_usage":          "内存使用占比",
			"open_files":         "打开的文件个数",
			"open_files_note":    "仅支持 Linux,且需开启 enable_open_files 选项",
			"rss":                "常驻内存大小",
			"threads":            "线程数",
			"top_n_cpu_usage":    "Top(n) CPU%",
			"top_n_memory_usage": "Top(n) 内存占用",
			"host_name":          "主机名",
			"process_name":       "进程名",
			"host":               "主机",
		}
	case inputs.I18nEn:
		return map[string]string{
			"cpu_usage":          "Cpu Usage",
			"mem_usage":          "Mem Usage",
			"open_files":         "Open Files",
			"open_files_note":    "Only for Linux.'enable_open_files' need be true",
			"rss":                "Resident Set Size",
			"threads":            "Threads",
			"top_n_cpu_usage":    "Top(n) CPU%",
			"top_n_memory_usage": "Top(n) Memory Usage",
			"host_name":          "Host Name",
			"process_name":       "Process Name",
			"host":               "Host",
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
