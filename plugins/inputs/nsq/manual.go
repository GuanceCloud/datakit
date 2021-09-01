package nsq

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type nsqServerMeasurement struct{}
type nsqTopicMeasurement struct{}
type nsqChannelMeasurement struct{}
type nsqClientMeasurement struct{}

func (*nsqServerMeasurement) LineProto() (*io.Point, error)  { return nil, nil }
func (*nsqTopicMeasurement) LineProto() (*io.Point, error)   { return nil, nil }
func (*nsqChannelMeasurement) LineProto() (*io.Point, error) { return nil, nil }
func (*nsqClientMeasurement) LineProto() (*io.Point, error)  { return nil, nil }

func (*nsqServerMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "nsq_server",
		Desc: "nsq server 简述",
		Tags: map[string]interface{}{
			"server_host":    inputs.NewTagInfo("server 主机连接地址"),
			"server_version": inputs.NewTagInfo("server 版本"),
		},
		Fields: map[string]interface{}{
			"server_count": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "server 数量，如果该 server health 是 `OK` 则为 `1`，反之为 `0`"},
			"topic_count":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "topic 数量"},
		},
	}
}

func (*nsqTopicMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "nsq_topic",
		Desc: "各个 topic 指标",
		Tags: map[string]interface{}{
			"server_host":    inputs.NewTagInfo("server 主机地址"),
			"server_version": inputs.NewTagInfo("server 版本"),
			"topic":          inputs.NewTagInfo("topic 名称"),
		},
		Fields: map[string]interface{}{
			"depth":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "当前 topic 未被消费的消息总数"},
			"backend_depth": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "超出 men-queue-size 的未被消费的消息总数"},
			"message_count": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "nsqd 运行之后这个 topic 产生的消息总数（包括已经没消费和没被消费的）"},
			"channel_count": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "当前 topic 的 channel 数量"},
		},
	}
}

func (*nsqChannelMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "nsq_channel",
		Desc: "连接到该 topic 的所有 channel 指标",
		Tags: map[string]interface{}{
			"server_host":    inputs.NewTagInfo("server 主机地址"),
			"server_version": inputs.NewTagInfo("server 版本"),
			"topic":          inputs.NewTagInfo("topic 名称"),
			"channel":        inputs.NewTagInfo("channel 名称"),
		},
		Fields: map[string]interface{}{
			"depth":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "在当前 channel 中未被消费的消息总数"},
			"backend_depth":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "超出 men-queue-size 的未被消费的消息总数"},
			"in_flight_count": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "发送过程中或者客户端处理过程中的数量，客户端没有发送 FIN、REQ(重新入队列) 和超时的消息数量"},
			"deferred_count":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "重新入队并且还没有准备好重新发送的消息数量"},
			"message_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "当前 channel 处理的消息总数量"},
			"requeue_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "超时或者客户端发送 REQ 的消息数量"},
			"timeout_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "超时未处理的消息数量"},
			"client_count":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "当前 channel 的 client 数量"},
		},
	}
}

func (*nsqClientMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "nsq_client",
		Desc: "连接到该 topic->channel 的所有 client 指标",
		Tags: map[string]interface{}{
			"server_host":    inputs.NewTagInfo("server 主机地址"),
			"server_version": inputs.NewTagInfo("server 版本"),
			"topic":          inputs.NewTagInfo("topic 名称"),
			"channel":        inputs.NewTagInfo("channel 名称"),
			"client_id":      inputs.NewTagInfo("客户端 ID"),
		},
		Fields: map[string]interface{}{
			"client_hostname":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.String, Desc: "客户端主机名称"},
			"client_version":    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.String, Desc: "客户端版本"},
			"client_address":    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.String, Desc: "客户端 remote 地址"},
			"client_user_agent": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.String, Desc: "客户端 user agent"},
			"ready_count":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "客户端能够处理的消息数量"},
			"in_flight_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "客户端正在处理的消息数量"},
			"finish_count":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "客户端处理完成的消息数量"},
			"message_count":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "客户端处理的消息数量"},
			"requeue_count":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "客户端发送重新排队的消息数量，重新排队时因为消息没有被成功处理"},
		},
	}
}
