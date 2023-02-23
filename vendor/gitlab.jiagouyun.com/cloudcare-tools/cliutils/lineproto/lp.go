// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package lineproto

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"time"
	"unsafe"

	"github.com/influxdata/influxdb1-client/models"
	lp "github.com/influxdata/line-protocol/v2/lineprotocol"
)

func unsafeBytesToString(runes []byte) string {
	return *(*string)(unsafe.Pointer(&runes)) //nolint:gosec
}

// GetSafeString get a copy of string s.
func GetSafeString(s string) string {
	if s == "" {
		return ""
	}
	return string(*(*[]byte)(unsafe.Pointer(&s))) //nolint:gosec
}

type Precision = lp.Precision

const (
	Nanosecond  = lp.Nanosecond
	Microsecond = lp.Microsecond
	Millisecond = lp.Millisecond
	Second      = lp.Second
)

// ConvertPrecisionToV2 map version 1 precision to version 2.
func ConvertPrecisionToV2(precision string) (Precision, error) {
	switch precision {
	case "u":
		return Microsecond, nil
	case "ms":
		return Millisecond, nil
	case "s":
		return Second, nil
	case "m":
		return 0, fmt.Errorf("line protocol v2 precision do not support minute")
	case "h":
		return 0, fmt.Errorf("line protocol v2 precision do not support hour")
	default:
		return Nanosecond, nil
	}
}

type Point struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	Time   time.Time
}

func NewPoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time,
) (*Point, error) {
	var tm time.Time
	if len(t) > 0 {
		tm = t[0]
	}
	return &Point{
		Name:   name,
		Tags:   tags,
		Fields: fields,
		Time:   tm,
	}, nil
}

func (p *Point) ToInfluxdbPoint(defaultTime time.Time) (models.Point, error) {
	tm := p.Time
	if tm.IsZero() {
		tm = defaultTime
	}
	return models.NewPoint(p.Name, models.NewTags(p.Tags), p.Fields, tm)
}

func (p *Point) AddTag(key, val string) {
	p.Tags[key] = val
}

func (p *Point) AddField(key string, val interface{}) error {
	// check the val value
	if _, ok := InterfaceToValue(val); !ok {
		return fmt.Errorf("unsupported line protocol field interface{}: %T, [%v]", val, val)
	}
	p.Fields[key] = val
	return nil
}

func (p *Point) String() (string, error) {
	encoder := NewLineEncoder()

	if err := encoder.AppendPoint(p); err != nil {
		return "", fmt.Errorf("encoder append point err: %w", err)
	}

	line, err := encoder.UnsafeStringWithoutLn()
	if err != nil {
		return "", fmt.Errorf("line protocol encoding err: %w", err)
	}
	return line, nil
}

func ParseWithOptionSetter(data []byte, optionSetters ...OptionSetter) ([]*Point, error) {
	option := DefaultOption
	if len(optionSetters) > 0 {
		for _, setter := range optionSetters {
			setter(option)
		}
	}
	return Parse(data, option)
}

func Parse(data []byte, opt *Option) ([]*Point, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	if opt == nil {
		opt = DefaultOption
	}

	if opt.MaxFields <= 0 {
		opt.MaxFields = 1024
	}

	if opt.MaxTags <= 0 {
		opt.MaxTags = 256
	}

	ptTime := opt.Time
	if opt.Time.IsZero() {
		ptTime = time.Now()
	}

	dec := lp.NewDecoderWithBytes(data)

	var pts []*Point

	for dec.Next() {
		pt := &Point{
			Tags:   make(map[string]string),
			Fields: make(map[string]interface{}),
		}

		m, err := dec.Measurement()
		if err != nil {
			return pts, fmt.Errorf("parse measurement err: %w", err)
		}

		pt.Name = unsafeBytesToString(m)

		for {
			k, v, err := dec.NextTag()
			if err != nil {
				return pts, fmt.Errorf("read next tag err: %w", err)
			}

			if k == nil {
				break
			}

			pt.Tags[unsafeBytesToString(k)] = unsafeBytesToString(v)
		}

		for {
			k, v, err := dec.NextField()
			if err != nil {
				return pts, fmt.Errorf("get next field err: %w", err)
			}

			if k == nil {
				break
			}

			pt.Fields[unsafeBytesToString(k)] = v.Interface()
		}

		t, err := dec.Time(opt.PrecisionV2, ptTime)
		if err != nil {
			return pts, err
		}

		pt.Time = t

		if opt.ExtraTags != nil {
			for k, v := range opt.ExtraTags {
				if _, ok := pt.Tags[k]; !ok {
					pt.AddTag(k, v)
				}
			}
		}

		if opt.CallbackV2 != nil {
			newPoint, err := opt.CallbackV2(pt)
			if err != nil {
				return nil, fmt.Errorf("call callbackv2 err: %w", err)
			}
			pt = newPoint
		}

		if pt == nil {
			return nil, fmt.Errorf("line point is empty")
		}

		if err := checkPointV2(pt, opt); err != nil {
			return nil, err
		}

		pts = append(pts, pt)
	}

	return pts, nil
}

