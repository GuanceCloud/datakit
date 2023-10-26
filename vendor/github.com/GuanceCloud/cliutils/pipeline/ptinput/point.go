// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ptinput impl ppl input interface
package ptinput

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/pipeline/ptinput/ipdb"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/plmap"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/refertable"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/utils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/platypus/pkg/ast"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"

	"github.com/spf13/cast"
)

type (
	KeyKind uint
	PtFlag  uint
)

var _ PlInputPt = (*PlPoint)(nil)

type PlInputPt interface {
	GetPtName() string
	SetPtName(m string)

	Get(key string) (any, ast.DType, error)
	GetWithIsTag(key string) (any, bool, bool)
	Set(key string, value any, dtype ast.DType) error
	Delete(key string)
	RenameKey(from, to string) error

	SetTag(key string, value any, dtype ast.DType) error

	PtTime() time.Time

	GetAggBuckets() *plmap.AggBuckets
	SetAggBuckets(*plmap.AggBuckets)

	AppendSubPoint(PlInputPt)
	GetSubPoint() []PlInputPt
	Category() point.Category

	KeyTime2Time()

	MarkDrop(bool)
	Dropped() bool
	Point() *point.Point

	GetIPDB() ipdb.IPdb
	SetIPDB(ipdb.IPdb)

	GetPlReferTables() refertable.PlReferTables
	SetPlReferTables(refertable.PlReferTables)

	Tags() map[string]string
	Fields() map[string]any
}

const (
	PtMeasurement PtFlag = iota
	PtTag
	PtField
	PtTFDefaulutOrKeep
	PtTime
)

const (
	KindPtDefault KeyKind = iota
	KindPtTag
)

const Originkey = "message"

type InputWithVarbMapRW interface {
	Get(key string) (any, ast.DType, bool)
	Set(key string, value any, dtype ast.DType) bool
	Delete(key string) bool
}

type InputWithVarbMapR interface {
	Get(key string) (any, ast.DType, bool)
}

type InputWithoutVarbMap interface{}

type DoFeedCache func(name, category string, pt *point.Point) error

var DoFeedNOP = func(name, category string, pt *point.Point) error { return nil }

type PlmapManager interface {
	// createPtCaheMap(category string, source PtSource) (*fucs.PtCacheMap, bool)
}

type PlPoint struct {
	name   string
	tags   map[string]string
	fields map[string]any // int, float, bool, string, map, slice, array
	time   time.Time

	aggBuckets *plmap.AggBuckets
	ipdb       ipdb.IPdb
	refTable   refertable.PlReferTables

	subPlpt []PlInputPt

	category point.Category
	drop     bool
}

func NewPlPoint(category point.Category, name string,
	tags map[string]string, fields map[string]any, ptTime time.Time,
) PlInputPt {
	if tags == nil {
		tags = map[string]string{}
	}

	if fields == nil {
		fields = map[string]any{}
	}

	dPt := &PlPoint{
		name:     name,
		tags:     tags,
		fields:   fields,
		time:     ptTime,
		category: category,
	}
	return dPt
}

func valueDtype(v any) (any, ast.DType) {
	switch v := v.(type) {
	case int32, int8, int16, int,
		uint, uint16, uint32, uint64, uint8:
		return cast.ToInt64(v), ast.Int
	case int64:
		return v, ast.Int
	case float32:
		return cast.ToFloat64(v), ast.Float
	case float64:
		return v, ast.Float
	case bool:
		return v, ast.Bool
	case []byte:
		return string(v), ast.String
	case string:
		return v, ast.String
	}

	// ignore unknown type
	return nil, ast.Nil
}

func (pt *PlPoint) GetPtName() string {
	return pt.name
}

func (pt *PlPoint) SetPtName(m string) {
	pt.name = m
}

func (pt *PlPoint) AppendSubPoint(plpt PlInputPt) {
	pt.subPlpt = append(pt.subPlpt, plpt)
}

func (pt *PlPoint) GetSubPoint() []PlInputPt {
	return pt.subPlpt
}

func (pt *PlPoint) Category() point.Category {
	return pt.category
}

func (pt *PlPoint) Get(key string) (any, ast.DType, error) {
	if v, ok := pt.tags[key]; ok {
		return v, ast.String, nil
	}

	if v, ok := pt.fields[key]; ok {
		v, dtype := valueDtype(v)
		return v, dtype, nil
	}
	return nil, ast.Invalid, fmt.Errorf("unsupported pt key type")
}

