// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package point implements datakits basic data structure.
package point

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

const (
	Psent  = 1 << iota // The Point has been sent
	Ppb                // the point is Protobuf point
	Pcheck             // checked
	// more...
)

type Callback func(*Point) (*Point, error)

type Point struct {
	lpPoint *influxdb.Point
	pbPoint *PBPoint

	// warnnings and debug info about the point, for pbPoint,
	// they will wrapped in payload, but optional write to storage.
	warns  []*Warn
	debugs []*Debug

	// flags about the point
	flags uint64

	name   []byte
	tags   Tags
	fields Fields
	time   time.Time
}

// ClearFlag clear specific bit.
func (p *Point) ClearFlag(f uint64) {
	mask := ^(uint64(1) << f)
	p.flags &= mask
}

// SetFlag set specific bit.
func (p *Point) SetFlag(f uint) {
	p.flags |= (1 << f)
}

// HasFlag test if specific bit set.
func (p *Point) HasFlag(f uint) bool {
	return (p.flags & (1 << f)) > 0
}

// WrapPoint wrap lagacy line-protocol point into Point.
func WrapPoint(pts []*influxdb.Point) (x []*Point) {
	for _, pt := range pts {
		x = append(x, &Point{lpPoint: pt})
	}
	return
}

// NewLPPoint create Point based on a lineproto point.
func NewLPPoint(pt *influxdb.Point) *Point {
	return &Point{lpPoint: pt}
}

// LPPoint get line-protocol part of the point.
func (p *Point) LPPoint() *influxdb.Point {
	return p.lpPoint
}

// String return raw string for line-protocol point, or JSON format for protobuf point.
func (p *Point) String() string {
	if p.HasFlag(Ppb) {
		if p.pbPoint == nil {
			if err := p.buildPBPoint(); err != nil {
				return ""
			}
		}

		return p.pbPoint.String()
	} else {
		if x, err := p.LineProto(NS); err != nil {
			return ""
		} else {
			return x
		}
	}
}

// StringWithWarn get string representation of point, suffixed with all warning(if any)
// during build the point.
func (p *Point) StringWithWarn() string {
	arr := []string{p.String()}

	// only pbpoint attached with warns
	if p.pbPoint != nil && len(p.pbPoint.Warns) > 0 {
		p.warns = p.pbPoint.Warns
	}

	for _, w := range p.warns {
		arr = append(arr, fmt.Sprintf("[W] %s: %s", w.Type, w.Msg))
	}

	return strings.Join(arr, "\n")
}

func (p *Point) StringWithDebug() string {
	arr := []string{p.String()}

	// only pbpoint attached with debugs
	if p.pbPoint != nil && len(p.pbPoint.Warns) > 0 {
		p.debugs = p.pbPoint.Debugs
	}

	for _, d := range p.debugs {
		arr = append(arr, fmt.Sprintf("[D] %s", d.Info))
	}

	return strings.Join(arr, "\n")
}

// toLineProto build lineproto from @p's lpPoint.
func (p *Point) toLineProto(prec ...Precision) (string, error) {
	if len(prec) == 0 {
		return p.lpPoint.String(), nil
	} else {
		return p.lpPoint.PrecisionString(prec[0].String()), nil
	}
}

// makeLineproto build lineproto from @p's raw data(name/tag/field/time).
func (p *Point) makeLineproto(prec ...Precision) (string, error) {
	fs := p.Fields()

	lppt, err := influxdb.NewPoint(string(p.name), p.Tags().InfluxTags(), fs.InfluxFields(), p.time)
	if err != nil {
		return "", err
	}

	if len(prec) == 0 {
		return lppt.String(), nil
	} else {
		return lppt.PrecisionString(prec[0].String()), nil
	}
}

func (p *Point) MustLineProto(prec ...Precision) string {
	if x, err := p.LineProto(prec...); err != nil {
		panic(err.Error())
	} else {
		return x
	}
}

