// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"fmt"
	"reflect"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

type PointCheckOption func(*ptChecker)

type ptChecker struct {
	checkValues,
	checkTypes bool

	measurementCheckIgnored bool

	expect *point.Point

	// expect measurement info
	doc   Measurement
	mInfo *MeasurementInfo

	extraTags map[string]string

	optionalFields []string
	optionalTags   []string

	// check result
	checkMsg []string

	// the checking point's info
	expName,
	gotName string

	expTags,
	gotTags point.KVs

	expFields,
	gotFields point.KVs

	expTime,
	gotTime time.Time
}

func newPointChecker() *ptChecker {
	return &ptChecker{
		checkTypes: true,
	}
}

// WithMeasurementCheckIgnored set flag to not check measurement name.
// Some inputs's measurement name are user defined(with a default name `default`).
func WithMeasurementCheckIgnored(on bool) PointCheckOption {
	return func(c *ptChecker) {
		c.measurementCheckIgnored = on
	}
}

// WithDoc set the point's document to check consistency between doc and the real data.
func WithDoc(m Measurement) PointCheckOption {
	return func(c *ptChecker) {
		c.doc = m
		c.mInfo = m.Info()
	}
}

// WithExtraTags inject additional tags to check point.
func WithExtraTags(tags map[string]string) PointCheckOption {
	return func(c *ptChecker) { c.extraTags = tags }
}

// WithExpectPoint set expected point to check point.
func WithExpectPoint(pt *point.Point) PointCheckOption {
	return func(c *ptChecker) {
		c.expect = pt
		c.expName = string(pt.Name())
		c.expTags = pt.Tags()
		c.expFields = pt.Fields()
		c.expTime = pt.Time()
	}
}

// WithValueChecking will check point's tag/field value according to expected point.
func WithValueChecking(on bool) PointCheckOption {
	return func(c *ptChecker) { c.checkValues = on }
}

// WithTypeChecking will check point's all fields value type according to document.
func WithTypeChecking(on bool) PointCheckOption {
	return func(c *ptChecker) { c.checkTypes = on }
}

// WithOptionalFields set optional keys(field key) that will escape key-check.
// Sometimes the point's field keys are optional for different configures.
func WithOptionalFields(keys ...string) PointCheckOption {
	return func(c *ptChecker) {
		c.optionalFields = append(c.optionalFields, keys...)
	}
}

// WithOptionalTags set optional keys(tag key) that will escape key-check.
// Sometimes the point's tag keys are optional for different configures.
func WithOptionalTags(keys ...string) PointCheckOption {
	return func(c *ptChecker) {
		c.optionalTags = append(c.optionalTags, keys...)
	}
}

// CheckPoint used to check pt with various options. If any checking
// failed, the failed message are returned.
func CheckPoint(pt *point.Point, opts ...PointCheckOption) []string {
	c := newPointChecker()

	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	c.gotName = string(pt.Name())
	c.gotTags = pt.Tags()
	c.gotFields = pt.Fields()
	c.gotTime = pt.Time()

	c.doCheck(pt)

	return c.checkMsg
}

func (c *ptChecker) doCheck(pt *point.Point) {
	// check according to doc info
	if c.mInfo != nil {
		c.checkOnDoc(pt)
	}

	if c.expect != nil {
		c.checkOnPoint(pt)
	}
}

func (c *ptChecker) checkOnPoint(pt *point.Point) {
	if !c.measurementCheckIgnored && c.expName != c.gotName {
		c.addMsg(fmt.Sprintf("expect measurement name %q got %q", c.expName, c.gotName))
	}

	if len(c.expTags) != len(c.gotTags) {
		c.addMsg(fmt.Sprintf("expect %d tags got %d", len(c.expTags), len(c.gotTags)))
	}

	if len(c.expFields) != len(c.gotFields) {
		c.addMsg(fmt.Sprintf("expect %d fields got %d", len(c.expFields), len(c.gotFields)))
	}

	for _, kv := range c.expTags {
		if c.gotTags.Get(kv.Key) == nil {
			c.addMsg(fmt.Sprintf("unknown tag %q", kv.Key))
		}
	}

	for _, kv := range c.expFields {
		if got := c.gotFields.Get(kv.Key); got == nil {
			c.addMsg(fmt.Sprintf("expect field %q not found", kv.Key))
		} else if c.checkTypes || c.checkValues {
			if !reflect.DeepEqual(kv.Val, got.Val) {
				c.addMsg(fmt.Sprintf("expect field %q type %q value %v got type %q value %v",
					kv.Key,
					reflect.TypeOf(kv.Val), kv.Val,
					reflect.TypeOf(got.Val), got.Val))
			}
		}
	}
}

