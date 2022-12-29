// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package input

import (
	"fmt"
	"sync"
	"time"

	"github.com/GuanceCloud/platypus/pkg/ast"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"

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

type TFMeta struct {
	DType  ast.DType
	PtFlag PtFlag
}

type InputWithVarbMapRW interface {
	Get(key string) (any, ast.DType, bool)
	Set(key string, value any, dtype ast.DType) bool
	Delete(key string) bool
}

type InputWithVarbMapR interface {
	Get(key string) (any, ast.DType, bool)
}

type InputWithoutVarbMap interface{}

type Point struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]any // int, float, bool, string, map, slice, array
	Time        time.Time

	Drop bool
	Meta map[string]*TFMeta // [DType, PtFlag]
}

func InitPt(pt *Point, m string, t map[string]string, f map[string]any, tn time.Time) *Point {
	if f == nil {
		f = map[string]any{}
	}

	if t == nil {
		t = map[string]string{}
	}

	pt.Measurement = m
	pt.Tags = t
	pt.Fields = f
	pt.Time = tn
	pt.Drop = false
	pt.Meta = map[string]*TFMeta{}

	for k, v := range f {
		if v == nil {
			pt.Meta[k] = GetMeta(ast.Nil, PtField)
			continue
		}
		switch v.(type) {
		case int32, int8, int16, int,
			uint, uint16, uint32, uint64, uint8:
			pt.Meta[k] = GetMeta(ast.Int, PtField)
			f[k] = cast.ToInt64(v)
		case int64:
			pt.Meta[k] = GetMeta(ast.Int, PtField)
		case float32:
			pt.Meta[k] = GetMeta(ast.Float, PtField)
			f[k] = cast.ToFloat64(v)
		case float64:
			pt.Meta[k] = GetMeta(ast.Float, PtField)
		case bool:
			pt.Meta[k] = GetMeta(ast.Bool, PtField)
		case string:
			pt.Meta[k] = GetMeta(ast.String, PtField)
		}
	}

	for k := range t {
		pt.Meta[k] = GetMeta(ast.String, PtTag)
	}
	return pt
}

func (pt *Point) Get(key string) (any, ast.DType, error) {
	m, ok := pt.Meta[key]

	if !ok {
		return nil, ast.Invalid, fmt.Errorf("not found")
	}

	if m.DType == ast.Void || m.DType == ast.Nil {
		return nil, ast.Nil, nil
	}

	switch m.PtFlag { //nolint:exhaustive
	case PtField:
		if v, ok := pt.Fields[key]; ok {
			return v, m.DType, nil
		}
		return nil, ast.Nil, nil
	case PtTag:
		if v, ok := pt.Tags[key]; ok {
			return v, ast.String, nil
		}
		return nil, ast.Nil, nil
	default:
		return nil, ast.Invalid, fmt.Errorf("unsupported pt key type")
	}
}

func (pt *Point) Delete(key string) {
	m, ok := pt.Meta[key]
	if !ok {
		return
	}

	if m.PtFlag == PtField {
		delete(pt.Fields, key)
	} else {
		delete(pt.Tags, key)
	}

	delete(pt.Meta, key)
	PutMeta(m)
}

func (pt *Point) Set(key string, value any, dtype ast.DType) error {
	m, ok := pt.Meta[key]

	if !ok {
		m = GetMeta(dtype, PtField)
		pt.Meta[key] = m
	}

	// key 值为 nil 只记录 meta 中
	switch m.PtFlag { //nolint:exhaustive
	case PtField:
		switch dtype { //nolint:exhaustive
		case ast.Nil, ast.Void, ast.Invalid:
			m.DType = ast.Nil
			pt.Fields[key] = nil
			return nil
		case ast.List, ast.Map:
			if v, err := plruntime.Conv2String(value, dtype); err == nil {
				pt.Fields[key] = v
				m.DType = ast.String
			} else {
				m.DType = ast.Nil
				pt.Fields[key] = nil
				return nil
			}
		default:
			pt.Fields[key] = value
			m.DType = dtype
		}
	case PtTag:
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
	return nil
}

func (pt *Point) SetTag(key string, value any, dtype ast.DType) error {
	m, ok := pt.Meta[key]

	if !ok {
		m = GetMeta(ast.String, PtTag)
		pt.Meta[key] = m
	}

	if m.PtFlag == PtField {
		delete(pt.Fields, key)
		m.DType, m.PtFlag = ast.String, PtTag
	}

	if str, err := plruntime.Conv2String(value, dtype); err == nil {
		pt.Tags[key] = str
		return nil
	} else {
		pt.Tags[key] = ""
		return err
	}
}

func (pt *Point) Mv2Tag(key string) error {
	m, ok := pt.Meta[key]
	if !ok {
		pt.Meta[key] = GetMeta(ast.String, PtTag)
		pt.Tags[key] = ""
		return nil
	}

	if m.PtFlag == PtField {
		dType := m.DType

		m.DType, m.PtFlag = ast.String, PtTag
		v, ok := pt.Fields[key]
		if !ok {
			return nil
		}
		delete(pt.Fields, key)
		if str, err := plruntime.Conv2String(v, dType); err == nil {
			pt.Tags[key] = str
		} else {
			pt.Tags[key] = ""
			return err
		}
	}

	return nil
}

func (pt *Point) SetMeasurement(m string) {
	pt.Measurement = m
}

func (pt *Point) GetMeasurement() string {
	return pt.Measurement
}

func (pt *Point) KeyTime2Time() {
	if v, _, err := pt.Get("time"); err == nil {
		if nanots, ok := v.(int64); ok {
			t := time.Unix(nanots/int64(time.Second),
				nanots%int64(time.Second))
			if !t.IsZero() {
				pt.Time = t
			}
			pt.Delete("time")
		}
	}
}

var pointPool = sync.Pool{
	New: func() any {
		return &Point{}
	},
}

var metaPool = sync.Pool{
	New: func() any {
		return &TFMeta{}
	},
}

func GetMeta(dtype ast.DType, ptflag PtFlag) *TFMeta {
	meta, _ := metaPool.Get().(*TFMeta)
	meta.DType = dtype
	meta.PtFlag = ptflag
	return meta
}

func PutMeta(meta *TFMeta) {
	metaPool.Put(meta)
}

func GetPoint() *Point {
	pt, _ := pointPool.Get().(*Point)
	return pt
}

func PutPoint(pt *Point) {
	if pt == nil {
		return
	}

	pt.Measurement = ""

	pt.Fields = nil
	pt.Tags = nil

	for k, v := range pt.Meta {
		delete(pt.Meta, k)
		PutMeta(v)
	}

	pt.Drop = false

	pointPool.Put(pt)
}
