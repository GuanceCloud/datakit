// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"crypto/md5" //nolint:gosec
	"crypto/sha256"
	"fmt"
	"sort"
)

type EqualOption func(*eqopt)

type eqopt struct {
	withMeasurement bool
	excludeKeys     []string
}

func (o *eqopt) keyExlcuded(key string) bool {
	for _, k := range o.excludeKeys {
		if k == key {
			return true
		}
	}

	return false
}

func (o *eqopt) kvsEq(l, r KVs) (bool, string) {
	if len(o.excludeKeys) == 0 && len(l) != len(r) {
		return false, fmt.Sprintf("count not equal(%d <> %d)", len(l), len(r))
	}

	for _, f := range l {
		if o.keyExlcuded(f.Key) {
			continue
		}

		if !r.Has(f.Key) { // key not exists
			return false, fmt.Sprintf("%s not exists", f.Key)
		}

		v := r.Get(f.Key)
		if f.String() != v.String() { // compare proto-string format value
			return false, fmt.Sprintf("%q value not deep equal(%s <> %s)", f.Key, f, v)
		}
	}
	return true, ""
}

// EqualWithMeasurement set compare on points with/without measurement.
func EqualWithMeasurement(on bool) EqualOption {
	return func(o *eqopt) { o.withMeasurement = on }
}

// EqualWithoutKeys set compare on points without specific keys.
func EqualWithoutKeys(keys ...string) EqualOption {
	return func(o *eqopt) { o.excludeKeys = append(o.excludeKeys, keys...) }
}

// Equal test if two point are the same.
// Equality test NOT check on warns and debugs.
// If two points equal, they have the same ID(MD5/Sha256),
// but same ID do not means they are equal.
func (p *Point) Equal(x *Point, opts ...EqualOption) bool {
	eq, _ := p.EqualWithReason(x, opts...)
	return eq
}

func (p *Point) EqualWithReason(x *Point, opts ...EqualOption) (bool, string) {
	if x == nil {
		return false, "empty point"
	}

	eopt := &eqopt{withMeasurement: true}
	for _, opt := range opts {
		if opt != nil {
			opt(eopt)
		}
	}

	pname := p.Name() //nolint:ifshort
	ptags := p.Tags()
	pfields := p.Fields()

	xtags := x.Tags()
	xfields := x.Fields()

	if !eopt.keyExlcuded("time") {
		if xtime, ptime := x.Time().UnixNano(), p.Time().UnixNano(); xtime != ptime {
			return false, fmt.Sprintf("timestamp not equal(%d <> %d)", ptime, xtime)
		}
	}

	if eopt.withMeasurement {
		if xname := x.Name(); eopt.withMeasurement && xname != pname {
			return false, fmt.Sprintf("measurement not equla(%s <> %s)", pname, xname)
		}
	}

	if len(eopt.excludeKeys) == 0 && len(xtags) != len(ptags) {
		return false, fmt.Sprintf("tag count not equal(%d <> %d)", len(ptags), len(xtags))
	}

	if eq, reason := eopt.kvsEq(pfields, xfields); !eq {
		return eq, fmt.Sprintf("field: %s", reason)
	}

	if eq, reason := eopt.kvsEq(ptags, xtags); !eq {
		return eq, fmt.Sprintf("tag: %s", reason)
	}

	return true, ""
}

// MD5 get point MD5 id.
func (p *Point) MD5() string {
	x := p.hashstr()
	return fmt.Sprintf("%x", md5.Sum(x)) //nolint:gosec
}

// Sha256 get point Sha256 id.
func (p *Point) Sha256() string {
	x := p.hashstr()
	h := sha256.New()
	h.Write(x)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// hashstr only count measurement/tag-keys/tag-values as hash string,
// other fields(fields/time/debugs/warns ignored).
func (p *Point) hashstr() []byte {
	tags := p.Tags()

	var data []byte

	data = append(data, []byte(p.Name())...)

	sort.Sort(tags)

	for _, t := range tags {
		data = append(data, []byte(t.Key)...)
		data = append(data, []byte(t.GetS())...)
	}
	return data
}

func (p *Point) TimeSeriesHash() []string {
	fields := p.Fields()
	ts := make([]string, len(fields))
	hash := p.hashstr()

	for idx, f := range fields {
		hash := append(hash, []byte(f.Key)...)
		ts[idx] = fmt.Sprintf("%x", md5.Sum(hash)) //nolint:gosec
	}

	return ts
}
