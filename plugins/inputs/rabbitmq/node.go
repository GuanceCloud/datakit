package rabbitmq

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func getNode(n *Input) {
	var Nodes []Node
	err := n.requestJSON("/api/nodes", &Nodes)
	if err != nil {
		l.Error(err.Error())
		n.lastErr = err
		return
	}
	ts := time.Now()
	for _, node := range Nodes {
		tags := map[string]string{
			"url":       n.URL,
			"node_name": node.Name,
		}
		for k, v := range n.Tags {
			tags[k] = v
		}
		fields := map[string]interface{}{
			"disk_free_alarm": node.DiskFreeAlarm,
			"disk_free":       node.DiskFree,
			"fd_used":         node.FdUsed,
			"mem_alarm":       node.MemAlarm,
			"mem_limit":       node.MemLimit,
			"mem_used":        node.MemUsed,
			"run_queue":       node.RunQueue,
			"running":         node.Running,
			"sockets_used":    node.SocketsUsed,
		}
		metric := &NodeMeasurement{
			name:   NodeMetric,
			tags:   tags,
			fields: fields,
			ts:     ts,
		}
		metricAppend(metric)
	}
}

type NodeMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *NodeMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *NodeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: NodeMetric,
		Fields: map[string]interface{}{
			"disk_free_alarm": newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.UnknownUnit, "Does the node have disk alarm"),
			"disk_free":       newByteFieldInfo("Current free disk space"),
			"fd_used":         newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.UnknownUnit, "Used file descriptors"),
			"mem_alarm":       newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.UnknownUnit, "Does the node have mem alarm"),
			"mem_limit":       newByteFieldInfo("Memory usage high watermark in bytes"),
			"mem_used":        newByteFieldInfo("Memory used in bytes"),
			"run_queue":       newCountFieldInfo("Average number of Erlang processes waiting to run"),
			"running":         newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.UnknownUnit, "Is the node running or not"),
			"sockets_used":    newCountFieldInfo("Number of file descriptors used as sockets"),
		},

		Tags: map[string]interface{}{
			"url":       inputs.NewTagInfo("rabbitmq url"),
			"node_name": inputs.NewTagInfo("rabbitmq node name"),
		},
	}
}
