// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing, metrics and logging.
package skywalking

import (
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	loggingv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/logging/v3"
)

func processLog(plog *loggingv3.LogData) {
	source := plog.Service
	extraTags := make(map[string]string)

	extraTags["endpoint"] = plog.Endpoint
	extraTags["service"] = source
	extraTags["service_instance"] = plog.ServiceInstance
	if plog.Layer != "" {
		extraTags["layer"] = plog.Layer
	}
	for k, v := range tags {
		extraTags[k] = v
	}
	for _, datum := range plog.GetTags().Data {
		switch datum.Key {
		case "level":
			extraTags["status"] = strings.ToLower(datum.Value)
		case "logger":
			extraTags["filename"] = datum.Value
		default:
			extraTags[datum.Key] = datum.Value
		}
	}

	// 不用 pipeline 就是因为这里已经处理好了。
	if ctx := plog.GetTraceContext(); ctx != nil {
		extraTags["trace_id"] = ctx.TraceId
		extraTags["span_id"] = strconv.FormatInt(int64(ctx.SpanId), 10)
		extraTags["trace_segment_id"] = ctx.TraceSegmentId
	}
	line := ""
	switch i := plog.Body.Content.(type) {
	case *loggingv3.LogDataBody_Text:
		line = i.Text.GetText()
		log.Debugf("log string = %s", line)
	case *loggingv3.LogDataBody_Json:
		line = i.Json.GetJson()
		log.Debugf("log string = %s", line)
	case *loggingv3.LogDataBody_Yaml:
		line = i.Yaml.GetYaml()
		log.Debugf("log string = %s", line)
	}
	if line == "" {
		return
	}
	pt, err := point.NewPoint(source, extraTags,
		map[string]interface{}{
			pipeline.FieldMessage: line,
			pipeline.FieldStatus:  pipeline.DefaultStatus,
		}, point.LOpt())
	if err != nil {
		log.Errorf("mew point err=%v", err)
		return
	}

	err = dkio.Feed(source, datakit.Logging, []*point.Point{pt}, nil)
	if err != nil {
		log.Errorf("feed logging err=%v", err)
	}
}
