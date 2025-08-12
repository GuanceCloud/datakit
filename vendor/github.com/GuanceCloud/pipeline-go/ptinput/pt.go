// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ptinput

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
	"github.com/GuanceCloud/pipeline-go/ptinput/plcache"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
	"github.com/GuanceCloud/pipeline-go/ptinput/ptwindow"
	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
	"github.com/GuanceCloud/pipeline-go/ptinput/utils"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/goccy/go-json"
	"github.com/spf13/cast"
)

type Pt struct {
	pt       *point.Point
	category point.Category

	aggBuckets *plmap.AggBuckets
	ipdb       ipdb.IPdb
	refTable   refertable.PlReferTables
	cache      *plcache.Cache

	subPlpt []PlInputPt

	ptWindowPool       *ptwindow.WindowPool
	winKeyVal          [2][]string
	ptWindowRegistered bool

	disableStatusMapping bool

	// status
	drop bool
}

func (pp *Pt) GetStatusMapping() bool {
	return !pp.disableStatusMapping
}

func (pp *Pt) SetStatusMapping(val bool) {
	pp.disableStatusMapping = !val
}

func (pp *Pt) GetPtName() string {
	return pp.pt.Name()
}

func (pp *Pt) SetPtName(name string) {
	pp.pt.SetName(name)
}

func (pp *Pt) Get(k string) (any, ast.DType, error) {
	kv := pp.pt.KVs().Get(k)
	if kv == nil {
		return nil, ast.Nil, ErrKeyNotExist
	}
	v1, v2 := getVal(kv, false)
	return v1, v2, nil
}

func (pp *Pt) Set(k string, v any, dtype ast.DType) bool {
	pp._set(k, v, false, false)
	return true
}

func (pp *Pt) SetTag(k string, v any, dtype ast.DType) bool {
	pp._set(k, v, true, false)
	return true
}

func (pp *Pt) _set(k string, v any, asTag bool, asField bool) {
	// replace high level
	kv := pp.pt.KVs().Get(k)
	if kv != nil && kv.IsTag && !asField {
		asTag = true
	}

	v, _ = normalVal(v, asTag, false)

	pp.pt.MustAddKVs(point.NewKV(k, v, point.WithKVTagSet(asTag)))
}

func (pp *Pt) Delete(k string) {
	pp.pt.Del(k)
}

func (pp *Pt) RenameKey(from, to string) error {
	if from == to {
		return nil
	}

	idxFrom, idxTo := -1, -1
	kvs := pp.pt.KVs()
	for i := range kvs {
		switch kvs[i].Key {
		case from:
			idxFrom = i
		case to:
			idxTo = i
		}
	}
	if idxFrom < 0 {
		return nil
	}
	kvs[idxFrom].Key = to
	if idxTo >= 0 {
		tail := len(kvs) - 1
		kvs[idxTo] = kvs[tail]
		kvs = kvs[:tail]
	}
	pp.pt.PBPoint().Fields = kvs

	return nil
}

func (pp *Pt) MarkDrop(drop bool) {
	pp.drop = drop
}

func (pp *Pt) Dropped() bool {
	return pp.drop
}

func (pp *Pt) PtTime() time.Time {
	return pp.pt.Time()
}

func (pp *Pt) Category() point.Category {
	return pp.category
}

func (pp *Pt) Tags() map[string]string {
	tags := map[string]string{}
	for _, kv := range pp.pt.KVs() {
		if kv.IsTag {
			if v, ok := kv.Raw().(string); ok {
				tags[kv.Key] = v
			}
		}
	}
	return tags
}

func (pp *Pt) Fields() map[string]any {
	fields := map[string]any{}
	for _, kv := range pp.pt.KVs() {
		if !kv.IsTag {
			fields[kv.Key] = kv.Raw()
		}
	}
	return fields
}

func (pp *Pt) Point() *point.Point {
	return pp.pt
}

func (pp *Pt) KeyTime2Time() {
	if v, _, err := pp.Get("time"); err == nil {
		if nanots, ok := v.(int64); ok && nanots != 0 {
			pp.pt.PBPoint().Time = nanots
		}
		pp.Delete("time")
	}
}

func PtWrap(cat point.Category, pt *point.Point) *Pt {
	return &Pt{
		category: cat,
		pt:       pt,
	}
}

func getVal(kv *point.Field, allowComposite bool) (any, ast.DType) {
	if kv == nil {
		return nil, ast.Nil
	}

	var dt ast.DType
	var val any

	switch kv.Val.(type) {
	case *point.Field_I:
		return kv.GetI(), ast.Int
	case *point.Field_U:
		return int64(kv.GetU()), ast.Int
	case *point.Field_F:
		return kv.GetF(), ast.Float
	case *point.Field_B:
		return kv.GetB(), ast.Bool
	case *point.Field_D:
		return string(kv.GetD()), ast.String
	case *point.Field_S:
		return kv.GetS(), ast.String

	case *point.Field_A:
		v, err := point.AnyRaw(kv.GetA())
		if err != nil {
			return nil, ast.Nil
		}
		switch v.(type) {
		case []any:
			val, dt = v, ast.List
		case map[string]any:
			val, dt = v, ast.Map
		default:
			return nil, ast.Nil
		}
	default:
		return nil, ast.Nil
	}

	if !allowComposite {
		if dt == ast.Map || dt == ast.List {
			if v, err := Conv2String(val, dt); err == nil {
				return v, ast.String
			}
		}
	} else {
		return val, dt
	}

	return nil, ast.Nil
}

func normalVal(v any, conv2Str bool, allowComposite bool) (any, ast.DType) {
	var val any
	var dt ast.DType
	switch v := v.(type) {
	case string:
		val, dt = v, ast.String
	case int64:
		val, dt = v, ast.Int
	case int32, int8, int16, int,
		uint, uint16, uint32, uint64, uint8:
		val, dt = cast.ToInt64(v), ast.Int
	case float64:
		val, dt = v, ast.Float
	case float32:
		val, dt = cast.ToFloat64(v), ast.Float
	case bool:
		val, dt = v, ast.Bool
	case []any:
		val, dt = v, ast.List
	case map[string]any:
		val, dt = v, ast.Map
	case []byte:
		val, dt = string(v), ast.String
	default:
		val, dt = nil, ast.Nil
	}

	if conv2Str {
		if dt == ast.Nil {
			return "", ast.String
		}
		if v, err := Conv2String(val, dt); err == nil {
			return v, ast.String
		} else {
			return "", ast.String
		}
	} else {
		if !allowComposite {
			if dt == ast.Map || dt == ast.List {
				if v, err := Conv2String(val, dt); err == nil {
					return v, ast.String
				} else {
					return nil, ast.Nil
				}
			}
		}
		return val, dt
	}
}

func Conv2String(v any, dtype ast.DType) (string, error) {
	switch dtype { //nolint:exhaustive
	case ast.Int, ast.Float, ast.Bool, ast.String:
		return cast.ToString(v), nil
	case ast.List, ast.Map:
		res, err := json.Marshal(v)
		return string(res), err
	case ast.Nil:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported data type %d", dtype)
	}
}

func NewPlPt(cat point.Category, name string,
	tags map[string]string, fields map[string]any, ptTime time.Time,
) PlInputPt {
	kvs := point.NewKVs(fields)
	for k, v := range tags {
		kvs = kvs.SetTag(k, v)
	}

	opt := utils.PtCatOption(cat)

	if !ptTime.IsZero() {
		opt = append(opt, point.WithTime(ptTime))
	}

	pt := point.NewPoint(name, kvs, opt...)
	return &Pt{
		pt:       pt,
		category: cat,
	}
}
