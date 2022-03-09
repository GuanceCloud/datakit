package grok

import (
	"testing"

	"github.com/ubwbu/grok"
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
	pattern := grok.CopyDefalutPatterns()

	for i := 0; i < b.N; i++ {
		if v, err := grok.DenormalizePatternsFromMap(pattern); err != nil {
			b.Log(v)
			b.Error(err)
		}
	}
}
