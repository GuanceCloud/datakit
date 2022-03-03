package parser

import (
	"encoding/json"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestContrast(t *testing.T) {
	cases := []struct {
		name         string
		x, y         interface{}
		operator     string
		expect, fail bool
	}{
		{
			x:        json.Number("10.0"),
			operator: "==",
			y:        json.Number("10.0"),
			expect:   true,
		},
		{
			x:        json.Number("10.0"),
			operator: "==",
			y:        json.Number("12.0"),
			expect:   false,
		},
		{
			x:        json.Number("10"),
			operator: "!=",
			y:        json.Number("12"),
			expect:   true,
		},
		{
			x:        json.Number("10"),
			operator: "<=",
			y:        json.Number("12"),
			expect:   true,
		},
		{
			name:     "float==int",
			x:        float64(10.0),
			operator: "==",
			y:        int64(10),
			fail:     true,
		},
		{
			name:     "flaot==string",
			x:        float64(10.0),
			operator: "==",
			y:        "hello",
			fail:     true,
		},
		{
			name:     "flaot==bool",
			x:        float64(10.0),
			operator: "==",
			y:        true,
			fail:     true,
		},
		{
			x:        float64(10.0),
			operator: "==",
			y:        nil,
			expect:   false,
		},
		{
			x:        float64(3.1415),
			operator: "==",
			y:        float64(3.1415),
			expect:   true,
		},
		{
			x:        float64(3.1415),
			operator: "!=",
			y:        float64(3.1415),
			expect:   false,
		},
		{
			x:        float64(3.1415),
			operator: "==",
			y:        float64(12.25),
			expect:   false,
		},
		{
			x:        float64(3.1415),
			operator: "<=",
			y:        float64(12.25),
			expect:   true,
		},
		{
			x:        float64(3.1415),
			operator: ">=",
			y:        float64(12.25),
			expect:   false,
		},
		{
			x:        int64(3),
			operator: "==",
			y:        int64(3),
			expect:   true,
		},
		{
			x:        int64(3),
			operator: "!=",
			y:        int64(3),
			expect:   false,
		},
		{
			x:        int64(3),
			operator: "<=",
			y:        int64(10),
			expect:   true,
		},
		{
			x:        int64(3),
			operator: ">=",
			y:        int64(10),
			expect:   false,
		},
		{
			x:        "ABCD",
			operator: "==",
			y:        "ABCD",
			expect:   true,
		},
		{
			x:        "ABCD",
			operator: "!=",
			y:        "ABCDEEEEEE",
			expect:   true,
		},
		{
			x:        "ABCD",
			operator: "<=",
			y:        "ABCD",
			expect:   false,
		},
		{
			name:     "string<=int",
			x:        "ABCD",
			operator: "<=",
			y:        int64(10),
			fail:     true,
		},
		{
			x:        "ABCD",
			operator: "==",
			y:        nil,
			expect:   false,
		},
		{
			x:        true,
			operator: "==",
			y:        true,
			expect:   true,
		},
		{
			x:        true,
			operator: "!=",
			y:        true,
			expect:   false,
		},
		{
			name:     "bool<=bool",
			x:        true,
			operator: "<=",
			y:        false,
			fail:     true,
		},
		{
			x:        nil,
			operator: "==",
			y:        nil,
			expect:   true,
		},
		{
			x:        nil,
			operator: "!=",
			y:        nil,
			expect:   false,
		},
		{
			x:        nil,
			operator: "<=",
			y:        nil,
			expect:   false,
		},

		{
			x:        nil,
			operator: "<=",
			y:        10,
			expect:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := contrast(tc.x, tc.operator, tc.y)
			if !tc.fail {
				tu.Ok(t, err)
				tu.Equals(t, tc.expect, b)
			} else {
				tu.NotOk(t, err, "")
				return
			}
		})
	}

	t.Log("END")
}

func TestCheckOutPutNilPtr(t *testing.T) {
	var out *Output
	checkOutPutNilPtr(&out)
	if out == nil {
		t.Error("nil")
	}
}
