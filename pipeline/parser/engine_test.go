package parser

import (
	"encoding/json"
	"testing"
)

func TestContrast(t *testing.T) {
	tests := []struct {
		x, y     interface{}
		operator string
		ok       bool
	}{
		{
			x:        3.1415,
			operator: "==",
			y:        3.1415,
			ok:       true,
		},
		{
			x:        3.1415,
			operator: "==",
			y:        12.25,
			ok:       false,
		},
		{
			x:        3.1415,
			operator: "!=",
			y:        12.25,
			ok:       true,
		},
		{
			x:        3.1415,
			operator: ">",
			y:        12.25,
			ok:       false,
		},
		{
			x:        3.1415,
			operator: ">=",
			y:        12.25,
			ok:       false,
		},
		{
			x:        3.1415,
			operator: "<",
			y:        12.25,
			ok:       true,
		},
		{
			x:        3.1415,
			operator: "<=",
			y:        12.25,
			ok:       true,
		},
		/* // int64(3)
		{
			x:        3,
			operator: "<=",
			y:        12.25,
			ok:       true,
		},
		*/
		{
			x:        int64(3),
			operator: "<=",
			y:        12.25,
			ok:       true,
		},
		{
			x:        int64(3),
			operator: "!=",
			y:        12.25,
			ok:       true,
		},
		{
			x:        json.Number("10"),
			operator: "==",
			y:        json.Number("10.0"),
			ok:       true,
		},
		{
			x:        "ABCD",
			operator: "==",
			y:        "ABCD",
			ok:       true,
		},
		{
			x:        "ABCD",
			operator: "!=",
			y:        "ABCDEEEEEE",
			ok:       true,
		},
		{
			x:        "ABCD",
			operator: "<=",
			y:        "ABCD",
			ok:       false,
		},
		{
			x:        true,
			operator: "==",
			y:        true,
			ok:       true,
		},
		{
			x:        true,
			operator: "!=",
			y:        true,
			ok:       false,
		},
		{
			x:        true,
			operator: "<=",
			y:        false,
			ok:       false,
		},
	}

	var b bool

	for idx, ts := range tests {
		b = contrast(ts.x, ts.operator, ts.y)
		if b == ts.ok {
			t.Logf("[%d] OK, pass: (%v %s %v)\n", idx, ts.x, ts.operator, ts.y)
		} else {
			t.Logf("[%d] not:  (%v %s %v)\n", idx, ts.x, ts.operator, ts.y)
		}
	}

	t.Log("END")
}
