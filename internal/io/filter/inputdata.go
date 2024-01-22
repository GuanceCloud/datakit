// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package filter

import (
	"sync"

	fp "github.com/GuanceCloud/cliutils/filter"
	"github.com/GuanceCloud/cliutils/point"
)

// Before checks, should adjust tags under some conditions.
// Must stay the same 'switch' logic with kodo project function named 'getSourceValue' in source file apis/esFields.go.

var _ fp.KVs = (*KVs)(nil)

func (d *KVs) Setup(category point.Category, pt *point.Point) {
	d.extKVs = append(d.extKVs, [2]string{"category", category.String()})
	d.pt = pt

	// Before checks, should adjust tags under some conditions.
	// Must stay the same 'switch' logic with kodo project function named 'getSourceValue' in source file apis/esFields.go.
	switch category {
	case
		point.Logging,
		point.Network,
		point.KeyEvent,
		point.RUM:

		// set measurement name as tag `source'
		d.extKVs = append(d.extKVs, [2]string{"source", pt.Name()})

	case
		point.Tracing,
		point.Security,
		point.Profiling:
		// using measurement name as tag `service'.

	case point.Metric, point.MetricDeprecated:
		// set measurement name as tag `measurement'
		d.extKVs = append(d.extKVs, [2]string{"measurement", pt.Name()})

	case point.Object, point.CustomObject:
		// set measurement name as tag `class'
		d.extKVs = append(d.extKVs, [2]string{"class", pt.Name()})

	case point.DynamicDWCategory, point.UnknownCategory:
		// pass
	}
}

var kvsPool sync.Pool

func getTFData() *KVs {
	if x := kvsPool.Get(); x == nil {
		return &KVs{}
	} else {
		return x.(*KVs)
	}
}

func putTFData(d *KVs) {
	d.pt = nil
	d.extKVs = d.extKVs[:0]
	d.cat = point.UnknownCategory
	kvsPool.Put(d)
}

type KVs struct {
	pt     *point.Point
	cat    point.Category
	extKVs [][2]string
}

func (d *KVs) Get(name string) (any, bool) {
	if v := d.pt.Get(name); v != nil {
		return v, true
	}

	for _, kv := range d.extKVs {
		if kv[0] == name {
			return kv[1], true
		}
	}

	return nil, false
}