// LineProto convert point to text lineprotocol(both for
// lineproto point and protobuf point).
func (p *Point) LineProto(prec ...Precision) (string, error) {
	if !p.HasFlag(Ppb) {
		if p.lpPoint == nil {
			return p.makeLineproto(prec...)
		} else {
			return p.toLineProto(prec...)
		}
	}

	if p.pbPoint == nil {
		return p.makeLineproto(prec...)
	} else { // feedback raw data to @p
		p.tags = p.Tags()
		p.fields = p.Fields()
		if len(p.fields) == 0 {
			return "", ErrNoFields
		}
		p.time = p.Time()
		p.name = p.Name()
		return p.makeLineproto(prec...)
	}
}

// Tags return point's all tag.
func (p *Point) Tags() Tags {
	if p.HasFlag(Ppb) {
		if p.pbPoint == nil {
			return p.tags
		}

		return p.pbPoint.Tags
	} else {
		if p.lpPoint == nil {
			return p.tags
		}

		tags := p.lpPoint.Tags()
		for k, v := range tags {
			p.tags = append(p.tags, &Tag{Key: []byte(k), Val: []byte(v)})
		}

		return p.tags
	}
}

// Name return point's measurement name.
func (p *Point) Name() []byte {
	if !p.HasFlag(Ppb) {
		if p.lpPoint == nil {
			return p.name
		}
		return []byte(p.lpPoint.Name())
	} else {
		if p.pbPoint == nil {
			return p.name
		}
		return p.pbPoint.GetName()
	}
}

// Time return point's time.
func (p *Point) Time() time.Time {
	if !p.HasFlag(Ppb) {
		if p.lpPoint == nil {
			return p.time
		}
		return p.lpPoint.Time()
	}

	if p.pbPoint == nil {
		return p.time
	} else {
		return time.Unix(0, p.pbPoint.Time)
	}
}

// Fields return point's all fields.
func (p *Point) Fields() Fields {
	if !p.HasFlag(Ppb) {
		if p.lpPoint == nil {
			return p.fields
		} else {
			fs, err := p.lpPoint.Fields()
			if err != nil {
				return nil
			}

			return PBFields(fs)
		}
	}

	if p.pbPoint == nil {
		return p.fields
	}

	return p.pbPoint.Fields
}

// Key get specific key from point.
// The returned value comes from tags or fields.
func (p *Point) Key(k []byte) any {
	tags := p.Tags()
	if v := tags.Key(k); v != nil {
		return v
	}

	fs := p.Fields()
	if v := fs.Key(k); v != nil {
		switch x := v.(type) {
		case *Field_I:
			return x.I
		case *Field_U:
			return x.U
		case *Field_S:
			return x.S
		case *Field_F:
			return x.F
		case *Field_F32:
			return x.F32
		case *Field_B:
			return x.B
		case *Field_D:
			return x.D

		default:
			return nil
		}
	}
	return nil
}

func (p *Point) AddDebug(d *Debug) {
	p.debugs = append(p.debugs, d)
}

// buildPBPoint create Point based on a protobuf point.
func (p *Point) buildPBPoint() error {
	fs := p.Fields()

	p.pbPoint = &PBPoint{ // we have to create the pbpoint
		Name:   p.Name(),
		Tags:   p.Tags(),
		Fields: fs,
		Time:   p.Time().UnixNano(),

		Warns:  p.warns,
		Debugs: p.debugs,
	}
	return nil
}

func (p *Point) buildLPPoint() error {
	fields := p.Fields()

	lppt, err := influxdb.NewPoint(string(p.Name()), p.Tags().InfluxTags(), fields.InfluxFields(), p.Time())
	if err != nil {
		return err
	}
	p.lpPoint = lppt
	return nil
}

// PBTags create tags slice from map structure.
func PBTags(tags map[string]string) (res Tags) {
	for k, v := range tags {
		res = append(res, &Tag{Key: []byte(k), Val: []byte(v)})
	}
	return res
}

