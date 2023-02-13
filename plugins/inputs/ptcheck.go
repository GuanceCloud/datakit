// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"fmt"
	"reflect"

	"github.com/GuanceCloud/cliutils/point"
)

type ptChecker struct {
	checkKeys      bool
	checkTypes     bool
	allowExtraTags bool
}

func WithAllowExtraTags(on bool) PointCheckOption {
	return func(c *ptChecker) { c.allowExtraTags = true }
}

func newPointChecker() *ptChecker {
	return &ptChecker{
		checkKeys:  true,
		checkTypes: true,
	}
}

type PointCheckOption func(*ptChecker)

func CheckPoint(pt *point.Point, m Measurement, opts ...PointCheckOption) []string {
	c := newPointChecker()

	var errMsg []string

	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	info := m.Info()

	if c.checkKeys {
		// check measurement name
		if info.Name != string(pt.Name()) {
			errMsg = append(errMsg, fmt.Sprintf("measurement name not equal: %s <> %s", info.Name, string(pt.Name())))
		}

		tags := pt.Tags()
		fields := pt.Fields()

		if len(tags) != len(info.Tags) {

			if len(tags) < len(info.Tags) {
				errMsg = append(errMsg, fmt.Sprintf("expect %d tags, got %d", len(info.Tags), len(tags)))
			} else {
				if !c.allowExtraTags {
					errMsg = append(errMsg, fmt.Sprintf("tag cound not equal: %d <> %d", len(tags), len(info.Tags)))
				} else {
					// pass
				}
			}
		}

		if len(fields) != len(info.Fields) {
			errMsg = append(errMsg, fmt.Sprintf("field cound not equal: %d <> %d", len(fields), len(info.Fields)))
		}

		// check each tags
		for k, _ := range info.Tags { // expect
			if v := tags.Get([]byte(k)); v != nil {
				// TODO: check tag value
			} else {
				errMsg = append(errMsg, fmt.Sprintf("tag %s not found", k))
			}
		}

		// check each field
		for k, f := range info.Fields { // expect
			if got := fields.Get([]byte(k)); got != nil {
				// TODO: check field type

				switch x := f.(type) {
				case *FieldInfo:
					if c.checkTypes && !typeEqual(x.DataType, got) {
						errMsg = append(errMsg,
							fmt.Sprintf("field '%s' expect type %s, got %s", k, x.DataType, reflect.TypeOf(got.GetVal())))
					}
				default:
					errMsg = append(errMsg, fmt.Sprintf("missing type info on field %s", k))
				}

			} else {
				errMsg = append(errMsg, fmt.Sprintf("field %s not found", k))
			}
		}
	}

	return errMsg
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
