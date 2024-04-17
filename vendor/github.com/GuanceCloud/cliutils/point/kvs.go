// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"math"
	"sort"
	"strings"

	influxm "github.com/influxdata/influxdb1-client/models"
	"golang.org/x/exp/slices"
)

type KVs []*Field

// Raw return underlying raw data.
func (kv *Field) Raw() any {
	switch kv.Val.(type) {
	case *Field_I:
		return kv.GetI()
	case *Field_U:
		return kv.GetU()
	case *Field_F:
		return kv.GetF()
	case *Field_B:
		return kv.GetB()
	case *Field_D:
		return kv.GetD()
	case *Field_S:
		return kv.GetS()

	case *Field_A:

		if v, err := AnyRaw(kv.GetA()); err != nil {
			return nil
		} else {
			return v
		}
	default:
		return nil
	}
}

func (x KVs) Len() int {
	return len(x)
}

func (x KVs) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x KVs) Less(i, j int) bool {
	return strings.Compare(x[i].Key, x[j].Key) < 0 // stable sort
}

func (x KVs) Pretty() string {
	var arr []string
	for _, kv := range x {
		arr = append(arr, kv.String())
	}

	// For key-values are not sorted while building the point, we
	// think they are equal, so sort the string array to remove the
	// ordering difference between points.
	sort.Strings(arr)

	return strings.Join(arr, "\n")
}

// InfluxFields convert KVs to map structure.
func (x KVs) InfluxFields() map[string]any {
	res := map[string]any{}

	for _, kv := range x {
		if kv.IsTag {
			continue
		}

		switch x := kv.Val.(type) {
		case *Field_I:
			res[kv.Key] = x.I
		case *Field_U:
			if x.U <= math.MaxInt64 {
				res[kv.Key] = int64(x.U)
			} // else: dropped, see lp_test.go/parse-uint
		case *Field_F:
			res[kv.Key] = x.F
		case *Field_B:
			res[kv.Key] = x.B
		case *Field_D:
			res[kv.Key] = string(x.D)
		case *Field_S:
			res[kv.Key] = x.S

		case *Field_A:
			if v, err := AnyRaw(kv.GetA()); err != nil {
				// pass
			} else {
				res[kv.Key] = v
			}
		default:
			continue
		}
	}

	return res
}

// InfluxTags convert tag KVs to map structure.
func (x KVs) InfluxTags() (res influxm.Tags) {
	for _, kv := range x {
		if !kv.IsTag {
			continue
		}

		res = append(res, influxm.Tag{
			Key:   []byte(kv.Key),
			Value: []byte(kv.GetS()),
		})
	}

	// keep tags sorted used to build lineprotocol text
	sort.Sort(res)

	return
}

func clearKV(kv *Field) *Field {
	kv.Key = ""
	kv.IsTag = false
	kv.Type = UNSPECIFIED
	kv.Unit = ""
	return kv
}

func resetKV(kv *Field) *Field {
	switch v := kv.Val.(type) {
	case *Field_I:
		v.I = 0
	case *Field_U:
		v.U = 0
	case *Field_F:
		v.F = 0.0
	case *Field_D:
		v.D = v.D[:0]
	case *Field_B:
		v.B = false
	case *Field_S:
		v.S = ""
	case *Field_A:
		v.A.TypeUrl = ""
		v.A.Value = v.A.Value[:0]
	}

	return kv
}

// ResetFull reset and reuse key-value.
func (x KVs) ResetFull() {
	for i, kv := range x {
		x[i] = resetKV(clearKV(kv))
	}
}

// Reset reset but drop value.
func (x KVs) Reset() {
	for i, kv := range x {
		kv = clearKV(kv)
		kv.Val = nil // drop Val
		x[i] = kv
	}
}

// Has test if k exist.
func (x KVs) Has(k string) bool {
	for _, f := range x {
		if f.Key == k {
			return true
		}
	}

	return false
}

// Get get k's value, if k not exist, return nil.
func (x KVs) Get(k string) *Field {
	for _, f := range x {
		if f.Key == k {
			return f
		}
	}

	return nil
}

// GetTag get tag k's value, if the tag not exist, return nil.
func (x KVs) GetTag(k string) string {
	for _, f := range x {
		if !f.IsTag {
			continue
		}

		if f.Key == k {
			return f.GetS()
		}
	}

	return ""
}

func (x KVs) Tags() (arr KVs) {
	for _, kv := range x {
		if !kv.IsTag {
			continue
		}

		arr = append(arr, kv)
	}

	// should we buffer point's tags like this?
	//   p.tags = arr
	return arr
}

func (x KVs) Fields() (arr KVs) {
	for _, kv := range x {
		if kv.IsTag {
			continue
		}

		arr = append(arr, kv)
	}

	// should we buffer point's tags like this?
	//   p.tags = arr
	return arr
}

// TrimFields keep max-n field kvs and drop the rest.
func (x KVs) TrimFields(n int) (arr KVs) {
	cnt := 0

	if len(x) <= n {
		return x
	}

	for _, kv := range x {
		if kv.IsTag {
			arr = append(arr, kv)
			continue
		} else {
			if cnt < n {
				arr = append(arr, kv)
				cnt++
			} else if defaultPTPool != nil { // drop the kv
				defaultPTPool.PutKV(kv)
			}
		}
	}

	return arr
}