func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

func (c *ptChecker) checkOnDoc(pt *point.Point) {
	if !c.measurementCheckIgnored && c.mInfo.Name != string(pt.Name()) {
		c.addMsg(fmt.Sprintf("measurement name not equal: %q <> %q", c.mInfo.Name, string(pt.Name())))
	}

	// check tag key count
	mGotTags := make(map[string]struct{})
	for k := range c.gotTags.InfluxTags() {
		if len(k) > 0 {
			mGotTags[k] = struct{}{}
		}
	}
	for _, v := range c.optionalTags {
		if len(v) > 0 {
			mGotTags[v] = struct{}{}
		}
	}
	if len(c.mInfo.Tags)+len(c.extraTags) != len(mGotTags) {
		c.addMsg(fmt.Sprintf("expect %d tags got %d",
			len(c.mInfo.Tags)+len(c.extraTags),
			len(c.gotTags)+len(c.optionalTags)))
	}

	// check field key count
	mGotFields := make(map[string]struct{})
	for k := range c.gotFields.InfluxFields() {
		if len(k) > 0 {
			mGotFields[k] = struct{}{}
		}
	}
	for _, v := range c.optionalFields {
		if len(v) > 0 {
			mGotFields[v] = struct{}{}
		}
	}
	if len(c.mInfo.Fields) != len(mGotFields) {
		var a, b []string
		for k := range c.mInfo.Fields {
			a = append(a, k)
		}
		for k := range mGotFields {
			b = append(b, k)
		}
		d := difference(a, b)
		_ = d
		c.addMsg(fmt.Sprintf("expect %d fields got %d(%d keys optional)",
			len(c.mInfo.Fields), len(c.gotFields), len(c.optionalFields)))
	}

	// check all documented tags are exist in got tags.
	for key := range c.mInfo.Tags {
		if got := c.gotTags.Get([]byte(key)); got == nil {
			if !c.isOptionalTag(key) {
				c.addMsg(fmt.Sprintf("tag %q not found", key))
			}
		}
	}

	// check all tag key are documented(exclude extra tags).
	for _, kv := range c.gotTags {
		key := string(kv.Key)
		if _, ok := c.mInfo.Tags[key]; ok {
			continue
		}

		if _, ok := c.extraTags[key]; ok {
			continue
		}

		c.addMsg(fmt.Sprintf("tag %q not expected", key))
	}

	// check field key and value
	for key, info := range c.mInfo.Fields {
		if got := c.gotFields.Get([]byte(key)); got == nil { // field not found in @pt
			if c.isOptionalField(key) {
				continue
			} else {
				c.addMsg(fmt.Sprintf("field %q not found in point", key))
			}
		} else {
			// check field value

			switch x := info.(type) {
			case *FieldInfo:
				if c.checkTypes && !typeEqual(x.DataType, got) {
					c.addMsg(fmt.Sprintf("field %q expect type %q got %q",
						key, x.DataType, reflect.TypeOf(got.GetVal())))
				}

				// TODO: check metric type(gauge/count) and unit.
			default:
				c.addMsg(fmt.Sprintf("missing type info on field %q", key))
			}
		}
	}
}

func (c *ptChecker) addMsg(s string) {
	c.checkMsg = append(c.checkMsg, s)
}

func (c *ptChecker) isOptionalTag(key string) bool {
	for _, x := range c.optionalTags {
		if x == key {
			return true
		}
	}

	return false
}

func (c *ptChecker) isOptionalField(key string) bool {
	for _, x := range c.optionalFields {
		if x == key {
			return true
		}
	}

	return false
}

func typeEqual(expect string, f *point.Field) bool {
	switch f.Val.(type) {
	case *point.Field_I, *point.Field_U:
		return expect == Int
	case *point.Field_F:
		return expect == Float
	case *point.Field_B:
		return expect == Bool
	case *point.Field_D:
		return expect == String
	default:
		return false
	}
}
