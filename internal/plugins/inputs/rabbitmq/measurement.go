// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func newRateFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     desc,
	}
}

func newOtherFieldInfo(datatype, ftype, unit, desc string) *inputs.FieldInfo { //nolint:unparam
	return &inputs.FieldInfo{
		DataType: datatype,
		Type:     ftype,
		Unit:     unit,
		Desc:     desc,
	}
}

func newByteFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

type queueMeasurement struct{}

// Point implement MeasurementV2.
func (m *queueMeasurement) Point() *point.Point {
	return nil
}

//nolint:lll
func (m *queueMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: queueMeasurementName,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"consumers":                    newCountFieldInfo("Number of consumers"),
			"consumer_utilization":         newRateFieldInfo("The ratio of time that a queue's consumers can take new messages"),
			"head_message_timestamp":       newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.TimestampMS, "Timestamp of the head message of the queue. Shown as millisecond"),
			"memory":                       newByteFieldInfo("Bytes of memory consumed by the Erlang process associated with the queue, including stack, heap and internal structures"),
			"messages":                     newCountFieldInfo("Count of the total messages in the queue"),
			"messages_rate":                newRateFieldInfo("Count per second of the total messages in the queue"),
			"messages_ready":               newCountFieldInfo("Number of messages ready to be delivered to clients"),
			"messages_ready_rate":          newRateFieldInfo("Number per second of messages ready to be delivered to clients"),
			"messages_unacknowledged":      newCountFieldInfo("Number of messages delivered to clients but not yet acknowledged"),
			"messages_unacknowledged_rate": newRateFieldInfo("Number per second of messages delivered to clients but not yet acknowledged"),
			"message_ack_count":            newCountFieldInfo("Number of messages in queues delivered to clients and acknowledged"),
			"message_ack_rate":             newRateFieldInfo("Number per second of messages delivered to clients and acknowledged"),
			"message_deliver_count":        newCountFieldInfo("Count of messages delivered in acknowledgement mode to consumers"),
			"message_deliver_rate":         newRateFieldInfo("Rate of messages delivered in acknowledgement mode to consumers"),
			"message_deliver_get_count":    newCountFieldInfo("Sum of messages in queues delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get."),
			"message_deliver_get_rate":     newRateFieldInfo("Rate per second of the sum of messages in queues delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get."),
			"message_publish_count":        newCountFieldInfo("Count of messages in queues published"),
			"message_publish_rate":         newRateFieldInfo("Rate per second of messages published"),
			"message_redeliver_count":      newCountFieldInfo("Count of subset of messages in queues in deliver_get which had the redelivered flag set"),
			"message_redeliver_rate":       newRateFieldInfo("Rate per second of subset of messages in deliver_get which had the redelivered flag set"),
			"bindings_count":               newCountFieldInfo("Number of bindings for a specific queue"),
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
		Name: nodeMeasurementName,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"disk_free_alarm": newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.NoUnit, "Does the node have disk alarm"),
			"disk_free":       newByteFieldInfo("Current free disk space"),
			"fd_used":         newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "Used file descriptors"),
			"mem_alarm":       newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.NoUnit, "Does the node have mem alarm"),
			"mem_limit":       newByteFieldInfo("Memory usage high watermark in bytes"),
			"mem_used":        newByteFieldInfo("Memory used in bytes"),
			"run_queue":       newCountFieldInfo("Average number of Erlang processes waiting to run"),
			"running":         newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.NoUnit, "Is the node running or not"),
			"sockets_used":    newCountFieldInfo("Number of file descriptors used as sockets"),

			// See: https://documentation.solarwinds.com/en/success_center/appoptics/content/kb/host_infrastructure/integrations/rabbitmq.htm
			"io_read_avg_time":  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "Average wall time (milliseconds) for each disk read operation in the last statistics interval"),
			"io_write_avg_time": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "Average wall time (milliseconds) for each disk write operation in the last statistics interval"),
			"io_seek_avg_time":  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "Average wall time (milliseconds) for each seek operation in the last statistics interval"),
			"io_sync_avg_time":  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "Average wall time (milliseconds) for each fsync() operation in the last statistics interval"),
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
		Name: overviewMeasurementName,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"object_totals_channels":    newCountFieldInfo("Total number of channels"),
			"object_totals_connections": newCountFieldInfo("Total number of connections"),
			"object_totals_consumers":   newCountFieldInfo("Total number of consumers"),
			"object_totals_queues":      newCountFieldInfo("Total number of queues"),

			"message_ack_count":                    newCountFieldInfo("Number of messages delivered to clients and acknowledged"),
			"message_ack_rate":                     newRateFieldInfo("Rate of messages delivered to clients and acknowledged per second"),
			"message_confirm_count":                newCountFieldInfo("Count of messages confirmed"),
			"message_confirm_rate":                 newRateFieldInfo("Rate of messages confirmed per second"),
			"message_deliver_get_count":            newCountFieldInfo("Sum of messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get"),
			"message_deliver_get_rate":             newRateFieldInfo("Rate per second of the sum of messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get "),
			"message_publish_count":                newCountFieldInfo("Count of messages published"),
			"message_publish_rate":                 newRateFieldInfo("Rate of messages published per second"),
			"message_publish_in_count":             newCountFieldInfo("Count of messages published from channels into this overview"),
			"message_publish_in_rate":              newRateFieldInfo("Rate of messages published from channels into this overview per sec "),
			"message_publish_out_count":            newCountFieldInfo("Count of messages published from this overview into queues"),
			"message_publish_out_rate":             newRateFieldInfo("Rate of messages published from this overview into queues per second"),
			"message_redeliver_count":              newCountFieldInfo("Count of subset of messages in deliver_get which had the redelivered flag set"),
			"message_redeliver_rate":               newRateFieldInfo("Rate of subset of messages in deliver_get which had the redelivered flag set per second"),
			"message_return_unroutable_count_rate": newRateFieldInfo("Rate of messages returned to publisher as unroutable per second"),
			"message_return_unroutable_count":      newCountFieldInfo("Count of messages returned to publisher as unroutable "),

			"queue_totals_messages_count":                newCountFieldInfo("Total number of messages (ready plus unacknowledged)"),
			"queue_totals_messages_rate":                 newRateFieldInfo("Total rate of messages (ready plus unacknowledged)"),
			"queue_totals_messages_ready_count":          newCountFieldInfo("Number of messages ready for delivery "),
			"queue_totals_messages_ready_rate":           newRateFieldInfo("Rate of number of messages ready for delivery"),
			"queue_totals_messages_unacknowledged_count": newCountFieldInfo("Number of unacknowledged messages"),
			"queue_totals_messages_unacknowledged_rate":  newRateFieldInfo("Rate of number of unacknowledged messages"),
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
		Name: exchangeMeasurementName,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"message_ack_count":                    newCountFieldInfo("Number of messages in exchanges delivered to clients and acknowledged"),
			"message_ack_rate":                     newRateFieldInfo("Rate of messages in exchanges delivered to clients and acknowledged per second"),
			"message_confirm_count":                newCountFieldInfo("Count of messages in exchanges confirmed"),
			"message_confirm_rate":                 newRateFieldInfo("Rate of messages in exchanges confirmed per second"),
			"message_deliver_get_count":            newCountFieldInfo("Sum of messages in exchanges delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get"),
			"message_deliver_get_rate":             newRateFieldInfo("Rate per second of the sum of exchange messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get"),
			"message_publish_count":                newCountFieldInfo("Count of messages in exchanges published"),
			"message_publish_rate":                 newRateFieldInfo("Rate of messages in exchanges published per second"),
			"message_publish_in_count":             newCountFieldInfo("Count of messages published from channels into this exchange"),
			"message_publish_in_rate":              newRateFieldInfo("Rate of messages published from channels into this exchange per sec"),
			"message_publish_out_count":            newCountFieldInfo("Count of messages published from this exchange into queues"),
			"message_publish_out_rate":             newRateFieldInfo("Rate of messages published from this exchange into queues per second"),
			"message_redeliver_count":              newCountFieldInfo("Count of subset of messages in exchanges in deliver_get which had the redelivered flag set"),
			"message_redeliver_rate":               newRateFieldInfo("Rate of subset of messages in exchanges in deliver_get which had the redelivered flag set per second"),
			"message_return_unroutable_count_rate": newRateFieldInfo("Rate of messages in exchanges returned to publisher as un-routable per second"),
			"message_return_unroutable_count":      newCountFieldInfo("Count of messages in exchanges returned to publisher as un-routable"),
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
