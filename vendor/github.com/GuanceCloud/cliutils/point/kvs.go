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
	anypb "google.golang.org/protobuf/types/known/anypb"
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

// TrimFields keep max-n field kvs.
func (x KVs) TrimFields(n int) (arr KVs) {
	cnt := 0

	for _, kv := range x {
		if kv.IsTag {
			arr = append(arr, kv)
			continue
		} else if cnt < n {
			arr = append(arr, kv)
			cnt++
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
		} else if cnt < n {
			arr = append(arr, kv)
			cnt++
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

// Del delete specified k.
func (x KVs) Del(k string) KVs {
	i := 0
	for _, f := range x {
		if f.Key != k {
			x[i] = f
			i++
		}
	}

	// remove not-needed elements.
	for j := i; j < len(x); j++ {
		x[j] = nil
	}

	x = x[:i]
	return x
}

// Add add new field
//
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

			goto out
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
		return KeyType_I
	case *Field_U:
		return KeyType_U
	case *Field_F:
		return KeyType_F
	case *Field_B:
		return KeyType_B
	case *Field_D:
		return KeyType_D
	case *Field_S:
		return KeyType_S
	case *Field_A:
		return KeyType_A

	default: // nil or other types
		return KeyType_X
	}
}

// Keys get k's value, if k not exist, return nil.
func (x KVs) Keys() *Keys {
	arr := []*Key{KeyMeasurement, KeyTime}

	for _, f := range x {
		t := PBType(f.Val)
		if t == KeyType_X {
			continue // ignore
		}

		arr = append(arr, NewKey(f.Key, t))
	}

	return &Keys{arr: arr}
}

func KVKey(f *Field) *Key {
	t := PBType(f.Val)

	return NewKey(f.Key, t)
}

type KVOption func(kv *Field)

func WithKVUnit(u string) KVOption {
	return func(kv *Field) {
		kv.Unit = u
	}
}

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

// NewKV get kv from specified key and value.
func NewKV(k string, v any, opts ...KVOption) *Field {
	var kv *Field

	switch x := v.(type) {
	case int8:
		kv = &Field{Key: k, Val: &Field_I{int64(x)}}
	case []int8:
		iarr, err := NewIntArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{iarr}}
		}

	case uint8:
		kv = &Field{Key: k, Val: &Field_U{uint64(x)}}
		// case []uint8 is []byte, skip it.

	case int16:
		kv = &Field{Key: k, Val: &Field_I{int64(x)}}

	case []int16:
		iarr, err := NewIntArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{iarr}}
		}

	case uint16:
		kv = &Field{Key: k, Val: &Field_U{uint64(x)}}

	case []uint16:
		iarr, err := NewUintArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{iarr}}
		}

	case int32:
		kv = &Field{Key: k, Val: &Field_I{int64(x)}}

	case []int32:
		iarr, err := NewIntArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{iarr}}
		}

	case uint32:
		kv = &Field{Key: k, Val: &Field_U{uint64(x)}}

	case []uint32:
		iarr, err := NewUintArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{iarr}}
		}

	case int:
		kv = &Field{Key: k, Val: &Field_I{int64(x)}}

	case []int:
		iarr, err := NewIntArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{iarr}}
		}

	case uint:
		kv = &Field{Key: k, Val: &Field_U{uint64(x)}}

	case []uint:
		iarr, err := NewUintArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{iarr}}
		}

	case int64:
		kv = &Field{Key: k, Val: &Field_I{x}}

	case []int64:
		iarr, err := NewIntArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{iarr}}
		}

	case uint64:
		kv = &Field{Key: k, Val: &Field_U{x}}

	case []uint64:
		iarr, err := NewUintArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{iarr}}
		}

	case float64:
		kv = &Field{Key: k, Val: &Field_F{x}}

	case []float64:
		farr, err := NewFloatArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{farr}}
		}

	case float32:
		kv = &Field{Key: k, Val: &Field_F{float64(x)}}

	case []float32:
		farr, err := NewFloatArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{farr}}
		}

	case string:
		kv = &Field{Key: k, Val: &Field_S{x}}

	case []string:
		sarr, err := NewStringArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{sarr}}
		}

	case []byte:
		kv = &Field{Key: k, Val: &Field_D{x}}

	case bool:
		kv = &Field{Key: k, Val: &Field_B{x}}

	case []bool:
		barr, err := NewBoolArray(x...)
		if err != nil {
			kv = &Field{Key: k, Val: nil}
		} else {
			kv = &Field{Key: k, Val: &Field_A{barr}}
		}

	case *anypb.Any:
		kv = &Field{Key: k, Val: &Field_A{x}}

	case nil: // pass
		kv = &Field{Key: k, Val: nil}

	default: // value ignored
		kv = &Field{Key: k, Val: nil}
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
		arr = append(arr, &Field{IsTag: true, Key: k, Val: &Field_S{S: v}})
	}

	return arr
}
