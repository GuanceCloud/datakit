// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package skywalking

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	loggingv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/logging/v3"
)

func processLogV3(plog *loggingv3.LogData) (*point.Point, error) {
	extraTags := make(map[string]string)
	extraTags["endpoint"] = plog.Endpoint
	extraTags["service"] = plog.Service
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
	case *loggingv3.LogDataBody_Json:
		line = i.Json.GetJson()
	case *loggingv3.LogDataBody_Yaml:
		line = i.Yaml.GetYaml()
	}
	if line == "" {
		return nil, errors.New("unknown log data body")
	}

	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(time.Now()))
	return point.NewPointV2(plog.Service,
			append(point.NewTags(extraTags), point.NewKVs(map[string]interface{}{
				constants.FieldMessage: line,
			})...),
			opts...),
		nil
}
