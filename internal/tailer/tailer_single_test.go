package tailer

import (
	"testing"
)

func TestProcessText(t *testing.T) {
	var cases = []struct {
		in string
	}{
		{
			"2020-10-23 06:41:56,688 INFO demo.py 5.0",
		},
	}

	tl := &TailerSingle{opt: &Option{}}

	for _, tc := range cases {
		out := tl.processText(tc.in)
		t.Log(out)
	}
}
