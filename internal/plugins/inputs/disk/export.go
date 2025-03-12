// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"x_title_bytes":     "磁盘空间",
			"title_free_bytes":  "可用空间",
			"title_total_bytes": "总空间",
			"title_bytes_usage": "百分比",

			"x_title_inode":     "Inode",
			"title_inode_free":  "可用 Inode",
			"title_inode_total": "总量",
			"title_inode_usage": "百分比",

			"x_title_io_time":        "Disk IO time",
			"title_io_time":          "IO 耗时",
			"title_io_wtime":         "Write 耗时",
			"title_io_rtime":         "Read 耗时",
			"title_io_weighted_time": "Weighted 耗时",
			"title_io_iops":          "IOPS",

			"x_title_io_bytes":  "IO 字节",
			"x_io_rbytes":       "Read",
			"x_io_wbytes":       "Write",
			"x_io_merged_read":  "合并 read",
			"x_io_merged_write": "合并 write",
			"x_io_in_progress":  "当前 I/O 次数",
			"x_io_rcount":       "Read 次数",
			"x_io_wcount":       "Write 次数",
		}
	case inputs.I18nEn:
		return map[string]string{
			"x_title_bytes":     "Disk space",
			"title_free_bytes":  "Free",
			"title_total_bytes": "Total",
			"title_bytes_usage": "Usage",

			"x_title_inode":     "Inode",
			"title_inode_free":  "Free",
			"title_inode_total": "Total",
			"title_inode_usage": "Usage",

			"x_title_io_time":        "Disk IO time",
			"title_io_time":          "IO time",
			"title_io_wtime":         "Write time",
			"title_io_rtime":         "Read time",
			"title_io_weighted_time": "Weighted time",
			"title_io_iops":          "IOPS",

			"x_title_io_bytes":  "IO bytes",
			"x_io_rbytes":       "Read",
			"x_io_wbytes":       "Write",
			"x_io_merged_read":  "Merged read",
			"x_io_merged_write": "Merged write",
			"x_io_in_progress":  "I/O in progress",
			"x_io_rcount":       "Read count",
			"x_io_wcount":       "Write count",
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
