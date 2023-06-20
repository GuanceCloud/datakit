// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package filter

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/parser"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

// Before checks, should adjust tags under some conditions.
// Must stay the same 'switch' logic with kodo project function named 'getSourceValue' in source file apis/esFields.go.

var _ parser.KVs = (*tfData)(nil)

func newTFData(category point.Category, pt *dkpt.Point) (*tfData, error) {
	tags := pt.Tags()
	fields, err := pt.Fields()
	if err != nil {
		return nil, err
	}

	// Before checks, should adjust tags under some conditions.
	// Must stay the same 'switch' logic with kodo project function named 'getSourceValue' in source file apis/esFields.go.
	switch category {
	case
		point.Logging,
		point.Network,
		point.KeyEvent,
		point.RUM:
		tags["source"] = pt.Name() // set measurement name as tag `source'
	case
		point.Tracing,
		point.Security,
		point.Profiling:
		// using measurement name as tag `service'.
	case point.Metric, point.MetricDeprecated:
		tags["measurement"] = pt.Name() // set measurement name as tag `measurement'
	case point.Object, point.CustomObject:
		tags["class"] = pt.Name() // set measurement name as tag `class'
	case point.DynamicDWCategory, point.UnknownCategory:
		return nil, fmt.Errorf("unsupport category: %s", category)
	}
	return &tfData{
		tags:   tags,
		fields: fields,
	}, nil
}

type tfData struct {
	tags   map[string]string
	fields map[string]any
}

func (d *tfData) Get(name string) (any, bool) {
	if v, ok := d.tags[name]; ok {
		return v, true
	}

	if v, ok := d.fields[name]; ok {
		return v, true
	}

	return nil, false
}
