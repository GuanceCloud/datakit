// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package grok

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDenormalizeGlobalPatterns(t *testing.T) {
	if denormalized, errs := DenormalizePatternsFromMap(defalutPatterns); len(errs) != 0 {
		t.Error(errs)
	} else {
		if len(defalutPatterns) != len(denormalized) {
			t.Error("len(GlobalPatterns) != len(denormalized)")
		}
		for k := range denormalized {
			if _, ok := defalutPatterns[k]; !ok {
				t.Errorf("%s not exists", k)
			}
		}
	}
}

func TestParse(t *testing.T) {
	patternINT, err := DenormalizePattern(defalutPatterns["INT"])
	if err != nil {
		t.Error(err)
	}

	patterns := map[string]*GrokPattern{
		"INT": patternINT,
	}

	denormalized, errs := DenormalizePatternsFromMap(defalutPatterns, patterns)
	if len(errs) != 0 {
		t.Error(errs)
	}
	g, err := CompilePattern("%{DAY:day}", denormalized)
	if err != nil {
		t.Error(err)
	}
	ret, err := g.Run("Tue qds")
	if err != nil {
		t.Error(err)
	}
	if ret["day"] != "Tue" {
		t.Fatalf("day should be 'Tue' have '%s'", ret["day"])
	}
}

func TestParseFromPathPattern(t *testing.T) {
	pathPatterns, err := LoadPatternsFromPath("./patterns")
	if err != nil {
		t.Error(err)
	}
	de, errs := DenormalizePatternsFromMap(pathPatterns)
	if len(errs) != 0 {
		t.Error(errs)
	}
	g, err := CompilePattern("%{DAY:day}", de)
	if err != nil {
		t.Error(err)
	}
	ret, err := g.Run("Tue qds")
	if err != nil {
		t.Error(err)
	}
	if ret["day"] != "Tue" {
		t.Fatalf("day should be 'Tue' have '%s'", ret["day"])
	}
}

func TestLoadPatternsFromPathErr(t *testing.T) {
	_, err := LoadPatternsFromPath("./Lorem ipsum Minim qui in.")
	if err == nil {
		t.Fatalf("AddPatternsFromPath should returns an error when path is invalid")
	}
}

func TestRunWithTypeInfo(t *testing.T) {
	tCase := []struct {
		data      string
		ptn       string
		ret       map[string]interface{}
		failedRet map[string]string
		failed    bool
	}{
		{
			data: `1
true 1.1`,
			ptn: `%{INT:A:int}
%{WORD:B:bool} %{BASE10NUM:C:float}`,
			ret: map[string]interface{}{
				"A": int64(1),
				"B": true,
				"C": float64(1.1),
			},
			failedRet: map[string]string{},
		},
		{
			data: `1
true 1.1`,
			ptn: `%{INT:A:int}
%{WORD:B:bool} %{BASE10NUM:C:int}`,
			ret: map[string]interface{}{
				"A": int64(1),
				"B": true,
				"C": int64(0),
			},
			failedRet: map[string]string{
				"C": "1.1",
			},
		},
		{
			data: `1 ijk123abc
true 1.1`,
			ptn: `%{INT:A:int} %{WORD:S:string}
%{WORD:B:bool} %{BASE10NUM:C:int}`,
			ret: map[string]interface{}{
				"A": int64(1),
				"S": "ijk123abc",
				"B": true,
				"C": int64(0),
			},
			failedRet: map[string]string{
				"C": "1.1",
			},
		},
		{
			data: `1
true 1.1`,
			ptn: `%{INT:A}
%{WORD:B:bool} %{BASE10NUM:C:int}`,
			ret: map[string]interface{}{
				"A": "1",
				"B": true,
				"C": int64(0),
			},
			failedRet: map[string]string{
				"C": "1.1",
			},
		},
	}

	for _, item := range tCase {
		g, err := CompilePattern(item.ptn, defalutDenormalizedPatterns)
		if err != nil {
			t.Fatal(err)
		}
		v, vf, err := g.RunWithTypeInfo(item.data)
		if err != nil && !item.failed {
			t.Fatal(err)
		}
		assert.Equal(t, item.ret, v)
		assert.Equal(t, item.failedRet, vf)
	}
}

func BenchmarkFromMap(b *testing.B) {
	pathPatterns, err := LoadPatternsFromPath("./patterns")
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		de, errs := DenormalizePatternsFromMap(pathPatterns)
		if len(errs) != 0 {
			b.Error(err)
			b.Error(de)
		}
	}
}
