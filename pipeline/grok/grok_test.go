package grok

import (
	"testing"
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

	patterns := map[string]string{
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
