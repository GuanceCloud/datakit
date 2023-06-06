// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

//go:generate msgp -file=span.pb.go -o span_gen.go -io=false
//go:generate msgp -o trace_gen.go -io=false

type DDTrace []*DDSpan

type DDTraces []DDTrace