func (pt *PlPoint) GetWithIsTag(key string) (any, bool, bool) {
	if v, ok := pt.tags[key]; ok {
		return v, true, true
	}

	if v, ok := pt.fields[key]; ok {
		v, _ := valueDtype(v)
		return v, false, true
	}
	return nil, false, false
}

func (pt *PlPoint) Set(key string, value any, dtype ast.DType) error {
	if _, ok := pt.tags[key]; ok { // is tag
		if dtype == ast.Void || dtype == ast.Invalid {
			delete(pt.tags, key)
			return nil
		}
		if v, err := plruntime.Conv2String(value, dtype); err == nil {
			pt.tags[key] = v
			return nil
		} else {
			return err
		}
	} else { // is field
		switch dtype { //nolint:exhaustive
		case ast.Nil, ast.Void, ast.Invalid:
			pt.fields[key] = nil
			return nil
		case ast.List, ast.Map:
			if v, err := plruntime.Conv2String(value, dtype); err == nil {
				pt.fields[key] = v
			} else {
				pt.fields[key] = nil
				return nil
			}
		default:
			pt.fields[key] = value
		}
	}
	return nil
}

func (pt *PlPoint) Delete(key string) {
	if _, ok := pt.tags[key]; ok {
		delete(pt.tags, key)
	} else {
		delete(pt.fields, key)
	}
}

func (pt *PlPoint) RenameKey(from, to string) error {
	if v, ok := pt.fields[from]; ok {
		pt.fields[to] = v
		delete(pt.fields, from)
	} else if v, ok := pt.tags[from]; ok {
		pt.tags[to] = v
		delete(pt.tags, from)
	} else {
		return fmt.Errorf("key(from) %s not found", from)
	}
	return nil
}

func (pt *PlPoint) SetTag(key string, value any, dtype ast.DType) error {
	delete(pt.fields, key)

	if str, err := plruntime.Conv2String(value, dtype); err == nil {
		pt.tags[key] = str
		return nil
	} else {
		pt.tags[key] = ""
		return err
	}
}

func (pt *PlPoint) PtTime() time.Time {
	return pt.time
}

func (pt *PlPoint) GetAggBuckets() *plmap.AggBuckets {
	return pt.aggBuckets
}

func (pt *PlPoint) SetAggBuckets(buks *plmap.AggBuckets) {
	pt.aggBuckets = buks
}

func (pt *PlPoint) SetPlReferTables(refTable refertable.PlReferTables) {
	pt.refTable = refTable
}

func (pt *PlPoint) GetPlReferTables() refertable.PlReferTables {
	return pt.refTable
}

func (pt *PlPoint) SetIPDB(db ipdb.IPdb) {
	pt.ipdb = db
}

func (pt *PlPoint) GetIPDB() ipdb.IPdb {
	return pt.ipdb
}

func (pt *PlPoint) KeyTime2Time() {
	if v, _, err := pt.Get("time"); err == nil {
		if nanots, ok := v.(int64); ok {
			t := time.Unix(nanots/int64(time.Second),
				nanots%int64(time.Second))
			if !t.IsZero() {
				pt.time = t
			}
		}
		pt.Delete("time")
	}
}

func (pt *PlPoint) MarkDrop(drop bool) {
	pt.drop = drop
}

func (pt *PlPoint) Dropped() bool {
	return pt.drop
}

func (pt *PlPoint) Tags() map[string]string {
	return pt.tags
}

func (pt *PlPoint) Fields() map[string]any {
	return pt.fields
}

func (pt *PlPoint) Point() *point.Point {
	opt := utils.PtCatOption(pt.category)
	opt = append(opt, point.WithTime(pt.PtTime()))

	fieldsKVS := point.NewTags(pt.tags)
	fieldsKVS = append(fieldsKVS, point.NewKVs(pt.fields)...)
	return point.NewPointV2(pt.name, fieldsKVS, opt...)
}

func WrapPoint(cat point.Category, pt *point.Point) PlInputPt {
	if pt == nil {
		return nil
	}

	return NewPlPoint(cat, pt.Name(),
		pt.MapTags(), pt.InfluxFields(), pt.Time())
}
