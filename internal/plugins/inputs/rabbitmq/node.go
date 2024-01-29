// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
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
		if n.host != "" {
			tags["host"] = n.host
		}
		for k, v := range n.Tags {
			tags[k] = v
		}

		if n.Election {
			tags = inputs.MergeTags(n.Tagger.ElectionTags(), tags, n.URL)
		} else {
			tags = inputs.MergeTags(n.Tagger.HostTags(), tags, n.URL)
		}

		fields := map[string]interface{}{
			"disk_free_alarm":   node.DiskFreeAlarm,
			"disk_free":         node.DiskFree,
			"fd_used":           node.FdUsed,
			"mem_alarm":         node.MemAlarm,
			"mem_limit":         node.MemLimit,
			"mem_used":          node.MemUsed,
			"run_queue":         node.RunQueue,
			"running":           node.Running,
			"sockets_used":      node.SocketsUsed,
			"io_write_avg_time": node.IoWriteAvgTime,
			"io_read_avg_time":  node.IoReadAvgTime,
			"io_sync_avg_time":  node.IoSyncAvgTime,
			"io_seek_avg_time":  node.IoSeekAvgTime,
		}
		metric := &NodeMeasurement{
			name:   NodeMetric,
			tags:   tags,
			fields: fields,
			ts:     ts,
		}
		n.metricAppend(metric.Point())
	}
}

type NodeMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *NodeMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
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

			// See: https://documentation.solarwinds.com/en/success_center/appoptics/content/kb/host_infrastructure/integrations/rabbitmq.htm
			"io_read_avg_time":  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "Average wall time (milliseconds) for each disk read operation in the last statistics interval"),
			"io_write_avg_time": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "Average wall time (milliseconds) for each disk write operation in the last statistics interval"),
			"io_seek_avg_time":  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "Average wall time (milliseconds) for each seek operation in the last statistics interval"),
			"io_sync_avg_time":  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "Average wall time (milliseconds) for each fsync() operation in the last statistics interval"),
		},

		Tags: map[string]interface{}{
			"url":       inputs.NewTagInfo("RabbitMQ url"),
			"node_name": inputs.NewTagInfo("RabbitMQ node name"),
			"host":      inputs.NewTagInfo("Hostname of RabbitMQ running on."),
		},
	}
}
