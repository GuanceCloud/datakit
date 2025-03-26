// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package multiline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkMultilineMatch(b *testing.B) {
	in := []string{"2021-05-31T11:15:26.043419Z INFO", "2021-05-31T11:15:26.043419Z WARN"}
	m, _ := New(nil)

	for i := 0; i < b.N; i++ {
		_, _ = m.ProcessLineString(in[0])
		_, _ = m.ProcessLineString(in[1])
		_, _ = m.ProcessLineString("")
	}
}

func TestMultilineMatch(t *testing.T) {
	t.Run("mysql-slowlog", func(t *testing.T) {
		pattern := "^(# Time|\\d{4}-\\d{2}-\\d{2}|\\d{6}\\s+\\d{2}:\\d{2}:\\d{2})"
		in := []string{
			"# Time: 2021-05-31T11:15:26.043419Z",
			"# User@Host: datakitMonitor[datakitMonitor] @ localhost []  Id:  1228",
			"# Query_time: 0.015214  Lock_time: 0.000112 Rows_sent: 4  Rows_examined: 288",
			"SET timestamp=1622459726;",
			"SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb",
			"                FROM     information_schema.tables",
			"                GROUP BY table_schema;",
		}
		out := "# Time: 2021-05-31T11:15:26.043419Z\n# User@Host: datakitMonitor[datakitMonitor] @ localhost []  Id:  1228\n# Query_time: 0.015214  Lock_time: 0.000112 Rows_sent: 4  Rows_examined: 288\nSET timestamp=1622459726;\nSELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb\n                FROM     information_schema.tables\n                GROUP BY table_schema;"

		m, err := New([]string{pattern})
		assert.NoError(t, err)

		for idx := range in {
			_, state := m.ProcessLineString(in[idx])

			if idx == 0 {
				assert.Equal(t, NewMultiline, state)
			} else {
				assert.Equal(t, Written, state)
			}
		}

		assert.Equal(t, out, m.FlushString())
	})

	t.Run("flushing-two-groups", func(t *testing.T) {
		patterns := []string{"^\\S"}
		in := []string{"2021-05-31T11:15:26.043419Z INFO", "2021-05-31T11:15:26.043419Z WARN"}
		out := []string{"2021-05-31T11:15:26.043419Z INFO", "2021-05-31T11:15:26.043419Z WARN"}

		m, err := New(patterns)
		assert.NoError(t, err)

		_, state := m.ProcessLineString(in[0])
		assert.Equal(t, NewMultiline, state)

		res, state := m.ProcessLineString(in[1])
		assert.Equal(t, NewMultiline, state)

		assert.Equal(t, out[0], res)
		assert.Equal(t, out[1], m.FlushString())
	})
}

func TestMultilineMatchLimit(t *testing.T) {
	t.Run("buff-is-zero", func(t *testing.T) {
		patterns := []string{}
		m, err := New(patterns)
		assert.NoError(t, err)

		assert.Equal(t, 0, m.BuffLength())

		// 当 buff 为空时，即使匹配失败也会 flush
		res, state := m.ProcessLineString("\tnomatched-head-is-space")
		assert.Equal(t, "\tnomatched-head-is-space", res)
		assert.Equal(t, NoContext, state)
	})

	t.Run("max-length-50", func(t *testing.T) {
		patterns := []string{"^\\S"}
		m, err := New(patterns, WithMaxLength(50))
		assert.NoError(t, err)

		_, state := m.ProcessLineString("2021-05-31T11:15:26.043419Z INFO")
		assert.Equal(t, NewMultiline, state)

		res, state := m.ProcessLineString("\t1234567890-1234567890-1234567890")
		assert.Equal(t, "2021-05-31T11:15:26.043419Z INFO\n\t1234567890-1234567890-1234567890", res)
		assert.Equal(t, FlushPartial, state)
	})
}

func TestNewMultiline(t *testing.T) {
	t.Run("ok-1", func(t *testing.T) {
		_, err := New(nil)
		assert.NoError(t, err)
	})

	t.Run("ok-2", func(t *testing.T) {
		_, err := New([]string{"^\\S"})
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		_, err := New([]string{"(?!"})
		assert.Error(t, err)
	})
}

func TestTrimRightSpace(t *testing.T) {
	cases := []struct {
		in, out string
	}{
		{
			in:  "",
			out: "",
		},
		{
			in:  "123",
			out: "123",
		},
		{
			in:  "\n",
			out: "",
		},
		{
			in:  "123\n",
			out: "123",
		},
		{
			in:  "123\r\n",
			out: "123",
		},
		{
			in:  "123\t\t",
			out: "123",
		},
		{
			in:  "123\t\r\n",
			out: "123",
		},
		{
			in:  "\t123\t\r\n",
			out: "\t123",
		},
		{
			in:  "\t123\t456\r\n",
			out: "\t123\t456",
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			in := []byte(tc.in)
			out := []byte(tc.out)
			assert.Equal(t, TrimRightSpace(in), out)
		})
	}
}

func TestFlushPartial(t *testing.T) {
	in := []string{
		"2021-05-31T11:15:26 INFO 123", // len(28)
		"    A-0123456789",             // len(16)
		"    B-0123456789",
		"    C-0123456789",
		"    D-0123456789",
		"    E-0123456789",
	}

	out := []string{
		"",
		"2021-05-31T11:15:26 INFO 123\n    A-0123456789",
		"",
		"    B-0123456789\n    C-0123456789",
		"",
		"    D-0123456789\n    E-0123456789",
	}

	patterns := []string{"^\\S"}
	m, err := New(patterns, WithMaxLength(30))
	assert.NoError(t, err)

	for idx := range in {
		res, _ := m.ProcessLineString(in[idx])
		assert.Equal(t, out[idx], res)
	}
}
