package rabbitmq

import (
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func getExchange(n *Input) {
	var exchanges []Exchange
	err := n.requestJSON("/api/exchanges", &exchanges)
	if err != nil {
		l.Error(err.Error())
		n.lastErr = err
		return
	}
	ts := time.Now()
	for _, exchange := range exchanges {
		if exchange.Name == "" {
			exchange.Name = "(AMQP default)"
		}

		tags := map[string]string{
			"url":           n.URL,
			"exchange_name": exchange.Name,
			"type":          exchange.Type,
			"vhost":         exchange.Vhost,
			"internal":      strconv.FormatBool(exchange.Internal),
			"durable":       strconv.FormatBool(exchange.Durable),
			"auto_delete":   strconv.FormatBool(exchange.AutoDelete),
		}
		for k, v := range n.Tags {
			tags[k] = v
		}

		fields := map[string]interface{}{
			"message_ack_count":                    exchange.MessageStats.Ack,
			"message_ack_rate":                     exchange.MessageStats.AckDetails.Rate,
			"message_confirm_count":                exchange.MessageStats.Confirm,
			"message_confirm_rate":                 exchange.MessageStats.ConfirmDetail.Rate,
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
		metric := &ExchangeMeasurement{
			name:   ExchangeMetric,
			tags:   tags,
			fields: fields,
			ts:     ts,
		}
		metricAppend(metric)
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

//nolint:lll
func (m *ExchangeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: ExchangeMetric,
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
			"message_return_unroutable_count_rate": newRateFieldInfo("Rate of messages in exchanges returned to publisher as unroutable per second"),
			"message_return_unroutable_count":      newCountFieldInfo("Count of messages in exchanges returned to publisher as unroutable"),
		},

		Tags: map[string]interface{}{
			"url":           inputs.NewTagInfo("rabbitmq url"),
			"exchange_name": inputs.NewTagInfo("rabbitmq exchange name"),
			"type":          inputs.NewTagInfo("rabbitmq exchange type"),
			"vhost":         inputs.NewTagInfo("rabbitmq exchange virtual hosts"),
			"internal":      inputs.NewTagInfo("If set, the exchange may not be used directly by publishers, but only when bound to other exchanges. Internal exchanges are used to construct wiring that is not visible to applications"),
			"durable":       inputs.NewTagInfo("If set when creating a new exchange, the exchange will be marked as durable. Durable exchanges remain active when a server restarts. Non-durable exchanges (transient exchanges) are purged if/when a server restarts."),
			"auto_delete":   inputs.NewTagInfo("If set, the exchange is deleted when all queues have finished using it"),
		},
	}
}
