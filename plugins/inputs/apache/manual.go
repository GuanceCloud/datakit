package apache

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *Measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Desc: "采集到的指标，受 Apache 安装环境影响。具体以 `http://<your-apache-server>/server-status?auto` 页面展示的为准。",
		Fields: map[string]interface{}{
			"idle_workers":           newCountFieldInfo("The number of idle workers"),
			"busy_workers":           newCountFieldInfo("The number of workers serving requests."),
			"cpu_load":               newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "The percent of CPU used,windows not support"),
			"uptime":                 newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.DurationSecond, "The amount of time the server has been running"),
			"net_bytes":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.SizeByte, "The total number of bytes served."),
			"net_hits":               newCountFieldInfo("The total number of requests performed"),
			"conns_total":            newCountFieldInfo("The total number of requests performed,windows not support"),
			"conns_async_writing":    newCountFieldInfo("The number of asynchronous writes connections,windows not support"),
			"conns_async_keep_alive": newCountFieldInfo("The number of asynchronous keep alive connections,windows not support"),
			"conns_async_closing":    newCountFieldInfo("The number of asynchronous closing connections,windows not support"),
			WaitingForConnection:     newCountFieldInfo("The number of workers that can immediately process an incoming request"),
			StartingUp:               newCountFieldInfo("The workers that are still starting up and not yet able to handle a request"),
			ReadingRequest:           newCountFieldInfo("The workers reading the incoming request"),
			SendingReply:             newCountFieldInfo("The number of workers sending a reply/response or waiting on a script (like PHP) to finish so they can send a reply"),
			KeepAlive:                newCountFieldInfo("The workers intended for a new request from the same client, because it asked to keep the connection alive"),
			DNSLookup:                newCountFieldInfo("The workers waiting on a DNS lookup"),
			ClosingConnection:        newCountFieldInfo("The amount of workers that are currently closing a connection"),
			Logging:                  newCountFieldInfo("The workers writing something to the Apache logs"),
			GracefullyFinishing:      newCountFieldInfo("The number of workers finishing their request"),
			IdleCleanup:              newCountFieldInfo("These workers were idle and their process is being stopped"),
			OpenSlot:                 newCountFieldInfo("The amount of workers that Apache can still start before hitting the maximum number of workers"),
		},
		Tags: map[string]interface{}{
			"url":            inputs.NewTagInfo("apache server status url"),
			"server_version": inputs.NewTagInfo("apache server version"),
			"server_mpm":     inputs.NewTagInfo("apache server Multi-Processing Module,prefork、worker and event"),
		},
	}
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newOtherFieldInfo(datatype, ftype, unit, desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: datatype,
		Type:     ftype,
		Unit:     unit,
		Desc:     desc,
	}
}