// PBField get single field from specified key and value.
func PBField(k []byte, v any) *Field {
	switch x := v.(type) {
	case int8:
		return &Field{Key: k, Val: &Field_I{int64(x)}}
	case uint8:
		return &Field{Key: k, Val: &Field_I{int64(x)}}
	case int16:
		return &Field{Key: k, Val: &Field_I{int64(x)}}
	case uint16:
		return &Field{Key: k, Val: &Field_I{int64(x)}}
	case int32:
		return &Field{Key: k, Val: &Field_I{int64(x)}}
	case uint32:
		return &Field{Key: k, Val: &Field_I{int64(x)}}
	case int:
		return &Field{Key: k, Val: &Field_I{int64(x)}}
	case uint:
		return &Field{Key: k, Val: &Field_I{int64(x)}}
	case int64:
		return &Field{Key: k, Val: &Field_I{x}}
	case uint64:
		return &Field{Key: k, Val: &Field_U{x}}
	case string:
		return &Field{Key: k, Val: &Field_S{x}}

	case float64:
		return &Field{Key: k, Val: &Field_F{x}}

	case float32:
		return &Field{Key: k, Val: &Field_F32{x}}

	case []byte:
		return &Field{Key: k, Val: &Field_D{x}}

	case bool:
		return &Field{Key: k, Val: &Field_B{x}}

	case *anypb.Any:
		return &Field{Key: k, Val: &Field_A{x}}

	default: // value ignored
		return &Field{Key: k, Val: nil}
	}
}

// PBFields create fields slice from map structure.
func PBFields(fields map[string]interface{}) (res Fields) {
	for k, v := range fields {
		res = append(res, PBField([]byte(k), v))
	}
	return res
}

type Tags []*Tag

func (x Tags) Len() int {
	return len(x)
}

func (x Tags) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x Tags) Less(i, j int) bool {
	return bytes.Compare(x[i].Key, x[j].Key) < 0 // stable sort
}

// InfluxTags convert tags to map structure.
func (x Tags) InfluxTags() map[string]string {
	res := map[string]string{}
	for _, t := range x {
		res[string(t.Key)] = string(t.Val)
	}

	return res
}

// KeyExist test if key k exist.
func (x Tags) KeyExist(k []byte) bool {
	for _, t := range x {
		if bytes.Equal(t.Key, k) {
			return true
		}
	}

	return false
}

// Key get k's value, if k not exist, return nil.
func (x Tags) Key(k []byte) []byte {
	for _, t := range x {
		if bytes.Equal(t.Key, k) {
			return t.Val
		}
	}

	return nil
}

func (x Tags) Pretty() string {
	var arr []string
	for _, t := range x {
		arr = append(arr, t.String())
	}

	return strings.Join(arr, "\n")
}

type Fields []*Field

func (x Fields) Len() int {
	return len(x)
}

func (x Fields) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x Fields) Less(i, j int) bool {
	return bytes.Compare(x[i].Key, x[j].Key) < 0 // stable sort
}

func (x Fields) Pretty() string {
	var arr []string
	for _, f := range x {
		arr = append(arr, f.String())
	}

	return strings.Join(arr, "\n")
}

// InfluxFields convert fields to map structure.
func (x Fields) InfluxFields() map[string]any {
	res := map[string]any{}

	for _, f := range x {
		switch x := f.Val.(type) {
		case *Field_I:
			res[string(f.Key)] = x.I
		case *Field_U:
			res[string(f.Key)] = x.U
		case *Field_S:
			res[string(f.Key)] = x.S
		case *Field_F:
			res[string(f.Key)] = x.F
		case *Field_F32:
			res[string(f.Key)] = x.F32
		case *Field_B:
			res[string(f.Key)] = x.B
		case *Field_D:
			res[string(f.Key)] = b64(x.D)
		default:
			continue
		}
	}

	return res
}

// KeyExist test if k exist in fields.
func (x Fields) KeyExist(k []byte) bool {
	for _, f := range x {
		if bytes.Equal(f.Key, k) {
			return true
		}
	}

	return false
}

// Key get k's value, if k not exist, return nil.
func (x Fields) Key(k []byte) isField_Val {
	for _, f := range x {
		if bytes.Equal(f.Key, k) {
			return f.Val
		}
	}

	return nil
}

func b64(x []byte) string {
	return base64.StdEncoding.EncodeToString(x)
}
