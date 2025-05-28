// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostdir

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
		//nolint:lll
		return map[string]string{
			"monitor_name":                                `主机检测库`,
			"host_mem_usage_title":                        `主机 {{ host }} 内存使用率过高`,
			"host_mem_usage_message":                      `>等级：{{status}}  \n>主机：{{host}}  \n>内容：内存使用率为 {{ Result |  to_fixed(2) }}%  \n>建议：基础设施-进程-选择主机-内存使用率 (排序) 查看是否为异常导致`,
			"host_cpu_usage_title":                        `主机 {{ host }} CPU 使用率过高`,
			"host_cpu_usage_message":                      `>等级：{{status}}  \n>主机：{{host}}  \n>内容：系统 CPU 使用率为 {{ Result |  to_fixed(2) }}%  \n>建议：基础设施-进程-选择主机-CPU 使用率 (排序) 查看是否为异常导致`,
			"host_mem_free_title":                         `主机 {{ host }} 内存剩余不足`,
			"host_mem_free_message":                       `>等级：{{status}}  \n>主机：{{host}}  \n>内容：内存为 {{ Result |  to_fixed(2) }}M  \n>建议：基础设施-进程-选择主机-内存使用率 (排序) 查看是否为异常导致`,
			"host_cpu_load_title":                         `主机 {{ host }} CPU 负载过高`,
			"host_cpu_load_message":                       `>等级：{{status}}  \n>主机：{{host}}  \n>内容：系统 CPU 平均负载为 {{ Result |  to_fixed(2) }}  \n>建议：平均负载过高，可能是 CPU 密集型应用进程导致；如果同时 CPU 使用率不高，可能是 I/O 密集型应用进程导致`,
			"host_disk_free_title":                        `主机 {{ host }} 磁盘剩余空间过低`,
			"host_disk_free_message":                      `>等级：{{status}}  \n>主机：{{host}}  \n>内容：磁盘 {{device}} 剩余空间为 {{ Result |  to_fixed(2) }}%  \n>建议：磁盘空间即将耗尽，导致无法正常写入数据，请及时清理不必要的文件`,
			"host_swap_usage_title":                       `主机 {{ host }} 内存 Swap 使用率过高`,
			"host_swap_usage_message":                     `>等级：{{status}}  \n>主机：{{host}}  \n>内容：内存 Swap 使用率为 {{ Result |  to_fixed(2) }}%  \n>建议：内存 Swap 耗尽可能导致宕机风险，请查看内存使用率高的进程/应用是否为异常导致`,
			"host_inode_free_title":                       `主机 {{ host }} 文件系统剩余 inode 过低`,
			"host_inode_free_message":                     `>等级：{{status}}  \n>主机：{{host}}  \n>内容：文件系统剩余 inode 为 {{ (100 - Result) |  to_fixed(2) }}%  \n>建议：文件系统 inode 耗尽将无法写入数据，请查看是否有大量小文件占用 inode`,
			"host_iowait_title":                           `主机 {{ host }} CPU IOwait 过高`,
			"host_iowait_message":                         `>等级：{{status}}  \n>主机：{{host}}  \n>内容：系统 CPU IOwait 为 {{ Result |  to_fixed(2) }}%  \n>建议：等待 I/O 的 CPU 时间过长，可能存在频繁写入或 I/O 瓶颈`,

		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"monitor_name":                                `Host Detection Library`,
			"host_mem_usage_title":                        `Host {{ host }} Memory Usage is High`,
			"host_mem_usage_message":                      `>Level: {{status}}  \n>Host: {{host}}  \n>Content: Memory Usage is {{ Result |  to_fixed(2) }}%  \n>Suggest: Infrastructure-Process-Select Host-Memory Usage (Sort) to check if it is caused byabnormal`,
			"host_cpu_usage_title":                        `Host {{ host }} CPU Usage is High`,
			"host_cpu_usage_message":                      `>Level: {{status}}  \n>Host: {{host}}  \n>Content: System CPU Usage is {{ Result |  to_fixed(2) }}%  \n>Suggest: Infrastructure-Process-Select Host-CPU Usage (Sort) to check if it iscaused byabnormal`,
			"host_mem_free_title":                         `Host {{ host }} Memory Free is Low`,
			"host_mem_free_message":                       `>Level: {{status}}  \n>Host: {{host}}  \n>Content: Memory is {{ Result |  to_fixed(2) }}M  \n>Suggest: Infrastructure-Process-Select Host-Memory Usage (Sort) to check if it iscaused byabnormal`,
			"host_cpu_load_title":                         `Host {{ host }} CPU Load is High`,
			"host_cpu_load_message":                        `>Level: {{status}}  \n>Host: {{host}}  \n>Content: System CPU Load is {{ Result |  to_fixed(2) }}  \n>Suggest: Average load is too high, which may be caused by CPU-intensive application processes; If CPU usage is low, it may be caused by IO-intensive application processes`,
			"host_disk_free_title":                        `Host {{ host }} Disk Free Space is Low`,
			"host_disk_free_message":                      `>Level: {{status}}  \n>Host: {{host}}  \n>Content: Disk {{device}} Free Space is {{ Result |  to_fixed(2) }}%  \n>Suggest: Disk space is running out, which will cause data to be unable to be written normally, please clear unnecessary files as soon as possible`,
			"host_swap_usage_title":                       `Host {{ host }} Memory Swap Usage is High`,
			"host_swap_usage_message":                     `>Level: {{status}}  \n>Host: {{host}}  \n>Content: Memory Swap Usage is {{ Result |  to_fixed(2) }}%  \n>Suggest: Memory Swap usage is too high, which may cause a risk of system crash, please check the memory usage of the process/application that is abnormal`,
			"host_inode_free_title":                       `Host {{ host }} File System Inode Free is Low`,
			"host_inode_free_message":                     `>Level: {{status}}  \n>Host: {{host}}  \n>Content: File System Inode Free is {{ (100 - Result) |  to_fixed(2) }}%  \n>Suggest: File System Inode is running out, which will cause data to be unable to be written normally, please check if there are many small files occupying inodes`,
			"host_iowait_title":                            `The host iowait of Host {{ host }} is too high`,
			"host_iowait_message":                          `>Level: {{status}}  \n>Host: {{host}}  \n>Content: Host iowait is {{ Result }}%.  \n>Suggest: The host iowait is too high, which may affect the performance of the host.`,


		}
	default:
		return nil
	}
}
