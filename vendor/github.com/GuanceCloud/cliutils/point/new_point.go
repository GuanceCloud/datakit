// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"sort"
	"time"
)

func NewPoint(name string, kvs KVs, opts ...Option) *Point {
	c := GetCfg(opts...)
	defer PutCfg(c)

	return doNewPoint(name, kvs, c)
}

// NewPointDeprecated returns a new Point given name(measurement), tags, fields and optional options.
//
// If fields empty(or nil), error ErrNoField will returned.
//
// Values in fields only allowed for int/uint(8-bit/16-bit/32-bit/64-bit), string, bool,
// float(32-bit/64-bit) and []byte, other types are ignored.
//
// Deprecated: use NewPoint.
func NewPointDeprecated(name string, tags map[string]string, fields map[string]any, opts ...Option) (*Point, error) {
	if len(fields) == 0 {
		return nil, ErrNoFields
	}

	kvs := NewKVs(fields)
	for k, v := range tags {
		kvs = kvs.SetTag(k, v) // force add these tags
	}

	c := GetCfg(opts...)
	defer PutCfg(c)

	return doNewPoint(name, kvs, c), nil
}

func doNewPoint(name string, kvs KVs, c *cfg) *Point {
	var pt *Point

	if defaultPTPool != nil {
		pt = defaultPTPool.Get()

		pt.SetName(name)
		pt.pt.Fields = kvs
		pt.SetFlag(Ppooled)
	} else {
		pt = &Point{
			pt: &PBPoint{
				Name:   name,
				Fields: kvs,
			},
		}
	}

	// add extra tags
	if len(c.extraTags) > 0 {
		for _, kv := range c.extraTags {
			pt.AddTag(kv.Key, kv.GetS()) // NOTE: do-not-override exist keys
		}
	}

	if c.enc == Protobuf {
		pt.SetFlag(Ppb)
	}

	if c.keySorted {
		kvs := KVs(pt.pt.Fields)
		sort.Sort(kvs)
		pt.pt.Fields = kvs
	}

	if c.precheck {
		chk := checker{cfg: c}
		pt = chk.check(pt)
		pt.SetFlag(Pcheck)
	}

	// sort again: during check, kv maybe update
	if c.keySorted {
		sort.Sort(KVs(pt.pt.Fields))
	}

	if c.timestamp >= 0 {
		pt.pt.Time = c.timestamp
	} else { // not set, use current time
		pt.pt.Time = time.Now().Round(0).UnixNano() // trim monotonic clock
	}

	return pt
}
