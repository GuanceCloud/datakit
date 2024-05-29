// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"title":                         "Rabbitmq 监控视图",
			"connection_number":             "连接数",
			"channel_number":                "通道数",
			"queue_number":                  "队列数",
			"overview":                      "概览",
			"introduction":                  "简介",
			"description":                   "说明",
			"description_content":           `采集器可以从 rabbitmq 实例中采取很多指标，比如连接数、队列数、消息总数等多种指标，并将指标采集到观测云，帮助监控分析 rabbitmq 各种异常情况。\n\nhttps://docs.guance.com/datakit/rabbitmq/`,
			"message":                       "消息",
			"queue":                         "队列",
			"node":                          "节点",
			"switch":                        "交换机",
			"host_name":                     "主机名",
			"message_published":             "发布消息",
			"message_published_number":      "消息发布数",
			"message_redelivered":           "重投消息",
			"message_redelivered_number":    "消息重传数",
			"client_confirmation":           "客户端确认",
			"confirmed_client_number":       "客户端确认数",
			"message_delivered":             "投递消息",
			"message_delivered_number":      "消息投递数",
			"message_not_delivered":         "待投递消息",
			"message_not_delivered_number":  "待投递消息数",
			"message_not_routable":          "无路由消息",
			"message_not_routable_number":   "消息不可路由数",
			"consumer_number":               "消费者数",
			"message_rate":                  "消息速率",
			"queue_message":                 "队列消息",
			"queue_message_rate":            "队列消息速率",
			"message_unconfirmed":           "待确认消息",
			"message_unconfirmed_number":    "待确认消息数",
			"message_unconsumed_number":     "待消费消息数",
			"socket_used_file_descriptor":   "Socket已用文件描述符",
			"max_memory_level":              "内存最高水位",
			"free_disk":                     "磁盘余量",
			"memory_used":                   "内存用量",
			"switch_queue_node":             "交换机/队列/节点",
			"client_acknowledgment_message": "客户端确认消息",
			"fd_used":                       "已用FD",
			"fd_socket":                     "Socket用FD",
			"runnable_process":              "待运行进程",
			"alarm_disk_free":               "磁盘剩余告警",
			"alarm_memory":                  "内存告警",
			"running_node":                  "节点运行",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"title":                         "Rabbitmq monitoring view",
			"connection_number":             "Connections",
			"channel_number":                "Number of channels",
			"queue_number":                  "Number of queues",
			"overview":                      "Overview",
			"introduction":                  "Introduction",
			"description":                   "Description",
			"description_content":           `The collector can take many metrics from the rabbitmq instance, such as the number of connections, the number of queues, the total number of messages, and other metrics, and collect the metrics to the observation cloud to help monitor and analyze various abnormal situations of rabbitmq.\n\nhttps://docs.guance.com/datakit/rabbitmq/`,
			"message":                       "Message",
			"queue":                         "Queue",
			"node":                          "Node",
			"switch":                        "Switch",
			"host_name":                     "Host name",
			"message_published":             "Message published",
			"message_published_number":      "Number of messages published",
			"message_redelivered":           "Message redelivered",
			"message_redelivered_number":    "Number of messages redelivered",
			"client_confirmation":           "Client confirmation",
			"confirmed_client_number":       "Number of confirmed clients",
			"message_delivered":             "Message delivery",
			"message_delivered_number":      "Number of message deliveries",
			"message_not_delivered":         "Messages not delivered",
			"message_not_delivered_number":  "Number of messages to be delivered",
			"message_not_routable":          "Message not routable",
			"message_not_routable_number":   "Number of messages not routable",
			"consumer_number":               "Number of consumers",
			"message_rate":                  "Message rate",
			"queue_message":                 "Queue message",
			"queue_message_rate":            "Queue message rate",
			"message_unconfirmed":           "Messages to be confirmed",
			"message_unconfirmed_number":    "Number of messages to be confirmed",
			"message_unconsumed_number":     "Number of messages to be consumed",
			"socket_used_file_descriptor":   "Socket used file descriptor",
			"max_memory_level":              "Maximum memory level",
			"free_disk":                     "Free disk",
			"memory_used":                   "Memory used",
			"switch_queue_node":             "Switch / Queue / Node",
			"client_acknowledgment_message": "Client acknowledgment message",
			"fd_used":                       "FD used",
			"fd_socket":                     "FD for Socket",
			"runnable_process":              "Runnable process",
			"alarm_disk_free":               "Alarm for disk free",
			"alarm_memory":                  "Alarm for memory",
			"running_node":                  "Running node",
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
			"message":              `>等级：{{df_status}}  \n>事件：{{ df_dimension_tags }}\n>监控器：{{ df_monitor_checker_name }}\n>告警策略：{{ df_monitor_name }}\n>事件状态： {{ df_status }}\n>内容：rabbitmq队列消息数过高\n>建议：登录集群查看是否有异常`,
			"title":                "rabbitmq队列消息数过高",
			"default_monitor_name": "默认",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"message":              `>Level: {{df_status}}  \n>Event: {{ df_dimension_tags }}\n>Monitor: {{ df_monitor_checker_name }}\n>Alarm policy: {{ df_monitor_name }}\n>Event status: {{ df_status }}\n>Content: The number of rabbitmq queue messages is too high\n>Suggestion: Log in to the cluster to see if there are any abnormalities`,
			"title":                "The number of rabbitmq queue messages is too high",
			"default_monitor_name": "Default",
		}
	default:
		return nil
	}
}
