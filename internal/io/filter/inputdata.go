// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package filter

import (
	fp "github.com/GuanceCloud/cliutils/filter"
	"github.com/GuanceCloud/cliutils/point"
)

// Before checks, should adjust tags under some conditions.
// Must stay the same 'switch' logic with kodo project function named 'getSourceValue' in source file apis/esFields.go.

var _ fp.KVs = (*TFData)(nil)

func NewTFDataFromMap(data map[string]string) *TFData {
	return &TFData{
		Tags: data,
	}
}

func NewTFData(category point.Category, pt *point.Point) *TFData {
	res := &TFData{
		Tags:   map[string]string{},
		Fields: map[string]any{},
	}

	for _, kv := range pt.KVs() {
		if kv.IsTag {
			res.Tags[kv.Key] = kv.GetS()
		} else {
			res.Fields[kv.Key] = kv.Raw()
		}
	}

	// Before checks, should adjust tags under some conditions.
	// Must stay the same 'switch' logic with kodo project function named 'getSourceValue' in source file apis/esFields.go.
	switch category {
	case
		point.Logging,
		point.Network,
		point.KeyEvent,
		point.RUM:
		res.Tags["source"] = pt.Name() // set measurement name as tag `source'

	case
		point.Tracing,
		point.Security,
		point.Profiling:
		// using measurement name as tag `service'.

	case point.Metric, point.MetricDeprecated:
		res.Tags["measurement"] = pt.Name() // set measurement name as tag `measurement'

	case point.Object, point.CustomObject:
		res.Tags["class"] = pt.Name() // set measurement name as tag `class'

	case point.DynamicDWCategory, point.UnknownCategory:
		// pass
	}

	res.Tags["category"] = category.String()

	return res
}

func NewTFDataFromPoint(category point.Category, pt *point.Point) *TFData {
	res := &TFData{
		Tags: map[string]string{
			"category": category.String(),
		},
		Fields: map[string]any{},
	}

	for _, t := range pt.Tags() {
		res.Tags[t.Key] = t.GetS()
	}

	for _, t := range pt.Fields() {
		if v := t.GetS(); v != "" {
			res.Fields[t.Key] = v
		}
	}

	// Before checks, should adjust tags under some conditions.
	// Must stay the same 'switch' logic with kodo project function named 'getSourceValue' in source file apis/esFields.go.
	switch category {
	case
		point.Logging,
		point.Network,
		point.KeyEvent,
		point.RUM:
		res.Tags["source"] = pt.Name() // set measurement name as tag `source'

	case
		point.Tracing,
		point.Security,
		point.Profiling:
		// using measurement name as tag `service'.

	case point.Metric, point.MetricDeprecated:
		res.Tags["measurement"] = pt.Name() // set measurement name as tag `measurement'

	case point.Object, point.CustomObject:
		res.Tags["class"] = pt.Name() // set measurement name as tag `class'

	case point.DynamicDWCategory, point.UnknownCategory:
		// pass
	}

	return res
}

type TFData struct {
	Tags   map[string]string
	Fields map[string]any
}

func (d *TFData) MergeStringKVs() {
	for k, v := range d.Fields {
		switch str := v.(type) {
		case string:
			if d.Tags == nil {
				d.Tags = map[string]string{}
			}
			d.Tags[k] = str
		default:
			// pass: ignore non-string field
		}
	}

	d.Fields = nil
}

func (d *TFData) Get(name string) (any, bool) {
	if v, ok := d.Tags[name]; ok {
		return v, true
	}

	if v, ok := d.Fields[name]; ok {
		return v, true
	}

	return nil, false
}
