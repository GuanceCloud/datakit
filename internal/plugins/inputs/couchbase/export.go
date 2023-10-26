// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package couchbase

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

//nolint:lll
func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"ram_used_note":    "interestingstats_mem_used: 使用的总内存。<br/>memory_total: 总计内存。<br/>memory_free: 空闲内存。",
			"node_resources":   "节点资源",
			"ram_used":         "内存使用情况",
			"total_ops_note":   "此存储桶每秒的总操作数。",
			"cluster_overview": "集群概览",
			"total_ops":        "总操作数/秒",
			"disk_note":        "disk_write_queue: 此存储桶中等待写入磁盘的项目数。<br/>interestingstats_couch_docs_actual_disk_size: 该存储桶在磁盘上的所有数据服务文件的大小。",
			"disk":             "磁盘数据大小",
			"cpu_used_note":    "systemstats_cpu_utilization_rate: 该服务器上所有可用内核的 CPU 使用百分比。",
			"cpu_used":         "CPU 使用百分比",
			"data_ram_note":    "interestingstats_couch_spatial_data_size: 统计空间数据大小。<br/>ep_mem_low_wat: 空间数据低位线。<br/>ep_mem_high_wat: 空间数据高位线。",
			"data_ram":         "统计空间数据大小",
		}
	case inputs.I18nEn:
		return map[string]string{
			"ram_used_note":    "interestingstats_mem_used: Total memory used in bytes.<br/>memory_total: Memory total.<br/>memory_free: Memory free.",
			"node_resources":   "Node Resources",
			"ram_used":         "Ram Used",
			"total_ops_note":   "Total operations per second (including `XDCR`) to this bucket.",
			"cluster_overview": "Cluster Overview",
			"total_ops":        "Total Operations",
			"disk_note":        "disk_write_queue: Number of items waiting to be written to disk in this bucket.<br/>interestingstats_couch_docs_actual_disk_size: The size of all data service files on disk for this bucket.",
			"disk":             "Disk Write Queue, Data Total Disk Size",
			"cpu_used_note":    "systemstats_cpu_utilization_rate: Percentage of CPU in use across all available cores on this server.",
			"cpu_used":         "CPU Used Percent",
			"data_ram_note":    "interestingstats_couch_spatial_data_size: Interestingstats couch spatial data size.<br/>ep_mem_low_wat: Low water mark for auto-evictions.<br/>ep_mem_high_wat: High water mark for auto-evictions.",
			"data_ram":         "Spatial Data Size",
		}
	default:
		return nil
	}
}

func (*Input) DashboardList() []string {
	return nil
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

func (*Input) MonitorList() []string {
	return nil
}
