// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"sort"
	"time"
)

func NewPointV2(name []byte, kvs KVs, opts ...Option) *Point {
	c := GetCfg(opts...)
	defer PutCfg(c)

	return doNewPoint(name, kvs, c)
}

// NewPoint returns a new Point given name(measurement), tags, fields and optional options.
//
// If fields empty(or nil), error ErrNoField will returned.
//
// Values in fields only allowed for int/uint(8-bit/16-bit/32-bit/64-bit), string, bool,
// float(32-bit/64-bit) and []byte, other types are ignored.
//
// Deprecated: use NewPointV2.
func NewPoint(name string, tags map[string]string, fields map[string]any, opts ...Option) (*Point, error) {
	if len(fields) == 0 {
		return nil, ErrNoFields
	}

	kvs := NewKVs(fields)
	for k, v := range tags {
		kvs = kvs.Add([]byte(k), []byte(v), true, true) // force add these tags
	}

	c := GetCfg(opts...)
	defer PutCfg(c)

	return doNewPoint([]byte(name), kvs, c), nil
}

func doNewPoint(name []byte, kvs KVs, c *cfg) *Point {
	pt := &Point{
		name: name,
		kvs:  kvs,
	}

	// add extra tags
	if len(c.extraTags) > 0 {
		for _, kv := range c.extraTags {
			pt.AddTag(kv.Key, kv.GetD()) // NOTE: do-not-override exist keys
		}
	}

	if c.enc == Protobuf {
		pt.SetFlag(Ppb)
	}

	if c.precheck {
		chk := checker{cfg: c}
		pt = chk.check(pt)
		pt.SetFlag(Pcheck)
		pt.warns = chk.warns
	}

	if !c.t.IsZero() {
		pt.time = c.t
	}

	if pt.time.IsZero() {
		pt.time = time.Now()
	}

	sort.Sort(pt.kvs)

	return pt
}
