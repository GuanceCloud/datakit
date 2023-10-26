// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package point implements datakits basic data structure.
package point

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	Psent  = 1 << iota // The Point has been sent
	Ppb                // the point is Protobuf point
	Pcheck             // checked
	// more...
)

type Callback func(*Point) (*Point, error)

type Point struct {
	// warnnings and debug info about the point, for pbPoint,
	// they will wrapped in payload, but optional write to storage.
	warns  []*Warn
	debugs []*Debug
	keys   *Keys // bufferred keys

	// flags about the point
	flags uint64

	name string
	kvs  KVs
	time time.Time
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
		if x := FromLP(pt); x != nil {
			arr = append(arr, x)
		}
	}
	return
}

// NewLPPoint create Point based on a lineproto point.
func NewLPPoint(lp influxm.Point) *Point {
	return FromLP(lp)
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
	return influxm.NewPoint(p.name, p.InfluxTags(), p.InfluxFields(), p.time)
}

// InfluxFields convert fields to map structure.
func (p *Point) InfluxFields() map[string]any {
	return p.kvs.InfluxFields()
}

// InfluxTags convert tags to map structure.
func (p *Point) InfluxTags() influxm.Tags {
	return p.kvs.InfluxTags()
}

// MapTags convert all key-value to map.
func (p *Point) MapTags() map[string]string {
	res := map[string]string{}

	for _, kv := range p.kvs {
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

	for _, kv := range p.kvs {
		res[kv.Key] = kv.Raw()
	}

	return res
}

// Pretty get string representation of point, suffixed with all warning(if any)
// during build the point.
func (p *Point) Pretty() string {
	arr := []string{
		"\n",
		p.Name(),
		"-----------",
		p.kvs.Pretty(),
		"-----------",
		fmt.Sprintf("%s | %d",
			p.Time().String(),
			p.Time().UnixNano()),
	}

	if len(p.warns) > 0 {
		arr = append(arr, "-----------")
	}

	// only pbpoint attached with warns
	for _, w := range p.warns {
		arr = append(arr, fmt.Sprintf("[W] %s: %s", w.Type, w.Msg))
	}

	// only pbpoint attached with debugs
	for _, d := range p.debugs {
		arr = append(arr, fmt.Sprintf("[D] %s", d.Info))
	}

	return strings.Join(arr, "\n")
}

// Warns return warnning info when build the point.
func (p *Point) Warns() []*Warn {
	return p.warns
}

// WarnsPretty return human readable warnning info.
func (p *Point) WarnsPretty() string {
	var arr []string
	for _, w := range p.warns {
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
	if err := protojson.Unmarshal(j, &pbpt); err != nil {
		return nil, err
	}

	return FromPB(&pbpt), nil
}

func FromJSONPoint(j *JSONPoint) *Point {
	pt := &Point{
		name: j.Measurement,
		kvs:  NewKVs(j.Fields),
		time: time.Unix(0, j.Time),
	}

	for k, v := range j.Tags {
		pt.MustAddTag(k, v)
	}

	return pt
}

func FromLP(lp influxm.Point) *Point {
	lpfs, err := lp.Fields()
	if err != nil { // invalid line-protocol point
		return nil
	}

	pt := &Point{
		name: string(lp.Name()),
		kvs:  NewKVs(lpfs),
		time: lp.Time(),
	}

	for _, tag := range lp.Tags() {
		pt.MustAddTag(string(tag.Key), string(tag.Value))
	}

	return pt
}

func FromModelsLP(lp influxm.Point) *Point {
	lpfs, err := lp.Fields()
	if err != nil {
		return nil
	}

	pt := &Point{
		name: string(lp.Name()),
		kvs:  NewKVs(lpfs),
		time: lp.Time(),
	}

	tags := lp.Tags()
	for _, t := range tags {
		pt.MustAddTag(string(t.Key), string(t.Value))
	}

	return pt
}

func FromPB(pb *PBPoint) *Point {
	kvs := KVs(pb.Fields)
	sort.Sort(kvs)

	pt := &Point{
		name:   pb.Name,
		kvs:    kvs,
		time:   time.Unix(0, pb.Time),
		warns:  pb.Warns,
		debugs: pb.Debugs,
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
	return protojson.Marshal(pbpt)
}

func (p *Point) PBJsonPretty() ([]byte, error) {
	pbpt := p.PBPoint()

	mo := &protojson.MarshalOptions{Multiline: true, Indent: "  "}

	return mo.Marshal(pbpt)
}

// Tags return point's key-values except fields.
func (p *Point) Tags() (arr KVs) {
	return p.kvs.Tags()
}

// Fields return point's key-values except tags.
func (p *Point) Fields() (arr KVs) {
	return p.kvs.Fields()
}

// KVs return point's all key-values.
func (p *Point) KVs() (arr KVs) {
	return p.kvs
}

// AddKV add kv, if the key exist, do nothing.
func (p *Point) AddKV(kv *Field) {
	p.kvs = p.kvs.AddKV(kv, false)
}

// MustAddKV add kv, if the key exist, override it.
func (p *Point) MustAddKV(kv *Field) {
	p.kvs = p.kvs.AddKV(kv, true)
}

// Name return point's measurement name.
func (p *Point) Name() string {
	return p.name
}

// Time return point's time.
func (p *Point) Time() time.Time {
	return p.time
}

// Get get specific key from point.
func (p *Point) Get(k string) any {
	if kv := p.kvs.Get(k); kv != nil {
		return kv.Raw()
	}
	return nil
}

// GetTag get value of tag k.
// If key k not tag or k not eixst, return nil.
func (p *Point) GetTag(k string) string {
	return p.kvs.GetTag(k)
}

// MustAdd add specific key value to fields, if k exist, override it.
func (p *Point) MustAdd(k string, v any) {
	p.kvs = p.kvs.Add(k, v, false, true)
}

// Add add specific key value to fields, if k exist, do nothing.
func (p *Point) Add(k string, v any) {
	p.kvs = p.kvs.Add(k, v, false, false)
}

// MustAddTag add specific key value to fields, if k exist, override it.
func (p *Point) MustAddTag(k, v string) {
	p.kvs = p.kvs.Add(k, v, true, true)
}

// AddTag add specific key value to fields, if k exist, do nothing.
func (p *Point) AddTag(k, v string) {
	p.kvs = p.kvs.Add(k, v, true, false)
}

// Del delete specific key from tags/fields.
func (p *Point) Del(k string) {
	p.kvs = p.kvs.Del(k)
}

func (p *Point) AddDebug(d *Debug) {
	p.debugs = append(p.debugs, d)
}

// PBPoint create Point based on a protobuf point.
func (p *Point) PBPoint() *PBPoint {
	return &PBPoint{ // we have to create the pbpoint
		Name:   p.name,
		Fields: p.kvs,
		Time:   p.Time().UnixNano(),

		Warns:  p.warns,
		Debugs: p.debugs,
	}
}

// Keys get points all keys.
func (p *Point) Keys() *Keys {
	if p.keys == nil {
		res := &Keys{
			hash: uint64(0),
			arr:  p.kvs.Keys().arr,
		}

		sort.Sort(res)
		p.keys = res
	}

	return p.keys
}

// Size get underling data size in byte(exclude warning/debug info).
func (p *Point) Size() int {
	n := len(p.name)
	for _, kv := range p.kvs {
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
			a := kv.GetA()
			n += (len(a.TypeUrl) + len(a.Value))

		default:
			// ignored
		}
	}

	for _, w := range p.warns {
		n += (len(w.Type) + len(w.Msg))
	}

	for _, d := range p.debugs {
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
	d, err := proto.Marshal(pbpt)
	if err != nil {
		return 0
	}

	return len(d)
}

func b64(x []byte) string {
	return base64.StdEncoding.EncodeToString(x)
}
