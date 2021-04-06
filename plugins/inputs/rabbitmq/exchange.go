package rabbitmq

import (
	"time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"strconv"
)

func getExchange(n *Input) {
	exchanges := make([]Exchange, 0)
	err := n.requestJSON("/api/exchanges", &exchanges)
	if err != nil {
		l.Error(err.Error())
		return
	}
	for _, exchange := range exchanges {
		tags := map[string]string{
			"url":      n.Url,
			"exchange": exchange.Name,
			"type":     exchange.Type,
			"vhost":    exchange.Vhost,
			"internal":    strconv.FormatBool(exchange.Internal),
			"durable":     strconv.FormatBool(exchange.Durable),
			"auto_delete": strconv.FormatBool(exchange.AutoDelete),
		}
		fields := map[string]interface{}{
			"message_ack_count":                    exchange.MessageStats.Ack,
			"message_ack_rate":                     exchange.MessageStats.AckDetails.Rate,
			"message_deliver_get_count":            exchange.MessageStats.DeliverGet,
			"message_deliver_get_rate":             exchange.MessageStats.DeliverGetDetails.Rate,
			"message_publish_count":                exchange.MessageStats.Publish,
			"message_publish_rate":                 exchange.MessageStats.PublishDetails.Rate,
			"message_publish_in_count":             exchange.MessageStats.PublishIn,
			"message_publish_in_rate":              exchange.MessageStats.PublishInDetails.Rate,
			"message_publish_out_count":            exchange.MessageStats.PublishOut,
			"message_publish_out_rate":             exchange.MessageStats.PublishOutDetails.Rate,
			"message_redeliver_count":              exchange.MessageStats.Redeliver,
			"message_redeliver_rate":               exchange.MessageStats.RedeliverDetails.Rate,
			"message_return_unroutable_count":      exchange.MessageStats.ReturnUnroutable,
			"message_return_unroutable_count_rate": exchange.MessageStats.ReturnUnroutableDetails.Rate,
		}
	}
}

type ExchangeMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *ExchangeMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *ExchangeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "",
		Fields: map[string]*inputs.FieldInfo{
			"object_totals_channels":    newCountFieldInfo("Total number of channels"),
			"object_totals_connections": newCountFieldInfo("Total number of connections"),
			"object_totals_consumers":   newCountFieldInfo("Total number of consumers"),
			"object_totals_queues":      newCountFieldInfo("Total number of queues"),

			"message_ack_count":                    newCountFieldInfo("Number of messages delivered to clients and acknowledged"),
			"message_ack_rate":                     newRateFieldInfo("Rate of messages delivered to clients and acknowledged per second"),
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
		Tags: map[string]*inputs.TagInfo{
			"url":              inputs.NewTagInfo("rabbitmq url"),
			"rabbitmq_version": inputs.NewTagInfo("rabbitmq version"),
			"cluster_name":     inputs.NewTagInfo("rabbitmq cluster name"),
		},
	}
}
