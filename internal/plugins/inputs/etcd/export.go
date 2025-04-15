// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package etcd

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"title":                                "Etcd 监控视图（非kubenetes集群）",
			"leaders":                              "领导者数量",
			"view":                                 "内置视图",
			"num_unit":                             "个",
			"time_unit":                            "次",
			"cluster_name":                         "集群名称",
			"bytes_received_grpc":                  "接收到 grpc 客户端的总字节数",
			"bytes_sent_grpc":                      "发送到 grpc 客户端的总字节数",
			"grpc_bytes":                           "grpc接收的字节数",
			"leader_change_number":                 "领导者变更次数",
			"consensus_proposals_applied_number":   "已应用的共识提案总数",
			"consensus_proposals_submitted_number": "提交的共识提案总数",
			"consensus_proposals_pending_number":   "当前待处理提案的数量",
			"failed_proposals_number":              "看到的失败提案总数",
			"etcd_view_title":                      "ETCD",
			"overview":                             "概览",
			"node_count":                           "节点数量",
			"health_check_success":                 "健康检查成功次数",
			"health_check_failed":                  "健康检查失败次数",
			"network":                              "网络",
			"server_state":                         "服务器状态",
			"file_description":                     "文件描述符",
			"used_file_description":                "已使用的文件描述符数量",
			"disk":                                 "磁盘",
			"wal_call_fsync_duration":              "由 wal 调用的 fsync 提交延迟",
			"proposals_failed_total":               "提案失败次数",
			"proposals_pending_number":             "待处理的提案数",
			"proposals_committed_total":            "已提交的提案数",
			"proposals_applied_total":              "已应用的提案数",
			"proposals_committed_un_applied_total": "已提交未应用的提案数",
			"memory":                               "内存",
			"storage":                              "存储",
			"max_file_description":                 "ETCD 最大文件描述符数量",
			"disk_backend_commit_duration":         "后端调用提交的延迟",
		}
	case inputs.I18nEn:
		return map[string]string{
			"title":                                "ETCD Monitor View",
			"leaders":                              "Number of Leaders",
			"view":                                 "Built-in View",
			"num_unit":                             "pcs",
			"time_unit":                            "time",
			"cluster_name":                         "Cluster Name",
			"bytes_received_grpc":                  "Total Bytes Received from grpc Client",
			"bytes_sent_grpc":                      "Total Bytes Sent to grpc Client",
			"grpc_bytes":                           "GrpcNumber of Bytes Received by grpc",
			"leader_change_number":                 "Number of Leader Changes",
			"consensus_proposals_applied_number":   "Total number of consensus proposals applied",
			"consensus_proposals_submitted_number": "Total number of consensus proposals submitted",
			"consensus_proposals_pending_number":   "The number of proposals currently pending",
			"failed_proposals_number":              "Total number of failed proposals seen",
			"etcd_view_title":                      "ETCD",
			"overview":                             "Overview",
			"node_count":                           "Node Count",
			"health_check_success":                 "Number of Health Check Successes",
			"health_check_failed":                  "Number of Health Check Failures",
			"network":                              "Network",
			"server_state":                         "Server State",
			"file_description":                     "File Description",
			"used_file_description":                "Number of Used File Descriptors",
			"disk":                                 "Disk",
			"wal_call_fsync_duration":              "Wal Call Fsync Duration",
			"proposals_failed_total":               "Total Number of Proposals Failed",
			"proposals_pending_number":             "Number of Proposals Pending",
			"proposals_committed_total":            "Total Number of Proposals Committed",
			"proposals_applied_total":              "Total Number of Proposals Applied",
			"proposals_committed_un_applied_total": "Total Number of Proposals Committed but Not Applied",
			"memory":                               "Memory",
			"storage":                              "Storage",
			"max_file_description":                 "Maximum File Descriptor Number for ETCD",
			"disk_backend_commit_duration":         "Disk Backend Commit Duration",
		}
	default:
		return nil
	}
}

//nolint:lll
func (ipt *Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"leader_change_title":   "etcd 集群 {{cluster_name_k8s}} 最近1小时leader切换频率过高",
			"leader_change_message": `级别状态: {{ df_status | to_status_human }}\n>集群：{{cluster_name_k8s}}\n>切换次数：{{Result}}}}\n>内容：etcd集群最近1小时leader切换频率过高\n触发时间: {{ date | to_datetime }}`,
			"no_leader_title":       "etcd 集群 {{cluster_name_k8s}} 无 leader",
			"no_leader_message":     `{% if  df_status != 'ok' %}\n级别状态: {{ df_status | to_status_human }}\n>集群：{{cluster_name_k8s}}\n>leader 数量：{{Result}}}}\n>内容：etcd 处于无 leader 状态，无法对外提供服务\n触发时间: {{ date | to_datetime }}\n\n{% else %}\n级别状态: {{ df_status | to_status_human }}\n>集群：{{cluster_name_k8s}}\n>内容：etcd leader 状态已经恢复\n>leader 数量：{{Result}}}}\n恢复时间: {{ date | to_datetime }}\n{% endif %}`,
		}
	case inputs.I18nEn:
		return map[string]string{
			"leader_change_title":   "ETCD Cluster {{cluster_name_k8s}} Leader Change Frequency is High",
			"leader_change_message": `>Level: {{status}}  \n>Cluster: {{cluster_name_k8s}}  \n>Number of Leader Changes: {{ Result }}  \n>Content: ETCD Cluster Leader Change Frequency is High  \n>Trigger Time: {{ date | to_datetime }}`,
			"no_leader_title":       "ETCD Cluster {{cluster_name_k8s}} No Leader",
			"no_leader_message":     `{% if  df_status != 'ok' %}\n>Level: {{status}}  \n>Cluster: {{cluster_name_k8s}}  \n>Number of Leader Changes: {{ Result }}  \n>Content: ETCD Cluster No Leader  \n>Trigger Time: {{ date | to_datetime }}  \n\n{% else %}\n>Level: {{status}}  \n>Cluster: {{cluster_name_k8s}}  \n>Content: ETCD Cluster Leader is Back  \n>Number of Leader Changes: {{ Result }}  \n>Recovery Time: {{ date | to_datetime }}  \n{% endif %}`,
		}
	default:
		return nil
	}
}
