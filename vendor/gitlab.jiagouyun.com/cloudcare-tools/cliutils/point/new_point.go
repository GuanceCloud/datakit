// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"sort"
	"time"
)

func NewPointV2(name []byte, tags Tags, fields Fields, opts ...Option) *Point {
	return doNewPoint(name, tags, fields, opts...)
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

	// We should deep copy @tags, do not reuse @tags here, the @tags may
	// comes from input's configure, we should not modify it.
	newTags := PBTags(tags)

	// Also deep copy @fields, but reuse of @fields maybe ok.
	newFields := PBFields(fields)

	return doNewPoint([]byte(name), newTags, newFields, opts...), nil
}

func doNewPoint(name []byte, tags Tags, fields Fields, opts ...Option) *Point {
	c := defaultCfg()

	for _, opt := range opts {
		opt(c)
	}

	// add extra tags
	if len(c.extraTags) > 0 {
		if tags == nil {
			tags = c.extraTags
		} else {
			for _, t := range c.extraTags {
				if tags.KeyExist(t.Key) { // NOTE: do-not-override exist tag
					continue
				} else {
					tags = append(tags, t)
				}
			}
		}
	}

	pt := &Point{
		name:   name,
		tags:   tags,
		fields: fields,
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

	sort.Sort(pt.tags)
	sort.Sort(pt.fields)

	return pt
}
