// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package grok

import (
	"testing"

	"github.com/ubwbu/grok"
	vgrok "github.com/vjeantet/grok"
)

func TestParsePattern(t *testing.T) {
	pattern := grok.CopyDefalutPatterns()
	if de, err := grok.DenormalizePatternsFromMap(pattern); err != nil {
		t.Error(err)
	} else {
		if len(pattern) != len(de) {
			t.Error("length not equal")
		}
		for k := range de {
			if _, ok := de[k]; !ok {
				t.Errorf("%s not exists", k)
			}
		}
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if v, err := grok.DenormalizePatternsFromMap(grok.CopyDefalutPatterns()); err != nil {
			b.Log(v)
			b.Error(err)
		}
	}
}

func BenchmarkParseVgrok(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if g, err := vgrok.NewWithConfig(&vgrok.Config{NamedCapturesOnly: true}); err != nil {
			b.Error(err)
			b.Error(g)
		}
	}
}
