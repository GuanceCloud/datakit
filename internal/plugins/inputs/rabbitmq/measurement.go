// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	queueFieldTags    = []string{"url", "queue_name", "node_name", "vhost"}
	nodeFieldTags     = []string{"url", "node_name"}
	overviewFieldTags = []string{"url", "rabbitmq_version", "cluster_name"}
	exchangeFieldTags = []string{"url", "exchange_name", "type", "vhost", "internal", "durable", "auto_delete"}
)

type queueMeasurement struct{}

// Point implement MeasurementV2.
func (m *queueMeasurement) Point() *point.Point {
	return nil
}

//nolint:lll
func (m *queueMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   queueMeasurementName,
		Cat:    point.Metric,
		Desc:   "Per-queue RabbitMQ metrics collected from the management API `/api/queues`, including queue depth, consumer capacity, message rates, memory, and binding count.",
		DescZh: "通过 RabbitMQ Management API `/api/queues` 采集的队列维度指标，包括队列消息堆积、消费者处理能力、消息速率、内存占用和绑定数量。",
		Fields: map[string]interface{}{
			"consumers":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of consumers", Taggedby: queueFieldTags},
			"consumer_utilization":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The ratio of time that a queue's consumers can take new messages", Taggedby: queueFieldTags},
			"head_message_timestamp":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS, Desc: "Timestamp of the head message of the queue. Shown as millisecond", Taggedby: queueFieldTags},
			"memory":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes of memory consumed by the Erlang process associated with the queue, including stack, heap and internal structures", Taggedby: queueFieldTags},
			"messages":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of the total messages in the queue", Taggedby: queueFieldTags},
			"messages_rate":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Count per second of the total messages in the queue", Taggedby: queueFieldTags},
			"messages_ready":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of messages ready to be delivered to clients", Taggedby: queueFieldTags},
			"messages_ready_rate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Number per second of messages ready to be delivered to clients", Taggedby: queueFieldTags},
			"messages_unacknowledged":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of messages delivered to clients but not yet acknowledged", Taggedby: queueFieldTags},
			"messages_unacknowledged_rate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Number per second of messages delivered to clients but not yet acknowledged", Taggedby: queueFieldTags},
			"message_ack_count":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of messages in queues delivered to clients and acknowledged", Taggedby: queueFieldTags},
			"message_ack_rate":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Number per second of messages delivered to clients and acknowledged", Taggedby: queueFieldTags},
			"message_deliver_count":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages delivered in acknowledgement mode to consumers", Taggedby: queueFieldTags},
			"message_deliver_rate":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages delivered in acknowledgement mode to consumers", Taggedby: queueFieldTags},
			"message_deliver_get_count":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Sum of messages in queues delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get.", Taggedby: queueFieldTags},
			"message_deliver_get_rate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate per second of the sum of messages in queues delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get.", Taggedby: queueFieldTags},
			"message_publish_count":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages in queues published", Taggedby: queueFieldTags},
			"message_publish_rate":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate per second of messages published", Taggedby: queueFieldTags},
			"message_redeliver_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of subset of messages in queues in deliver_get which had the redelivered flag set", Taggedby: queueFieldTags},
			"message_redeliver_rate":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate per second of subset of messages in deliver_get which had the redelivered flag set", Taggedby: queueFieldTags},
			"bindings_count":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bindings for a specific queue", Taggedby: queueFieldTags},
		},

		Tags: map[string]interface{}{
			"url":          inputs.NewTagInfo("RabbitMQ host URL"),
			"node_name":    inputs.NewTagInfo("RabbitMQ node name"),
			"queue_name":   inputs.NewTagInfo("RabbitMQ queue name"),
			"cluster_name": inputs.NewTagInfo("RabbitMQ cluster name"),
			"host":         inputs.NewTagInfo("Hostname of RabbitMQ running on."),
			"vhost":        inputs.NewTagInfo("RabbitMQ queue virtual hosts"),
		},
	}
}

type nodeMeasurement struct{}

// Point implement MeasurementV2.
func (m *nodeMeasurement) Point() *point.Point {
	return nil
}

