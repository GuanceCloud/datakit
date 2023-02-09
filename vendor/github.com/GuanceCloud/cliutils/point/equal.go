// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"bytes"
	"crypto/md5" //nolint:gosec
	"crypto/sha256"
	"fmt"
	"sort"
)

// Equal test if two point are the same.
// Equality test not check on warns and debugs.
// If two points equal, they have the same ID(MD5/Sha256),
// but same ID do not means they are equal.
func (p *Point) Equal(x *Point) bool {
	eq, _ := p.EqualWithReason(x)
	return eq
}

func (p *Point) EqualWithReason(x *Point) (bool, string) {
	if x == nil {
		return false, "empty point"
	}

	pname := p.Name() //nolint:ifshort
	ptags := p.Tags()
	pfields := p.Fields()

	xname := x.Name()
	xtags := x.Tags()
	xfields := x.Fields()

	if xtime, ptime := x.Time().UnixNano(), p.Time().UnixNano(); xtime != ptime {
		return false, fmt.Sprintf("timestamp not equal(%d <> %d)", ptime, xtime)
	}

	if !bytes.Equal(xname, pname) {
		return false, fmt.Sprintf("measurement not equla(%s <> %s)", pname, xname)
	}

	if eq, reason := kvsEq(pfields, xfields); !eq {
		return eq, reason
	}

	if len(xtags) != len(ptags) {
		return false, fmt.Sprintf("tag count not equal(%d <> %d)", len(ptags), len(xtags))
	}

	for _, t := range ptags {
		if !xtags.Has(t.Key) {
			return false, fmt.Sprintf("tag %s not exists", t.Key)
		}

		kv := xtags.Get(t.Key)
		if !bytes.Equal(t.GetD(), kv.GetD()) {
			return false, fmt.Sprintf("tag %q value not equal(%s <> %s)", t.Key, t, kv)
		}
	}

	return true, ""
}

func kvsEq(l, r KVs) (bool, string) {
	if len(l) != len(r) {
		return false, fmt.Sprintf("field count not equal(%d <> %d)", len(l), len(r))
	}

	for _, f := range l {
		if !r.Has(f.Key) { // key not exists
			return false, fmt.Sprintf("field %s not exists", string(f.Key))
		}

		v := r.Get(f.Key)
		if f.String() != v.String() {
			return false, fmt.Sprintf("field %q value not deep equal(%s <> %s)", f.Key, f, v)
		}
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

	data = append(data, p.Name()...)

	sort.Sort(tags)

	for _, t := range tags {
		data = append(data, t.Key...)
		data = append(data, t.GetD()...)
	}
	return data
}
