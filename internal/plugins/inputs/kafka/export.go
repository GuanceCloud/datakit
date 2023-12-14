// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kafka

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"title":                              "Kafka 监控视图",
			"host_name":                          "主机名",
			"overview":                           "概览",
			"partitions":                         "分区数",
			"controller":                         "Controller 数量",
			"message":                            "消息",
			"time":                               "时间",
			"log":                                "日志",
			"introduction":                       "简介",
			"description":                        "说明",
			"description_content":                `采集器可以从 Kafka 实例中采取很多指标，比如请求、topic等多种指标，并将指标采集到观测云，帮助监控分析 Kafka 各种异常情况。\n\nhttps://docs.guance.com/datakit/kafka/`,
			"total_time":                         "总计时间",
			"message_conversion_time":            "消息转换时间",
			"request_queue_time":                 "请求队列时间",
			"response_queue_time":                "响应队列时间",
			"queue_processing_time":              "队列处理时间",
			"message_conversion":                 "消息转换",
			"fetch_message_conversion_per_sec":   "每秒 Fetch 消息转换",
			"produce_message_conversion_per_sec": "每秒 Produce 消息转换",
			"inflow":                             "入流量",
			"network_incoming_traffic_topic":     "网络入流量-Topic",
			"outflow":                            "出流量",
			"network_outgoing_traffic_topic":     "网络出流量-Topic",
			"network_traffic":                    "网络流量",
			"total_fetch_requests_per_sec":       "每秒 Fetch 请求总计",
			"total_requests":                     "请求总计",
			"total_produce_requests_per_sec":     "每秒 Produce 请求总计",
			"failed_fetch_requests_per_sec":      "每秒 Fetch 请求失败数",
			"failed_request":                     "请求失败",
			"failed_produce_requests_per_sec":    "每秒 Produce 请求失败数",
			"avg_expansions_per_sec":             "平均每秒扩容数",
			"replica_management":                 "副本管理",
			"replica_scaling":                    "副本扩容",
			"avg_shrinks_per_sec":                "平均每秒缩容数",
			"min_failed_ISR_partitions":          "最小失败 ISR 分区数",
			"invalid_replication":                "失效副本",
			"failed_replica_partitions":          "失效副本分区数",
			"invalid_replica_partition":          "失效的副本分区",
			"below_min_ISR_partitions":           "低于最小 ISR 分区数",
			"min_partitions":                     "最小分区数",
			"offline_replication":                "离线副本数",
			"offline_directories":                "离线目录数",
			"broker_partitions":                  "Broker 分区数",
			"total_replications":                 "副本总计",
			"leader_replications":                "Leader 副本数",
			"delete_records":                     "删除记录",
			"heartbeat_detection":                "心跳检测",
			"producer":                           "生产者",
			"event_queue":                        "事件队列",
			"queue_information":                  "队列信息",
			"total_queue":                        "总队列",
			"producer_request_time":              "生产者请求时间",
			"consumer_request_time":              "消费者请求时间",
			"node_request_time":                  "从节点请求时间",
			"performance_monitoring":             "性能监控",
			"election":                           "选举",
			"leader_election":                    "Leader 选举次数",
			"unclear_leader_election":            "Unclean Leader 选举次数",
			"create_chart":                       "新建图表",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"title":                              "Kafka Monitor View",
			"host_name":                          "Host Name",
			"overview":                           "Overview",
			"partitions":                         "Number of Partitions",
			"controller":                         "Number of Controller",
			"message":                            "Message",
			"time":                               "Time",
			"log":                                "Log",
			"introduction":                       "Introduction",
			"description":                        "Description",
			"description_content":                `The collector can gather many metrics from the Kafka instance, such as requests, topics, etc., and collect the metrics to the Guance to help monitor and analyze various abnormal situations of Kafka.\n\nhttps://docs.guance.com/datakit/kafka/`,
			"total_time":                         "Total time",
			"message_conversion_time":            "Message conversion time",
			"request_queue_time":                 "Request queue time",
			"response_queue_time":                "Response queue time",
			"queue_processing_time":              "Queue processing time",
			"message_conversion":                 "Message conversion",
			"fetch_message_conversion_per_sec":   "Fetch message conversion per second",
			"produce_message_conversion_per_sec": "Produce message conversion per second",
			"inflow":                             "Inflow",
			"network_incoming_traffic_topic":     "Network incoming traffic-Topic",
			"outflow":                            "Outflow",
			"network_outgoing_traffic_topic":     "Network outgoing traffic-Topic",
			"network_traffic":                    "Network traffic",
			"total_fetch_requests_per_sec":       "Total fetch requests per second",
			"total_requests":                     "Total requests",
			"total_produce_requests_per_sec":     "Total produce requests per second",
			"failed_fetch_requests_per_sec":      "Number of Failed fetch requests per second",
			"failed_request":                     "Failed request",
			"failed_produce_requests_per_sec":    "Number of Failed produce requests per second",
			"avg_expansions_per_sec":             "Average number of expansions per second",
			"replica_management":                 "Replica Management",
			"replica_scaling":                    "Replica scaling",
			"avg_shrinks_per_sec":                "Average number of shrinks per second",
			"min_failed_ISR_partitions":          "Minimum number of failed ISR partitions",
			"invalid_replication":                "Invalid replication",
			"failed_replica_partitions":          "Number of failed replica partitions",
			"invalid_replica_partition":          "Invalid replica partition",
			"below_min_ISR_partitions":           "Below the minimum number of ISR partitions",
			"min_partitions":                     "Minimum number of partitions",
			"offline_replication":                "Number of offline replication",
			"offline_directories":                "Number of offline directories",
			"broker_partitions":                  "Number of partitions of Broker",
			"total_replications":                 "Total Replications",
			"leader_replications":                "Number of Replications of Leader",
			"delete_records":                     "Delete records",
			"heartbeat_detection":                "Heartbeat detection",
			"producer":                           "Producer",
			"event_queue":                        "Event Queue",
			"queue_information":                  "Queue Information",
			"total_queue":                        "Total Queue",
			"producer_request_time":              "Producer request time",
			"consumer_request_time":              "Consumer request time",
			"node_request_time":                  "Request time from node",
			"performance_monitoring":             "Performance monitoring",
			"election":                           "Election",
			"leader_election":                    "Number of Leader Elections",
			"unclear_leader_election":            "Number of Unclear Leader Elections",
			"create_chart":                       "Create Chart",
		}
	default:
		return nil
	}
}

func (ipt *Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"title":                "Kafka 请求失败数过高",
			"default_monitor_name": "默认",
			"level":                "等级",
			"event":                "事件",
			"event_status":         "事件状态",
			"monitor":              "监控器",
			"alarm_policy":         "告警策略",
			"content":              "内容",
			"content_info":         "Kafka 请求失败数过高",
			"suggestion":           "建议",
			"suggestion_info":      "登录集群查看是否有异常",
		}
	case inputs.I18nEn:
		return map[string]string{
			"title":                "Kafka request failure is too high",
			"default_monitor_name": "Default",
			"level":                "Level",
			"event":                "Event",
			"event_status":         "Event Status",
			"monitor":              "Monitor",
			"alarm_policy":         "Alarm Policy",
			"content":              "Content",
			"content_info":         "Kafka request failure is too high",
			"suggestion":           "Suggestion",
			"suggestion_info":      "Log in to the cluster to see if there are any exceptions",
		}
	default:
		return nil
	}
}
