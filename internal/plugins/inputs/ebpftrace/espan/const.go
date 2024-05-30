// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package espan

const (
	OUTG = "outgoing"
	INCO = "incoming"

	ENCHEX = "hex" // default
	ENCDEC = "dec"

	SpanTypEntry = "entry"
)

const (
	// 原始数据中必须存在.
	DirectionStr = "direction"

	// NetTraceID     = "net_trace_id".

	AppTraceIDL    = "app_trace_id_l"
	AppTraceIDH    = "app_trace_id_h"
	AppParentIDL   = "app_parent_id_l"
	AppSpanSampled = "app_span_sampled"
	AppTraceEncode = "app_trace_encode"

	ThrTraceID = "thread_trace_id"
	ReqSeq     = "req_seq"
	RespSeq    = "resp_seq"

	EBPFSpanType = "ebpf_span_type"

	// 插入时生成.
	SpanID = "span_id"

	ParentID = "parent_id"
	TraceID  = "trace_id"

	EBPFTraceID  = "ebpf_trace_id"
	EBPFParentID = "ebpf_parent_id"
)
