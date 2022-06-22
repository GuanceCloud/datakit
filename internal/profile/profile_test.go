package profile

import (
	"fmt"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	t1 := "2022-04-12T17:37:55Z"
	tm, err := time.Parse(time.RFC3339Nano, t1)
	if err != nil {
		t.Fatalf("can not parse in format time.RFC3339: %s, %s", t1, err)
	}
	t2 := "2022-04-12T17:37:55.638161100Z"
	tm2, err := time.Parse(time.RFC3339Nano, t2)
	if err != nil {
		t.Fatalf("can not parse in format time.RFC3339: %s, %s", t2, err)
	}
	fmt.Println(tm.UnixNano(), tm2.UnixNano())

	tm3 := time.Time{}

	fmt.Println(tm3)
}

func TestParseRuntime(t *testing.T) {
	cases := map[string]Lang{
		"jvm":     Java,
		"Java":    Java,
		"go":      Golang,
		"Go":      Golang,
		"Golang":  Golang,
		"CPython": Python,
		"Python":  Python,
	}

	for s, lang := range cases {
		parsed := ResolveLanguage([]string{s})
		if parsed != lang {
			t.Fatalf("parse language str:%s, assertion: %s, in fact: %s", s, lang, parsed)
		}
		fmt.Println(s, parsed)
	}
}
