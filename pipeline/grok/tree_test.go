package grok

import (
	"testing"
)

func TestTree(t *testing.T) {
	cases := []map[string]string{
		{
			"aaa": "aa",
			"bb":  "%{aaa}, %{dd}",
			"cc":  "%{bb} %{aaa} %{dd}",
			"dd":  "%{cc} %{bb}",
		},
		{
			"aaa": "aa",
			"bb":  "%{aaa}, %{dd}",
			"cc":  "%{bb} %{aaa} %{dd}",
			"dd":  "%{cc} %{bb}",
			"ee":  "%{ff}",
		},
	}

	ret := [][2]map[string]string{
		{
			map[string]string{"aaa": "aa"},
			map[string]string{
				"bb": "circular dependency: pattern bb -> dd -> cc -> bb",
				"cc": "circular dependency: pattern cc -> bb -> dd -> cc",
				"dd": "circular dependency: pattern dd -> cc -> bb -> dd",
			},
		},
		{
			map[string]string{"aaa": "aa"},
			map[string]string{
				"bb": "circular dependency: pattern bb -> dd -> cc -> bb",
				"cc": "circular dependency: pattern cc -> bb -> dd -> cc",
				"dd": "circular dependency: pattern dd -> cc -> bb -> dd",
				"ee": "no pattern found for %{ff}",
			},
		},
	}

	for i, pat := range cases {
		v, errs := DenormalizePatternsFromMap(pat)
		for k, v := range v {
			expected, ok := ret[i][0][k]
			if !ok {
				t.Fatal("")
			}
			if expected != v {
				t.Errorf("value %s act: %s exp: %s", k, v, expected)
			}
		}

		for k, v := range errs {
			expected, ok := ret[i][1][k]
			if !ok {
				t.Fatal("")
			}
			if expected != v {
				t.Errorf("error %s act: %s exp: %s", k, v, expected)
			}
		}
	}
}
