// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package apache

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

var apacheMetricTaggedby = []string{apacheTagURL, apacheTagServerVersion, apacheTagServerMPM}

// Point implement MeasurementV2.
func (m *Measurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Cat:    point.Metric,
		Desc:   "Apache HTTP Server metrics parsed from the machine-readable `mod_status` `server-status?auto` output. Available fields depend on Apache version, MPM, platform, and `ExtendedStatus` configuration.",
		DescZh: "从 Apache HTTP Server 的机器可读 `mod_status` `server-status?auto` 输出解析出的指标。可用字段取决于 Apache 版本、MPM、运行平台以及 `ExtendedStatus` 配置。",
		Fields: map[string]interface{}{
			idleWorkers:          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of idle workers.", Taggedby: apacheMetricTaggedby},
			busyWorkers:          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of workers serving requests.", Taggedby: apacheMetricTaggedby},
			maxWorkers:           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current total number of worker slots in the Apache scoreboard.", Taggedby: apacheMetricTaggedby},
			cpuLoad:              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Current CPU usage percentage reported by Apache `mod_status`; unavailable on Windows and optional depending on configuration.", Taggedby: apacheMetricTaggedby},
			uptime:               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Seconds since the Apache server started.", Taggedby: apacheMetricTaggedby},
			netBytes:             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Cumulative bytes served since the Apache server started.", Taggedby: apacheMetricTaggedby},
			netHits:              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative requests served since the Apache server started.", Taggedby: apacheMetricTaggedby},
			connsTotal:           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current total asynchronous connections; unavailable on Windows.", Taggedby: apacheMetricTaggedby},
			connsAsyncWriting:    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current asynchronous connections in writing state; unavailable on Windows.", Taggedby: apacheMetricTaggedby},
			connsAsyncKeepAlive:  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current asynchronous connections in keep-alive state; unavailable on Windows.", Taggedby: apacheMetricTaggedby},
			connsAsyncClosing:    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current asynchronous connections in closing state; unavailable on Windows.", Taggedby: apacheMetricTaggedby},
			waitingForConnection: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current workers waiting for a connection.", Taggedby: apacheMetricTaggedby},
			startingUp:           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current workers starting up.", Taggedby: apacheMetricTaggedby},
			readingRequest:       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current workers reading requests.", Taggedby: apacheMetricTaggedby},
			sendingReply:         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current workers sending replies.", Taggedby: apacheMetricTaggedby},
			keepAlive:            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current workers in keep-alive state.", Taggedby: apacheMetricTaggedby},
			dnsLookup:            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current workers waiting for DNS lookup.", Taggedby: apacheMetricTaggedby},
			closingConnection:    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current workers closing connections.", Taggedby: apacheMetricTaggedby},
			logging:              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current workers writing to Apache logs.", Taggedby: apacheMetricTaggedby},
			gracefullyFinishing:  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current workers gracefully finishing requests.", Taggedby: apacheMetricTaggedby},
			idleCleanup:          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current workers in idle cleanup.", Taggedby: apacheMetricTaggedby},
			openSlot:             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current open scoreboard slots with no process.", Taggedby: apacheMetricTaggedby},
			disabled:             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current disabled scoreboard slots.", Taggedby: apacheMetricTaggedby},
		},
		Tags: map[string]interface{}{
			apacheTagURL:           inputs.NewTagInfo("Apache server status url."),
			apacheTagServerVersion: inputs.NewTagInfo("Apache server version. Optional."),
			apacheTagServerMPM:     inputs.NewTagInfo("Apache server Multi-Processing Module, `prefork`, `worker` and `event`. Optional."),
			apacheTagHost:          inputs.NewTagInfo("Hostname."),
		},
	}
}
