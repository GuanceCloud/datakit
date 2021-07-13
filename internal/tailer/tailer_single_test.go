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

	for _, tc := range cases {
		logs := newLogs(tc.in)
		err := logs.pipeline(nil).
			checkFieldsLength().
			addStatus(false).
			ignoreStatus(nil).
			takeTime().
			point("testing", nil).
			error()
		if err != nil {
			t.Error(err)
		}

		t.Log(logs.output())
	}
}