//nolint:lll
func (m *nodeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   nodeMeasurementName,
		Cat:    point.Metric,
		Desc:   "Per-node RabbitMQ runtime and resource metrics collected from the management API `/api/nodes`, including disk, memory, file descriptor, socket, run queue, and disk I/O latency state.",
		DescZh: "通过 RabbitMQ Management API `/api/nodes` 采集的节点维度运行时与资源指标，包括磁盘、内存、文件描述符、socket、运行队列和磁盘 I/O 延迟状态。",
		Fields: map[string]interface{}{
			"disk_free_alarm": &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Bool, Unit: inputs.Bool, Desc: "Does the node have disk alarm", Taggedby: nodeFieldTags},
			"disk_free":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Current free disk space", Taggedby: nodeFieldTags},
			"fd_used":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Used file descriptors", Taggedby: nodeFieldTags},
			"mem_alarm":       &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Bool, Unit: inputs.Bool, Desc: "Does the node have mem alarm", Taggedby: nodeFieldTags},
			"mem_limit":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Memory usage high watermark in bytes", Taggedby: nodeFieldTags},
			"mem_used":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Memory used in bytes", Taggedby: nodeFieldTags},
			"run_queue":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Average number of Erlang processes waiting to run", Taggedby: nodeFieldTags},
			"running":         &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Bool, Unit: inputs.Bool, Desc: "Is the node running or not", Taggedby: nodeFieldTags},
			"sockets_used":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of file descriptors used as sockets", Taggedby: nodeFieldTags},

			// See: https://documentation.solarwinds.com/en/success_center/appoptics/content/kb/host_infrastructure/integrations/rabbitmq.htm
			"io_read_avg_time":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average wall time (milliseconds) for each disk read operation in the last statistics interval", Taggedby: nodeFieldTags},
			"io_write_avg_time": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average wall time (milliseconds) for each disk write operation in the last statistics interval", Taggedby: nodeFieldTags},
			"io_seek_avg_time":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average wall time (milliseconds) for each seek operation in the last statistics interval", Taggedby: nodeFieldTags},
			"io_sync_avg_time":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average wall time (milliseconds) for each fsync() operation in the last statistics interval", Taggedby: nodeFieldTags},
		},

		Tags: map[string]interface{}{
			"url":          inputs.NewTagInfo("RabbitMQ url"),
			"node_name":    inputs.NewTagInfo("RabbitMQ node name"),
			"cluster_name": inputs.NewTagInfo("RabbitMQ cluster name"),
			"host":         inputs.NewTagInfo("Hostname of RabbitMQ running on."),
		},
	}
}

type overviewMeasurement struct{}

// Point implement MeasurementV2.
func (m *overviewMeasurement) Point() *point.Point {
	return nil
}

//nolint:lll
func (m *overviewMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   overviewMeasurementName,
		Cat:    point.Metric,
		Desc:   "Cluster-level RabbitMQ overview metrics collected from the management API `/api/overview`, including object totals, message statistics, and queue totals.",
		DescZh: "通过 RabbitMQ Management API `/api/overview` 采集的集群概览指标，包括对象总数、消息统计和队列总量。",
		Fields: map[string]interface{}{
			"object_totals_channels":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of channels", Taggedby: overviewFieldTags},
			"object_totals_connections": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of connections", Taggedby: overviewFieldTags},
			"object_totals_consumers":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of consumers", Taggedby: overviewFieldTags},
			"object_totals_queues":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of queues", Taggedby: overviewFieldTags},

			"message_ack_count":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of messages delivered to clients and acknowledged", Taggedby: overviewFieldTags},
			"message_ack_rate":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages delivered to clients and acknowledged per second", Taggedby: overviewFieldTags},
			"message_confirm_count":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages confirmed", Taggedby: overviewFieldTags},
			"message_confirm_rate":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages confirmed per second", Taggedby: overviewFieldTags},
			"message_deliver_get_count":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Sum of messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get", Taggedby: overviewFieldTags},
			"message_deliver_get_rate":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate per second of the sum of messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get ", Taggedby: overviewFieldTags},
			"message_publish_count":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages published", Taggedby: overviewFieldTags},
			"message_publish_rate":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages published per second", Taggedby: overviewFieldTags},
			"message_publish_in_count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages published from channels into this overview", Taggedby: overviewFieldTags},
			"message_publish_in_rate":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages published from channels into this overview per sec ", Taggedby: overviewFieldTags},
			"message_publish_out_count":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages published from this overview into queues", Taggedby: overviewFieldTags},
			"message_publish_out_rate":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages published from this overview into queues per second", Taggedby: overviewFieldTags},
			"message_redeliver_count":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of subset of messages in deliver_get which had the redelivered flag set", Taggedby: overviewFieldTags},
			"message_redeliver_rate":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of subset of messages in deliver_get which had the redelivered flag set per second", Taggedby: overviewFieldTags},
			"message_return_unroutable_count_rate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages returned to publisher as unroutable per second", Taggedby: overviewFieldTags},
			"message_return_unroutable_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages returned to publisher as unroutable ", Taggedby: overviewFieldTags},

			"queue_totals_messages_count":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of messages (ready plus unacknowledged)", Taggedby: overviewFieldTags},
			"queue_totals_messages_rate":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Total rate of messages (ready plus unacknowledged)", Taggedby: overviewFieldTags},
			"queue_totals_messages_ready_count":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of messages ready for delivery ", Taggedby: overviewFieldTags},
			"queue_totals_messages_ready_rate":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of number of messages ready for delivery", Taggedby: overviewFieldTags},
			"queue_totals_messages_unacknowledged_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of unacknowledged messages", Taggedby: overviewFieldTags},
			"queue_totals_messages_unacknowledged_rate":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of number of unacknowledged messages", Taggedby: overviewFieldTags},
		},
		Tags: map[string]interface{}{
			"url":              inputs.NewTagInfo("RabbitMQ url"),
			"rabbitmq_version": inputs.NewTagInfo("RabbitMQ version"),
			"cluster_name":     inputs.NewTagInfo("RabbitMQ cluster name"),
			"host":             inputs.NewTagInfo("Hostname of RabbitMQ running on."),
		},
	}
}

