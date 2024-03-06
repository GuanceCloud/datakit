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
			//nolint:lll
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
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
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
