// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package point implements datakits basic data structure.
package point

import (
	bytes "bytes"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"

	protojson "github.com/gogo/protobuf/jsonpb"
	influxm "github.com/influxdata/influxdb1-client/models"
)

const (
	Psent   = 1 << iota // The Point has been sent
	Ppb                 // the point is Protobuf point
	Pcheck              // checked
	Ppooled             // from point pool
	// more...
)

type Callback func(*Point) (*Point, error)

type Point struct {
	flags uint64
	pt    *PBPoint
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
func WrapPoint(pts []influxm.Point) (arr []*Point) {
	for _, pt := range pts {
		if x := FromModelsLP(pt); x != nil {
			arr = append(arr, x)
		}
	}
	return
}

// NewLPPoint create Point based on a lineproto point.
func NewLPPoint(lp influxm.Point) *Point {
	return FromModelsLP(lp)
}

func (p *Point) MustLPPoint() influxm.Point {
	lppt, err := p.LPPoint()
	if err != nil {
		panic(err.Error())
	}

	return lppt
}

// LPPoint get line-protocol part of the point.
func (p *Point) LPPoint() (influxm.Point, error) {
	return influxm.NewPoint(p.pt.Name, p.InfluxTags(), p.InfluxFields(), time.Unix(0, p.pt.Time))
}

// InfluxFields convert fields to map structure.
func (p *Point) InfluxFields() map[string]any {
	kvs := KVs(p.pt.Fields)
	return kvs.InfluxFields()
}

// InfluxTags convert tags to map structure.
func (p *Point) InfluxTags() influxm.Tags {
	kvs := KVs(p.pt.Fields)
	return kvs.InfluxTags()
}

// MapTags convert all key-value to map.
func (p *Point) MapTags() map[string]string {
	res := map[string]string{}

	for _, kv := range p.pt.Fields {
		if !kv.IsTag {
			continue
		}

		res[kv.Key] = kv.GetS()
	}

	return res
}

// KVMap return all key-value in map.
func (p *Point) KVMap() map[string]any {
	res := map[string]any{}

	for _, kv := range p.pt.Fields {
		res[kv.Key] = kv.Raw()
	}

	return res
}

// Pretty get string representation of point, all key-valus will be sorted in output.
func (p *Point) Pretty() string {
	kvs := KVs(p.pt.Fields)

	arr := []string{
		"\n",
		p.Name(),
		"-----------",
		kvs.Pretty(),
		"-----------",
		fmt.Sprintf("%s | %d",
			p.Time().String(),
			p.Time().UnixNano()),
	}

	if len(p.pt.Warns) > 0 {
		arr = append(arr, "-----------")
	}

	// only pbpoint attached with warns
	for _, w := range p.pt.Warns {
		arr = append(arr, fmt.Sprintf("[W] %s: %s", w.Type, w.Msg))
	}

	// only pbpoint attached with debugs
	for _, d := range p.pt.Debugs {
		arr = append(arr, fmt.Sprintf("[D] %s", d.Info))
	}

	return strings.Join(arr, "\n")
}

// Warns return warnning info when build the point.
func (p *Point) Warns() []*Warn {
	return p.pt.Warns
}

// WarnsPretty return human readable warnning info.
func (p *Point) WarnsPretty() string {
	var arr []string
	for _, w := range p.pt.Warns {
		arr = append(arr, w.String())
	}
	return strings.Join(arr, "\n")
}

// makeLineproto build lineproto from @p's raw data(name/tag/field/time).
func (p *Point) makeLineproto(prec ...Precision) string {
	lp, err := p.LPPoint()
	if err != nil {
		return ""
	}

	if len(prec) == 0 {
		return lp.String()
	} else {
		return lp.PrecisionString(prec[0].String())
	}
}

func MustFromPBJson(j []byte) *Point {
	if pts, err := FromPBJson(j); err != nil {
		panic(err.Error())
	} else {
		return pts
	}
}

func FromPBJson(j []byte) (*Point, error) {
	var pbpt PBPoint
	m := &protojson.Unmarshaler{}
	buf := bytes.NewBuffer(j)
	if err := m.Unmarshal(buf, &pbpt); err != nil {
		return nil, err
	}

	return FromPB(&pbpt), nil
}

func FromJSONPoint(j *JSONPoint) *Point {
	kvs := NewKVs(j.Fields)

	for k, v := range j.Tags {
		kvs.MustAddTag(k, v)
	}
	return NewPointV2(j.Measurement, kvs, WithTime(time.Unix(0, j.Time)))

}

func FromModelsLP(lp influxm.Point) *Point {
	lpfs, err := lp.Fields()
	if err != nil {
		return nil
	}

	var kvs KVs
	kvs = NewKVs(lpfs)

	tags := lp.Tags()
	for _, t := range tags {
		kvs = kvs.MustAddTag(string(t.Key), string(t.Value))
	}

	return NewPointV2(string(lp.Name()), kvs, WithTime(lp.Time()))
}

func FromPB(pb *PBPoint) *Point {
	pt := NewPointV2(pb.Name, pb.Fields, WithTime(time.Unix(0, pb.Time)))
	if len(pb.Warns) > 0 {
		pt.pt.Warns = pb.Warns
	}

	if len(pb.Debugs) > 0 {
		pt.pt.Debugs = pb.Debugs
	}

	pt.SetFlag(Ppb)
	return pt
}

// LineProto convert point to text lineprotocol(both for
// lineproto point and protobuf point).
func (p *Point) LineProto(prec ...Precision) string {
	return p.makeLineproto(prec...)
}

func (p *Point) PBJson() ([]byte, error) {
	pbpt := p.PBPoint()
	m := protojson.Marshaler{}
	buf := &bytes.Buffer{}
	if err := m.Marshal(buf, pbpt); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Point) PBJsonPretty() ([]byte, error) {
	pbpt := p.PBPoint()

	m := &protojson.Marshaler{Indent: "  "}
	buf := &bytes.Buffer{}
	if err := m.Marshal(buf, pbpt); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Tags return point's key-values except fields.
func (p *Point) Tags() (arr KVs) {
	kvs := KVs(p.pt.Fields)
	return kvs.Tags()
}

// Fields return point's key-values except tags.
func (p *Point) Fields() (arr KVs) {
	kvs := KVs(p.pt.Fields)
	return kvs.Fields()
}

// KVs return point's all key-values.
func (p *Point) KVs() (arr KVs) {
	return KVs(p.pt.Fields)
}

// AddKVs add(shallow-copy) kvs, if keys exist, do nothing.
func (p *Point) AddKVs(kvs ...*Field) {
	old := KVs(p.pt.Fields)
	for _, kv := range kvs {
		old = old.AddKV(kv, false)
	}
	p.pt.Fields = old
}

// CopyTags deep-copy tag kvs, if keys exist, do nothing.
func (p *Point) CopyTags(kvs ...*Field) {
	old := KVs(p.pt.Fields)
	for _, kv := range kvs {
		old = old.AddTag(kv.Key, kv.GetS())
	}
	p.pt.Fields = old
}

// CopyField deep-copy field kvs, if keys exist, do nothing.
func (p *Point) CopyField(kvs ...*Field) {
	old := KVs(p.pt.Fields)
	for _, kv := range kvs {
		old = old.Add(kv.Key, kv.Raw(), false, false)
	}
	p.pt.Fields = old
}

// SetKVs set kvs as p's new KVs.
func (p *Point) SetKVs(kvs ...*Field) {
	old := KVs(p.pt.Fields)

	for _, kv := range kvs {
		old = old.AddKV(kv, false)
	}
	p.pt.Fields = old
}

// MustAddKVs add kv, if the key exist, override it.
func (p *Point) MustAddKVs(kvs ...*Field) {
	old := KVs(p.pt.Fields)
	for _, kv := range kvs {
		old = old.AddKV(kv, true)
	}
	p.pt.Fields = old
}

// Name return point's measurement name.
func (p *Point) Name() string {
	return p.pt.Name
}

// Time return point's time.
func (p *Point) Time() time.Time {
	return time.Unix(0, p.pt.Time)
}

// Get get specific key from point.
func (p *Point) Get(k string) any {
	kvs := KVs(p.pt.Fields)

	if kv := kvs.Get(k); kv != nil {
		return kv.Raw()
	}
	return nil
}

// GetTag get value of tag k.
// If key k not tag or k not eixst, return nil.
func (p *Point) GetTag(k string) string {
	kvs := KVs(p.pt.Fields)
	return kvs.GetTag(k)
}

// MustAdd add specific key value to fields, if k exist, override it.
func (p *Point) MustAdd(k string, v any) {
	kvs := KVs(p.pt.Fields)
	kvs = kvs.Add(k, v, false, true)
	p.pt.Fields = kvs
}

// Add add specific key value to fields, if k exist, do nothing.
func (p *Point) Add(k string, v any) {
	kvs := KVs(p.pt.Fields)
	p.pt.Fields = kvs.Add(k, v, false, false)
}

// MustAddTag add specific key value to fields, if k exist, override it.
func (p *Point) MustAddTag(k, v string) {
	kvs := KVs(p.pt.Fields)
	p.pt.Fields = kvs.Add(k, v, true, true)
}

// AddTag add specific key value to fields, if k exist, do nothing.
func (p *Point) AddTag(k, v string) {
	kvs := KVs(p.pt.Fields)
	p.pt.Fields = kvs.Add(k, v, true, false)
}

// Del delete specific key from tags/fields.
func (p *Point) Del(k string) {
	kvs := KVs(p.pt.Fields)
	p.pt.Fields = kvs.Del(k)
}

func (p *Point) AddDebug(d *Debug) {
	p.pt.Debugs = append(p.pt.Debugs, d)
}

// PBPoint create Point based on a protobuf point.
func (p *Point) PBPoint() *PBPoint {
	return p.pt
}

// Keys get points all keys.
func (p *Point) Keys() *Keys {
	kvs := KVs(p.pt.Fields)

	res := &Keys{
		hash: uint64(0),
		arr:  kvs.Keys().arr,
	}

	sort.Sort(res)

	return res
}

// Size get underling data size in byte(exclude warning/debug info).
func (p *Point) Size() int {
	n := len(p.pt.Name)
	for _, kv := range p.pt.Fields {
		n += len(kv.Key)
		n += 1 // IsTag
		n += 8 // time
		n += 4 // MetricType: uint32
		n += len(kv.Unit)

		switch kv.Val.(type) {
		case *Field_I,
			*Field_F,
			*Field_U:
			n += 8

		case *Field_B:
			n += 1

		case *Field_D:
			n += len(kv.GetD())

		case *Field_S:
			n += len(kv.GetS())

		case *Field_A:
			if a := kv.GetA(); a != nil {
				n += (len(a.TypeUrl) + len(a.Value))
			}

		default:
			// ignored
		}
	}

	for _, w := range p.pt.Warns {
		n += (len(w.Type) + len(w.Msg))
	}

	for _, d := range p.pt.Debugs {
		n += (len(d.Info))
	}

	return n
}

// LPSize get point line-protocol size.
func (p *Point) LPSize() int {
	lppt, err := p.LPPoint()
	if err != nil {
		return 0
	}

	return len(lppt.String())
}

// PBSize get point protobuf size.
func (p *Point) PBSize() int {
	pbpt := p.PBPoint()

	m := protojson.Marshaler{}
	buf := bytes.Buffer{}

	if err := m.Marshal(&buf, pbpt); err != nil {
		return 0
	}

	return buf.Len()
}

func b64(x []byte) string {
	return base64.StdEncoding.EncodeToString(x)
}
