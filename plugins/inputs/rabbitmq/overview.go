package rabbitmq

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func getOverview(n *Input) {
	overview := &OverviewResponse{}

	err := n.requestJSON("/api/overview", &overview)
	if err != nil {
		l.Errorf(err.Error())
		n.lastErr = err
		return
	}
	ts := time.Now()
	if overview.QueueTotals == nil || overview.ObjectTotals == nil || overview.MessageStats == nil {
		l.Errorf("Wrong answer from rabbitmq. Probably auth issue")
		return
	}
	tags := map[string]string{
		"url":              n.URL,
		"cluster_name":     overview.ClusterName,
		"rabbitmq_version": overview.Version,
	}
	for k, v := range n.Tags {
		tags[k] = v
	}
	fields := map[string]interface{}{
		"object_totals_channels":    overview.ObjectTotals.Channels,
		"object_totals_connections": overview.ObjectTotals.Connections,
		"object_totals_consumers":   overview.ObjectTotals.Consumers,
		"object_totals_queues":      overview.ObjectTotals.Queues,

		"message_ack_count":                    overview.MessageStats.Ack,
		"message_ack_rate":                     overview.MessageStats.AckDetails.Rate,
		"message_confirm_count":                overview.MessageStats.Confirm,
		"message_confirm_rate":                 overview.MessageStats.ConfirmDetail.Rate,
		"message_deliver_get_count":            overview.MessageStats.DeliverGet,
		"message_deliver_get_rate":             overview.MessageStats.DeliverGetDetails.Rate,
		"message_publish_count":                overview.MessageStats.Publish,
		"message_publish_rate":                 overview.MessageStats.PublishDetails.Rate,
		"message_publish_in_count":             overview.MessageStats.PublishIn,
		"message_publish_in_rate":              overview.MessageStats.PublishInDetails.Rate,
		"message_publish_out_count":            overview.MessageStats.PublishOut,
		"message_publish_out_rate":             overview.MessageStats.PublishOutDetails.Rate,
		"message_redeliver_count":              overview.MessageStats.Redeliver,
		"message_redeliver_rate":               overview.MessageStats.RedeliverDetails.Rate,
		"message_return_unroutable_count":      overview.MessageStats.ReturnUnroutable,
		"message_return_unroutable_count_rate": overview.MessageStats.ReturnUnroutableDetails.Rate,

		"queue_totals_messages_count":                overview.QueueTotals.Messages,
		"queue_totals_messages_rate":                 overview.QueueTotals.MessagesDetail.Rate,
		"queue_totals_messages_ready_count":          overview.QueueTotals.MessagesReady,
		"queue_totals_messages_ready_rate":           overview.QueueTotals.MessagesReadyDetail.Rate,
		"queue_totals_messages_unacknowledged_count": overview.QueueTotals.MessagesUnacknowledged,
		"queue_totals_messages_unacknowledged_rate":  overview.QueueTotals.MessagesUnacknowledgedDetail.Rate,
	}
	metric := &OverviewMeasurement{
		name:   OverviewMetric,
		tags:   tags,
		fields: fields,
		ts:     ts,
	}
	metricAppend(metric)
}

type OverviewMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *OverviewMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *OverviewMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: OverviewMetric,
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
			"url":              inputs.NewTagInfo("rabbitmq url"),
			"rabbitmq_version": inputs.NewTagInfo("rabbitmq version"),
			"cluster_name":     inputs.NewTagInfo("rabbitmq cluster name"),
		},
	}
}
