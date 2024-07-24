// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cockroachdb

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

//nolint:lll
func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return getMapping(true)
	case inputs.I18nEn:
		return getMapping(false)
	default:
		return nil
	}
}

func (ipt *Input) DashboardList() []string {
	return nil
}

func getMapping(zh bool) map[string]string {
	out := make(map[string]string)
	for k, v := range templateNamingMapping {
		if zh {
			out[k] = v[1]
		} else {
			out[k] = v[0]
		}
	}
	return out
}

//nolint:lll
var templateNamingMapping = map[string]([2]string){
	"Overview":                   {"Overview", "概览"},
	"Capacity_Available_by_Node": {"Capacity Available by Node", "节点可用容量"},
	"SQL_Statements_P99":         {"SQL Statements, P99", "SQL语句99分位"},
	"Live_Nodes":                 {"Live Nodes", "活跃节点数"},
	"SQL_Connections":            {"SQL Connections", "SQL连接数"},
	"Available_Capacity":         {"Available Capacity", "可用容量"},
	"Avg_Node_CPU_Usage":         {"Avg. Node CPU Usage (%)", "节点CPU使用率"},
	"Avg_Node_Memory_Usage":      {"Avg. Node Memory Usage (Bytes)", "节点内存使用量"},

	"SQL_Latency":                         {"SQL Latency", "SQL延迟"},
	"Service_Latency_SQL_99th_Percentile": {"Service Latency: SQL, 99th Percentile", "服务延迟：SQL99分位"},
	"Transaction_Latency_99th_Percentile": {"Transaction Latency: 99th Percentile", "事物延迟：99分位"},
	"Connection_Latency_99th_Percentile":  {"Connection Latency: 99th Percentile", "连接延迟：99分位"},

	"Open_SQL_Sessions":    {"Open SQL Sessions", "打开的SQL会话"},
	"SQL_Statements":       {"SQL Statements", "SQL语句"},
	"SQL_Statement_Errors": {"SQL Statement Errors", "SQL语句错误"},

	"Node_CPU_Usage":       {"Node CPU Usage", "CPU使用率"},
	"Node_Memory_Usage":    {"Node Memory Usage", "节点内存使用量"},
	"Cluster_Memory_Usage": {"Cluster Memory Usage", "集群内存使用量"},

	"Storage":                      {"Storage", "存储"},
	"Available_Node_Disk_Capacity": {"Available Node Disk Capacity", "可用节点磁盘容量"},
	"Storage_Capacity":             {"Storage Capacity", "存储容量"},
	"Live_Bytes":                   {"Live Bytes", "活跃的字节数"},

	"Replication_Admission_Control": {"Replication & Admission Control", "副本和准入控制"},
	"Replicas_per_store":            {"Replicas per store", "每个存储的副本数量"},
	"Admission_Delay_Rate":          {"Admission Delay Rate (micros/sec)", "准入延迟率"},
	"Service_Latency":               {"Service Latency)", "服务延迟"},
}