func InterfaceToValue(x interface{}) (lp.Value, bool) {
	switch y := x.(type) {
	case int64, uint64, float64, bool, string, []byte:
		// do nothing
	case int:
		x = int64(y)
	case rune:
		x = int64(y)
	case int16:
		x = int64(y)
	case int8:
		x = int64(y)
	case uint:
		x = uint64(y)
	case uint32:
		x = int64(y)
	case uint16:
		x = int64(y)
	case byte:
		x = int64(y)
	case float32:
		x = float64(y)
	default:
		k := reflect.ValueOf(y).Kind()
		if k == reflect.Slice || k == reflect.Map || k == reflect.Array ||
			k == reflect.Struct || k == reflect.Ptr {
			x = fmt.Sprintf("%v", y)
		}
	}
	return lp.NewValue(x)
}

type LineEncoder struct {
	Encoder *lp.Encoder
	Opt     *Option
}

func NewLineEncoder(optionSetters ...OptionSetter) *LineEncoder {
	opt := DefaultOption
	if len(optionSetters) > 0 {
		for _, setter := range optionSetters {
			setter(opt)
		}
	}

	encoder := &lp.Encoder{}
	encoder.SetPrecision(opt.PrecisionV2)

	return &LineEncoder{
		Encoder: encoder,
		Opt:     opt,
	}
}

func (le *LineEncoder) EnableLax() {
	le.Encoder.SetLax(true)
}

func (le *LineEncoder) AppendPoint(pt *Point) error {
	le.Encoder.StartLine(pt.Name)

	maxLen := len(pt.Tags)
	if len(pt.Fields) > maxLen {
		maxLen = len(pt.Fields)
	}

	keys := make([]string, 0, maxLen)

	for key, val := range pt.Tags {
		if val == "" {
			continue
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for _, key := range keys {
		le.Encoder.AddTag(key, pt.Tags[key])
	}

	if len(pt.Fields) == 0 {
		return models.ErrPointMustHaveAField
	}
	keys = keys[:0]
	for key := range pt.Fields {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for _, key := range keys {
		value, ok := InterfaceToValue(pt.Fields[key])
		if !ok {
			return fmt.Errorf("unable parse value from interface{}: %T, [%v]", pt.Fields[key], pt.Fields[key])
		}
		le.Encoder.AddField(key, value)
	}

	le.Encoder.EndLine(pt.Time)

	return le.Encoder.Err()
}

// Bytes return the line protocol bytes
// You should be **VERY CAREFUL** when using this function together with the Reset.
func (le *LineEncoder) Bytes() ([]byte, error) {
	return le.Encoder.Bytes(), le.Encoder.Err()
}

// BytesWithoutLn return the line protocol bytes without the trailing new line
// You should be **VERY CAREFUL** when using this function together with the Reset.
func (le *LineEncoder) BytesWithoutLn() ([]byte, error) {
	return bytes.TrimRightFunc(le.Encoder.Bytes(), func(r rune) bool {
		return r == '\r' || r == '\n'
	}), le.Encoder.Err()
}

// UnsafeString return string with no extra allocation
// You should be **VERY CAREFUL** when using this function together with the Reset.
func (le *LineEncoder) UnsafeString() (string, error) {
	lineBytes, err := le.Bytes()
	if len(lineBytes) > 0 {
		return *(*string)(unsafe.Pointer(&lineBytes)), err //nolint:gosec
	}
	return "", err
}

// UnsafeStringWithoutLn return the line protocol **UNSAFE** string without the trailing new line
// You should be **VERY CAREFUL** when using this function together with the Reset.
func (le *LineEncoder) UnsafeStringWithoutLn() (string, error) {
	lineBytes, err := le.BytesWithoutLn()
	if len(lineBytes) > 0 {
		return *(*string)(unsafe.Pointer(&lineBytes)), err //nolint:gosec
	}
	return "", err
}

func (le *LineEncoder) Reset() {
	le.Encoder.Reset()
}

func (le *LineEncoder) SetBuffer(buf []byte) {
	le.Encoder.SetBuffer(buf)
}

func Encode(pts []*Point, opt ...OptionSetter) ([]byte, error) {
	enc := NewLineEncoder(opt...)

	for _, pt := range pts {
		err := enc.AppendPoint(pt)
		if err != nil {
			return nil, fmt.Errorf("encoder append point err: %w", err)
		}
	}

	return enc.Bytes()
}