type exchangeMeasurement struct{}

// Point implement MeasurementV2.
func (m *exchangeMeasurement) Point() *point.Point {
	return nil
}

//nolint:lll
func (m *exchangeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   exchangeMeasurementName,
		Cat:    point.Metric,
		Desc:   "Per-exchange RabbitMQ message metrics collected from the management API `/api/exchanges`, including publish, confirm, deliver/get, redeliver, and unroutable-return activity.",
		DescZh: "通过 RabbitMQ Management API `/api/exchanges` 采集的交换机维度消息指标，包括发布、确认、投递/获取、重投递和不可路由退回等活动。",
		Fields: map[string]interface{}{
			"message_ack_count":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of messages in exchanges delivered to clients and acknowledged", Taggedby: exchangeFieldTags},
			"message_ack_rate":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages in exchanges delivered to clients and acknowledged per second", Taggedby: exchangeFieldTags},
			"message_confirm_count":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages in exchanges confirmed", Taggedby: exchangeFieldTags},
			"message_confirm_rate":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages in exchanges confirmed per second", Taggedby: exchangeFieldTags},
			"message_deliver_get_count":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Sum of messages in exchanges delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get", Taggedby: exchangeFieldTags},
			"message_deliver_get_rate":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate per second of the sum of exchange messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get", Taggedby: exchangeFieldTags},
			"message_publish_count":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages in exchanges published", Taggedby: exchangeFieldTags},
			"message_publish_rate":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages in exchanges published per second", Taggedby: exchangeFieldTags},
			"message_publish_in_count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages published from channels into this exchange", Taggedby: exchangeFieldTags},
			"message_publish_in_rate":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages published from channels into this exchange per sec", Taggedby: exchangeFieldTags},
			"message_publish_out_count":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages published from this exchange into queues", Taggedby: exchangeFieldTags},
			"message_publish_out_rate":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages published from this exchange into queues per second", Taggedby: exchangeFieldTags},
			"message_redeliver_count":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of subset of messages in exchanges in deliver_get which had the redelivered flag set", Taggedby: exchangeFieldTags},
			"message_redeliver_rate":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of subset of messages in exchanges in deliver_get which had the redelivered flag set per second", Taggedby: exchangeFieldTags},
			"message_return_unroutable_count_rate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.NCount, Desc: "Rate of messages in exchanges returned to publisher as un-routable per second", Taggedby: exchangeFieldTags},
			"message_return_unroutable_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of messages in exchanges returned to publisher as un-routable", Taggedby: exchangeFieldTags},
		},

		Tags: map[string]interface{}{
			"url":           inputs.NewTagInfo("RabbitMQ host URL"),
			"exchange_name": inputs.NewTagInfo("RabbitMQ exchange name"),
			"type":          inputs.NewTagInfo("RabbitMQ exchange type"),
			"vhost":         inputs.NewTagInfo("RabbitMQ exchange virtual hosts"),
			"internal":      inputs.NewTagInfo("If set, the exchange may not be used directly by publishers, but only when bound to other exchanges. Internal exchanges are used to construct wiring that is not visible to applications"),
			"durable":       inputs.NewTagInfo("If set when creating a new exchange, the exchange will be marked as durable. Durable exchanges remain active when a server restarts. Non-durable exchanges (transient exchanges) are purged if/when a server restarts."),
			"auto_delete":   inputs.NewTagInfo("If set, the exchange is deleted when all queues have finished using it"),
			"host":          inputs.NewTagInfo("Hostname of RabbitMQ running on."),
			"cluster_name":  inputs.NewTagInfo("RabbitMQ cluster name"),
		},
	}
}
