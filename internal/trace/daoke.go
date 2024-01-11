// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package trace for daoke
package trace

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
)

var log = logger.DefaultSLogger("jaeger_v2")

func SetLogger(set *logger.Logger) {
	log = set
}

func encoder(traceID string) string { //nolint
	lens := len(traceID)
	bs := make([]byte, lens/2)
	for i := 0; i < lens/2; i++ {
		b, _ := hex.DecodeString(traceID[i*2 : i*2+2])
		bs[i] = b[0]
	}
	return base64.StdEncoding.EncodeToString(bs)
}

func decoder(bs64 string) string {
	raw, err := base64.StdEncoding.DecodeString(bs64)
	if err != nil {
		log.Debugf("err=%v", err)
		return ""
	}

	return hex.EncodeToString(raw)
}

type SpanRefType int32

// nolint
type Span struct {
	TraceId       string      `json:"traceId,omitempty"`
	SpanId        string      `json:"spanId,omitempty"`
	OperationName string      `json:"operationName,omitempty"`
	References    []*SpanRef  `json:"references,omitempty"`
	Flags         uint32      `json:"flags,omitempty"`
	StartTime     *time.Time  `json:"startTime,omitempty"`
	Duration      string      `json:"duration,omitempty"`
	Tags          []*KeyValue `json:"tags,omitempty"`
	Logs          []*Log      `json:"logs,omitempty"`
	Process       *Process    `json:"process,omitempty"`
	ProcessId     string      `json:"processId,omitempty"`
	Warnings      []string    `json:"warnings,omitempty"`
}

// nolint
type SpanRef struct {
	TraceId string      `json:"traceId,omitempty"`
	SpanId  string      `json:"spanId,omitempty"`
	RefType SpanRefType `json:"ref_type,omitempty"`
}

// nolint
type KeyValue struct {
	Key      string `json:"key,omitempty"`
	VType    string `json:"vType,omitempty"`
	VStr     string `json:"vStr,omitempty"`
	VBool    bool   `json:"vBool,omitempty"`
	VInt64   string `json:"vInt64,omitempty"`
	VFloat64 string `json:"vFloat64,omitempty"`
	VBinary  []byte `json:"vBinary,omitempty"`
}

func (kv *KeyValue) getVal() string {
	switch kv.VType {
	case "STRING":
		return kv.VStr
	case "BOOL":
		return strconv.FormatBool(kv.VBool)
	case "INT64":
		return kv.VInt64
	case "FLOAT64":
		return kv.VFloat64
	case "BINARY":
		return ""
	default:
		return kv.VStr
	}
}

type Process struct {
	ServiceName string      `json:"serviceName,omitempty"`
	Tags        []*KeyValue `json:"tags,omitempty"`
}

type Log struct {
	Timestamp *time.Time  `protobuf:"bytes,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Fields    []*KeyValue `protobuf:"bytes,2,rep,name=fields,proto3" json:"fields,omitempty"`
}

func ParseToJaeger(msg []byte) (*point.Point, error) {
	jaegerSpan := &Span{}
	err := json.Unmarshal(msg, jaegerSpan)
	if err != nil {
		return nil, fmt.Errorf("message is %s ,err=%w", string(msg), err)
	}

	var spanKV point.KVs
	parentID := "0"
	spanType := "local"
	if len(jaegerSpan.References) != 0 {
		parentID = decoder(jaegerSpan.References[0].SpanId)
		spanType = "entry"
	}
	d, err := time.ParseDuration(jaegerSpan.Duration)
	if err != nil {
		d = 100
	}
	startTime := jaegerSpan.StartTime.UnixNano() / int64(time.Microsecond)
	spanKV = spanKV.Add(FieldTraceID, decoder(jaegerSpan.TraceId), false, false).
		Add(FieldParentID, parentID, false, false).
		Add(FieldSpanid, decoder(jaegerSpan.SpanId), false, false).
		Add(FieldStart, startTime, false, false).
		Add(FieldDuration, int64(d)/int64(time.Microsecond), false, false).
		AddTag(TagOperation, jaegerSpan.OperationName).
		Add(FieldResource, jaegerSpan.OperationName, false, false).
		AddTag(TagService, jaegerSpan.Process.ServiceName).
		AddTag(TagSpanType, spanType).
		AddTag(TagSpanStatus, StatusOk)
	for _, tag := range jaegerSpan.Process.Tags {
		spanKV = spanKV.AddTag(strings.ReplaceAll(tag.Key, ".", "_"), tag.getVal())
	}
	for _, sLog := range jaegerSpan.Logs {
		if len(sLog.Fields) > 0 {
			spanKV = spanKV.AddTag("log_"+strings.ReplaceAll(sLog.Fields[0].Key, ".", "_"), sLog.Fields[0].getVal())
		}
	}
	for _, tag := range jaegerSpan.Tags {
		if tag.Key == "error" {
			spanKV = spanKV.MustAddTag(TagSpanStatus, StatusErr)
		}
		spanKV = spanKV.AddTag(strings.ReplaceAll(tag.Key, ".", "_"), tag.getVal())
	}

	spanKV = spanKV.Add(FieldMessage, string(msg), false, false)
	pts := point.NewPointV2("otel_jaeger", spanKV, point.DefaultLoggingOptions()...)
	return pts, nil
}
