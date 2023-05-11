// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package apache

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	ipt    *Input
}

// Point implement MeasurementV2.
func (m *Measurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.ipt.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *Measurement) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(m.name, m.tags, m.fields, point.MOptElectionV2(m.election))
	return nil, fmt.Errorf("not implement")
}

//nolint:lll
func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Desc: "采集到的指标，受 Apache 安装环境影响。具体以 `http://<your-apache-server>/server-status?auto` 页面展示的为准。",
		Fields: map[string]interface{}{
			"idle_workers":           newCountFieldInfo("The number of idle workers"),
			"busy_workers":           newCountFieldInfo("The number of workers serving requests."),
			"cpu_load":               newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "The percent of CPU used,windows not support. Optional."),
			"uptime":                 newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.DurationSecond, "The amount of time the server has been running"),
			"net_bytes":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.SizeByte, "The total number of bytes served."),
			"net_hits":               newCountFieldInfo("The total number of requests performed"),
			"conns_total":            newCountFieldInfo("The total number of requests performed,windows not support"),
			"conns_async_writing":    newCountFieldInfo("The number of asynchronous writes connections,windows not support"),
			"conns_async_keep_alive": newCountFieldInfo("The number of asynchronous keep alive connections,windows not support"),
			"conns_async_closing":    newCountFieldInfo("The number of asynchronous closing connections,windows not support"),
			waitingForConnection:     newCountFieldInfo("The number of workers that can immediately process an incoming request"),
			startingUp:               newCountFieldInfo("The workers that are still starting up and not yet able to handle a request"),
			readingRequest:           newCountFieldInfo("The workers reading the incoming request"),
			sendingReply:             newCountFieldInfo("The number of workers sending a reply/response or waiting on a script (like PHP) to finish so they can send a reply"),
			keepAlive:                newCountFieldInfo("The workers intended for a new request from the same client, because it asked to keep the connection alive"),
			dnsLookup:                newCountFieldInfo("The workers waiting on a DNS lookup"),
			closingConnection:        newCountFieldInfo("The amount of workers that are currently closing a connection"),
			logging:                  newCountFieldInfo("The workers writing something to the Apache logs"),
			gracefullyFinishing:      newCountFieldInfo("The number of workers finishing their request"),
			idleCleanup:              newCountFieldInfo("These workers were idle and their process is being stopped"),
			openSlot:                 newCountFieldInfo("The amount of workers that Apache can still start before hitting the maximum number of workers"),
		},
		Tags: map[string]interface{}{
			"url":            inputs.NewTagInfo("Apache server status url."),
			"server_version": inputs.NewTagInfo("Apache server version. Optional."),
			"server_mpm":     inputs.NewTagInfo("Apache server Multi-Processing Module, `prefork`, `worker` and `event`. Optional."),
			"host":           inputs.NewTagInfo("Hostname of the DataKit."),
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
