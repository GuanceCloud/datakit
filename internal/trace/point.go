// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package trace for DK trace.
package trace

import (
	"github.com/GuanceCloud/cliutils/point"
)

type DatakitTrace []*DkSpan

type DatakitTraces []DatakitTrace

type DkSpan struct {
	*point.Point
	// todo 对于必填字段没有值 必须加上 unknown
}

func (ds *DkSpan) GetFiledToInt64(key string) int64 {
	val := ds.Get(key)
	switch val.(type) {
	case int, int8, int32, int64:
		return val.(int64)
	default:
		return -100
	}
}

func (ds *DkSpan) GetFiledToString(key string) string {
	val := ds.Get(key)
	switch v := val.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func TagsFilter(source string, tags map[string]string, fields map[string]interface{}) point.KVs {
	tkvs := point.KVs{}
	for k, v := range tags {
		// filter
		tkvs = tkvs.AddTag(replacer.Replace(k), v)
	}
	for k, i := range fields {
		tkvs = tkvs.Add(replacer.Replace(k), i, false, false)
	}
	return tkvs
}

func NewAPMPoint(source string, kvs point.KVs, opts ...point.Option) *DkSpan {
	pt := point.NewPointV2(source, kvs, opts...)
	return &DkSpan{Point: pt}
}
