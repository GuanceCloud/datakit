package parser

import (
	"encoding/json"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestContrast(t *testing.T) {
	cases := []struct {
		x, y     interface{}
		operator string
		pass     bool
	}{
		{
			x:        json.Number("10.0"),
			operator: "==",
			y:        json.Number("10.0"),
			pass:     true,
		},
		{
			x:        json.Number("10.0"),
			operator: "==",
			y:        json.Number("12.0"),
			pass:     false,
		},
		{
			x:        json.Number("10"),
			operator: "!=",
			y:        json.Number("12"),
			pass:     true,
		},
		{
			x:        json.Number("10"),
			operator: "<=",
			y:        json.Number("12"),
			pass:     true,
		},
		{
			x:        float64(10.0),
			operator: "==",
			y:        int64(10),
			pass:     false,
		},
		{
			x:        float64(10.0),
			operator: "==",
			y:        "hello",
			pass:     false,
		},
		{
			x:        float64(10.0),
			operator: "==",
			y:        true,
			pass:     false,
		},
		{
			x:        float64(10.0),
			operator: "==",
			y:        nil,
			pass:     false,
		},
		{
			x:        float64(3.1415),
			operator: "==",
			y:        float64(3.1415),
			pass:     true,
		},
		{
			x:        float64(3.1415),
			operator: "!=",
			y:        float64(3.1415),
			pass:     false,
		},
		{
			x:        float64(3.1415),
			operator: "==",
			y:        float64(12.25),
			pass:     false,
		},
		{
			x:        float64(3.1415),
			operator: "<=",
			y:        float64(12.25),
			pass:     true,
		},
		{
			x:        float64(3.1415),
			operator: ">=",
			y:        float64(12.25),
			pass:     false,
		},
		{
			x:        int64(3),
			operator: "==",
			y:        int64(3),
			pass:     true,
		},
		{
			x:        int64(3),
			operator: "!=",
			y:        int64(3),
			pass:     false,
		},
		{
			x:        int64(3),
			operator: "<=",
			y:        int64(10),
			pass:     true,
		},
		{
			x:        int64(3),
			operator: ">=",
			y:        int64(10),
			pass:     false,
		},
		{
			x:        "ABCD",
			operator: "==",
			y:        "ABCD",
			pass:     true,
		},
		{
			x:        "ABCD",
			operator: "!=",
			y:        "ABCDEEEEEE",
			pass:     true,
		},
		{
			x:        "ABCD",
			operator: "<=",
			y:        "ABCD",
			pass:     false,
		},
		{
			x:        "ABCD",
			operator: "<=",
			y:        int64(10),
			pass:     false,
		},
		{
			x:        "ABCD",
			operator: "==",
			y:        nil,
			pass:     false,
		},
		{
			x:        true,
			operator: "==",
			y:        true,
			pass:     true,
		},
		{
			x:        true,
			operator: "!=",
			y:        true,
			pass:     false,
		},
		{
			x:        true,
			operator: "<=",
			y:        false,
			pass:     false,
		},
		{
			x:        nil,
			operator: "==",
			y:        nil,
			pass:     true,
		},
		{
			x:        nil,
			operator: "!=",
			y:        nil,
			pass:     false,
		},
		{
			x:        nil,
			operator: "<=",
			y:        nil,
			pass:     false,
		},
	}

	for idx, tc := range cases {
		t.Logf("[%d] %v %v %v  %v\n", idx, tc.x, tc.operator, tc.y, tc.pass)
		b, err := contrast(tc.x, tc.operator, tc.y)
		tu.Equals(t, nil, err)
		tu.Equals(t, tc.pass, b)
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
