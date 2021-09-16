package tailer

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	ts, _ = time.Parse(time.RFC3339, "2021-06-30T16:25:27Z")

	timeStr = func() string {
		return strconv.Itoa(int(ts.UnixNano()))
	}()
)

func TestLogsAll(t *testing.T) {
	const source = "testing"

	cases := []struct {
		text string
		res  string
	}{
		{
			"2021-06-30T16:25:27.394+0800    DEBUG   config  config/inputcfg.go:27   ignore dir /usr/local/datakit/conf.d/jvm",
			//nolint:lll
			`testing message="2021-06-30T16:25:27.394+0800    DEBUG   config  config/inputcfg.go:27   ignore dir /usr/local/datakit/conf.d/jvm",status="info" ` + timeStr,
		},
	}

	for _, tc := range cases {
		logs := NewLogs(tc.text)
		logs.ts = ts

		output := logs.Pipeline(nil).
			CheckFieldsLength().
			AddStatus(false).
			IgnoreStatus(nil).
			Point(source, nil).
			Output()

		assert.Equal(t, tc.res, output)
	}
}

func TestRemoveAnsiColorOfText(t *testing.T) {
	cases := []struct {
		text string
		res  string
	}{
		{
			"\u001b[1m\u001b[38;5;231mHello World\u001b[0m\u001b[22m",
			"Hello World",
		},
		{
			"a\033[4A\033[4Abc",
			"abc",
		},
	}

	for _, tc := range cases {
		logs := NewLogs(tc.text)
		logs.RemoveAnsiEscapeCodesOfText(true)
		assert.Equal(t, tc.res, logs.text)
	}
}