// TrimTags keep max-n tag kvs.
func (x KVs) TrimTags(n int) (arr KVs) {
	cnt := 0

	for _, kv := range x {
		if !kv.IsTag {
			arr = append(arr, kv)
			continue
		} else {
			if cnt < n {
				arr = append(arr, kv)
				cnt++
			} else if defaultPTPool != nil {
				defaultPTPool.PutKV(kv)
			}
		}
	}

	return arr
}

func (x KVs) TagCount() (i int) {
	for _, kv := range x {
		if kv.IsTag {
			i++
		}
	}
	return
}

func (x KVs) FieldCount() (i int) {
	for _, kv := range x {
		if !kv.IsTag {
			i++
		}
	}
	return
}

// Del delete field from x with Key == k.
func (x KVs) Del(k string) KVs {
	for i, f := range x {
		if f.Key == k {
			x = slices.Delete(x, i, i+1)
			if defaultPTPool != nil {
				defaultPTPool.PutKV(f)
			}
		}
	}

	return x
}

// AddV2 add new field with opts.
// If force enabled, overwrite exist key.
func (x KVs) AddV2(k string, v any, force bool, opts ...KVOption) KVs {
	kv := NewKV(k, v, opts...)

	for i := range x {
		if x[i].Key == k { // k exist
			if force {
				x[i] = kv // override exist tag/field
			}

			goto out // ignore the key
		}
	}

	x = append(x, kv)

out:
	return x
}

// Add add new field.
// Deprecated: use AddV2
// If force enabled, overwrite exist key.
func (x KVs) Add(k string, v any, isTag, force bool) KVs {
	kv := NewKV(k, v)

	if isTag {
		switch v.(type) {
		case string:
			kv.IsTag = isTag
		default:
			// ignore isTag
		}
	}

	for i := range x {
		if x[i].Key == k { // k exist
			if force {
				x[i] = kv // override exist tag/field
			}

			goto out // ignore the key
		}
	}

	x = append(x, kv)

out:
	return x
}

func (x KVs) AddTag(k, v string) KVs {
	x = x.Add(k, v, true, false)
	return x
}

func (x KVs) MustAddTag(k, v string) KVs {
	return x.Add(k, v, true, true)
}

func (x KVs) AddKV(kv *Field, force bool) KVs {
	if kv == nil {
		return x
	}

	for i := range x {
		if x[i].Key == kv.Key {
			if force {
				x[i] = kv
			}
			goto out
		}
	}

	x = append(x, kv)

out:
	return x
}

func (x KVs) MustAddKV(kv *Field) KVs {
	x = x.AddKV(kv, true)
	return x
}

func PBType(v isField_Val) KeyType {
	switch v.(type) {
	case *Field_I:
		return I
	case *Field_U:
		return U
	case *Field_F:
		return F
	case *Field_B:
		return B
	case *Field_D:
		return D
	case *Field_S:
		return S
	case *Field_A:
		return A
	default: // nil or other types
		return X
	}
}

// Keys get k's value, if k not exist, return nil.
func (x KVs) Keys() *Keys {
	arr := []*Key{KeyMeasurement, KeyTime}

	for _, f := range x {
		t := PBType(f.Val)
		if t == X {
			continue // ignore
		}

		arr = append(arr, NewKey(f.Key, t))
	}

	return &Keys{arr: arr}
}

func KVKey(kv *Field) *Key {
	t := PBType(kv.Val)

	return NewKey(kv.Key, t)
}

type KVOption func(kv *Field)

// WithKVUnit set value's unit.
func WithKVUnit(u string) KVOption {
	return func(kv *Field) {
		kv.Unit = u
	}
}

// WithKVType set field type(count/gauge/rate).
func WithKVType(t MetricType) KVOption {
	return func(kv *Field) {
		kv.Type = t
	}
}

func WithKVTagSet(on bool) KVOption {
	return func(kv *Field) {
		switch kv.Val.(type) {
		case *Field_S:
			kv.IsTag = on
		default:
			// ignored
		}
	}
}

func doNewKV(k string, v any, opts ...KVOption) *Field {
	return &Field{
		Key: k,
		Val: newVal(v),
	}
}

// NewKV get kv on specified key and value.
func NewKV(k string, v any, opts ...KVOption) *Field {
	var kv *Field
	if defaultPTPool != nil {
		kv = defaultPTPool.GetKV(k, v)
	} else {
		kv = doNewKV(k, v, opts...)
	}

	for _, opt := range opts {
		if opt != nil {
			opt(kv)
		}
	}

	return kv
}

// NewKVs create kvs slice from map structure.
func NewKVs(kvs map[string]interface{}) (res KVs) {
	for k, v := range kvs {
		res = append(res, NewKV(k, v))
	}

	return res
}

// NewTags create tag kvs from map structure.
func NewTags(tags map[string]string) (arr KVs) {
	for k, v := range tags {
		arr = append(arr, NewKV(k, v, WithKVTagSet(true)))
	}

	return arr
}
