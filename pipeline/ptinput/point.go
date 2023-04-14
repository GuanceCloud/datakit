// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ptinput impl ppl input interface
package ptinput

import (
	"fmt"
	"sync"
	"time"

	"github.com/GuanceCloud/platypus/pkg/ast"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/plmap"

	"github.com/spf13/cast"
)

type (
	KeyKind uint
	PtFlag  uint
)

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

type Point struct {
	Name   string
	Tags   map[string]string
	Fields map[string]any // int, float, bool, string, map, slice, array
	Time   time.Time

	PtReadFrom *Point
	PtWriteTo  *Point

	AggBuckets *plmap.AggBuckets

	Category string
	Drop     bool
}

func InitPt(pt *Point, name string, t map[string]string, f map[string]any, tn time.Time) *Point {
	if f == nil {
		f = map[string]any{}
	}

	if t == nil {
		t = map[string]string{}
	}

	pt.Name = name
	pt.Tags = t
	pt.Fields = f
	pt.Time = tn
	pt.Drop = false

	return pt
}

func valueDtype(v any) (any, ast.DType) {
	switch v.(type) {
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
	case string:
		return v, ast.String
	}

	// ignore unknown type
	return nil, ast.Nil
}

func (pt *Point) Conv2Pt() (*point.Point, error) {
	pt.KeyTime2Time()

	opt := &point.PointOption{
		Category: pt.Category,
		Time:     pt.Time,
	}

	v, err := point.NewPoint(pt.Name, pt.Tags, pt.Fields, opt)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (pt *Point) Get(key string) (any, ast.DType, error) {
	if pt.PtReadFrom != nil {
		return pt.PtReadFrom.Get(key)
	}

	if v, ok := pt.Tags[key]; ok {
		return v, ast.String, nil
	}

	if v, ok := pt.Fields[key]; ok {
		v, dtype := valueDtype(v)
		return v, dtype, nil
	}
	return nil, ast.Invalid, fmt.Errorf("unsupported pt key type")
}

func (pt *Point) Delete(key string) {
	if pt.PtWriteTo != nil {
		pt.PtWriteTo.Delete(key)
		return
	}

	if _, ok := pt.Tags[key]; ok {
		delete(pt.Tags, key)
	} else {
		delete(pt.Fields, key)
	}
}

func (pt *Point) Set(key string, value any, dtype ast.DType) error {
	if pt.PtWriteTo != nil {
		return pt.PtWriteTo.Set(key, value, dtype)
	}

	if _, ok := pt.Tags[key]; ok {
		if dtype == ast.Void || dtype == ast.Invalid {
			delete(pt.Tags, key)
			return nil
		}
		if v, err := plruntime.Conv2String(value, dtype); err == nil {
			pt.Tags[key] = v
		} else {
			return err
		}
	}

	switch dtype { //nolint:exhaustive
	case ast.Nil, ast.Void, ast.Invalid:
		pt.Fields[key] = nil
		return nil
	case ast.List, ast.Map:
		if v, err := plruntime.Conv2String(value, dtype); err == nil {
			pt.Fields[key] = v
		} else {
			pt.Fields[key] = nil
			return nil
		}
	default:
		pt.Fields[key] = value
	}

	return nil
}

func (pt *Point) SetTag(key string, value any, dtype ast.DType) error {
	if pt.PtWriteTo != nil {
		return pt.PtWriteTo.SetTag(key, value, dtype)
	}

	delete(pt.Fields, key)

	if str, err := plruntime.Conv2String(value, dtype); err == nil {
		pt.Tags[key] = str
		return nil
	} else {
		pt.Tags[key] = ""
		return err
	}
}

func (pt *Point) Mv2Tag(key string) error {
	if pt.PtWriteTo != nil {
		return pt.PtWriteTo.Mv2Tag(key)
	}

	m, ok := pt.Fields[key]
	if ok {
		delete(pt.Fields, key)
	}

	v, dtype := valueDtype(m)

	if str, err := plruntime.Conv2String(v, dtype); err == nil {
		pt.Tags[key] = str
	} else {
		pt.Tags[key] = ""
		return err
	}

	return nil
}

func (pt *Point) SetMeasurement(m string) {
	if pt.PtWriteTo != nil {
		pt.PtWriteTo.SetMeasurement(m)
	}
	pt.Name = m
}

func (pt *Point) GetMeasurement() string {
	return pt.Name
}

func (pt *Point) KeyTime2Time() {
	if v, _, err := pt.Get("time"); err == nil {
		if nanots, ok := v.(int64); ok {
			t := time.Unix(nanots/int64(time.Second),
				nanots%int64(time.Second))
			if !t.IsZero() {
				pt.Time = t
			}
		}
		pt.Delete("time")
	}
}

var pointPool = sync.Pool{
	New: func() any {
		return &Point{}
	},
}

func GetPoint() *Point {
	pt, _ := pointPool.Get().(*Point)
	return pt
}

func PutPoint(pt *Point) {
	if pt == nil {
		return
	}

	pt.AggBuckets = nil

	pt.Name = ""

	pt.Fields = nil
	pt.Tags = nil

	pt.Drop = false

	pointPool.Put(pt)
}
