package rabbitmq

import (
	"fmt"
	"net/url"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func getQueues(n *Input) {
	var Queues []Queue
	err := n.requestJSON("/api/queues", &Queues)
	if err != nil {
		l.Error(err.Error())
		n.lastErr = err
		return
	}
	ts := time.Now()
	for _, queue := range Queues {
		tags := map[string]string{
			"url":        n.Url,
			"queue_name": queue.Name,
			"node_name":  queue.Node,
		}
		for k, v := range n.Tags {
			tags[k] = v
		}

		fields := map[string]interface{}{
			"consumers":                    queue.Consumers,
			"consumer_utilisation":         queue.ConsumerUtilisation,
			"memory":                       queue.Memory,
			"head_message_timestamp":       queue.HeadMessageTimestamp,
			"messages":                     queue.Messages,
			"messages_rate":                queue.MessagesDetail.Rate,
			"messages_ready":               queue.MessagesReady,
			"messages_ready_rate":          queue.MessagesReadyDetail.Rate,
			"messages_unacknowledged":      queue.MessagesUnacknowledged,
			"messages_unacknowledged_rate": queue.MessagesUnacknowledgedDetail.Rate,
			"message_ack_count":            queue.MessageStats.Ack,
			"message_ack_rate":             queue.MessageStats.AckDetails.Rate,
			"message_deliver_count":        queue.MessageStats.Deliver,
			"message_deliver_rate":         queue.MessageStats.DeliverDetails.Rate,
			"message_deliver_get_count":    queue.MessageStats.DeliverGet,
			"message_deliver_get_rate":     queue.MessageStats.DeliverGetDetails.Rate,
			"message_publish_count":        queue.MessageStats.Publish,
			"message_publish_rate":         queue.MessageStats.PublishDetails.Rate,
			"message_redeliver_count":      queue.MessageStats.Redeliver,
			"message_redeliver_rate":       queue.MessageStats.RedeliverDetails.Rate,
		}
		bindings, err := n.getBindingCount(queue.Vhost, queue.Name)
		if err != nil {
			l.Errorf("get bindings err:%s", err.Error())
		}
		fields["bindings_count"] = bindings
		metric := &QueueMeasurement{
			name:   QueueMetric,
			tags:   tags,
			fields: fields,
			ts:     ts,
		}
		metricAppend(metric)
	}
}

func (n *Input) getBindingCount(vHost, queueName string) (int, error) {
	var binds []interface{}
	// 此处 vhost 可能是 / 需 encode
	err := n.requestJSON(fmt.Sprintf("/api/queues/%s/%s/bindings", url.QueryEscape(vHost), queueName), &binds)
	if err != nil {
		return 0, err
	}
	return len(binds), nil
}

type QueueMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *QueueMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *QueueMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: QueueMetric,
		Fields: map[string]interface{}{
			"consumers":                    newCountFieldInfo("The ratio of time that a queue's consumers can take new messages"),
			"consumer_utilisation":         newRateFieldInfo("Number of consumers"),
			"head_message_timestamp":       newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.DurationMS, "Timestamp of the head message of the queue Shown as millisecond"),
			"memory":                       newByteFieldInfo("Bytes of memory consumed by the Erlang process associated with the queue, including stack, heap and internal structures"),
			"message":                      newCountFieldInfo("Count of the total messages in the queue"),
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
			"url":        inputs.NewTagInfo("rabbitmq url"),
			"node_name":  inputs.NewTagInfo("rabbitmq node name"),
			"queue_name": inputs.NewTagInfo("rabbitmq queue name"),
		},
	}
}
