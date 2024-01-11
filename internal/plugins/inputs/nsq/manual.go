// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nsq

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	nsqTopics = "nsq_topics"
	nsqNodes  = "nsq_nodes"
)

type nsqTopicMeasurement struct{}

//nolint:lll
func (*nsqTopicMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: nsqTopics,
		Type: "metric",
		Desc: "Metrics of all topics in the NSQ cluster",
		Tags: map[string]interface{}{
			"topic":   inputs.NewTagInfo("Topic name"),
			"channel": inputs.NewTagInfo("Channel name"),
		},
		Fields: map[string]interface{}{
			"depth":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of unconsumed messages in the current channel."},
			"backend_depth": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of unconsumed messages exceeding the max-queue-size."},
			"in_flight_count": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of messages during the sending process or client processing " +
				"that have not been sent FIN, REQ (requeued), or timed out."},
			"deferred_count": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of messages that have been requeued and are not yet ready for re-sending."},
			"message_count":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of messages processed in the current channel."},
			"requeue_count":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of messages that have timed out or have been sent REQ by the client."},
			"timeout_count":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of messages that have timed out and are still unprocessed."},
		},
	}
}

type nsqNodesMeasurement struct{}

//nolint:lll
func (*nsqNodesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: nsqNodes,
		Type: "metric",
		Desc: "Metrics of all nodes in the NSQ cluster.",
		Tags: map[string]interface{}{
			"server_host": inputs.NewTagInfo("Service address, that is `host:ip`."),
			"host":        inputs.NewTagInfo("Hostname"),
		},
		Fields: map[string]interface{}{
			"depth":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of unconsumed messages in the current node."},
			"backend_depth": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of unconsumed messages exceeding the max-queue-size."},
			"message_count": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of messages processed by the current node."},
		},
	}
}
