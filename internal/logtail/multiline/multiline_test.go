// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package multiline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMultilineMatch(t *testing.T) {
	cases := []struct {
		name    string
		pattern string
		in, out []string
	}{
		{
			name:    "mysql-slowlog",
			pattern: "^(# Time|\\d{4}-\\d{2}-\\d{2}|\\d{6}\\s+\\d{2}:\\d{2}:\\d{2})",
			in: []string{
				"# Time: 2021-05-31T11:15:26.043419Z",
				"# User@Host: datakitMonitor[datakitMonitor] @ localhost []  Id:  1228",
				"# Query_time: 0.015214  Lock_time: 0.000112 Rows_sent: 4  Rows_examined: 288",
				"SET timestamp=1622459726;",
				"SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb",
				"                FROM     information_schema.tables",
				"                GROUP BY table_schema;",
			},
			out: []string{"# Time: 2021-05-31T11:15:26.043419Z\n# User@Host: datakitMonitor[datakitMonitor] @ localhost []  Id:  1228\n# Query_time: 0.015214  Lock_time: 0.000112 Rows_sent: 4  Rows_examined: 288\nSET timestamp=1622459726;\nSELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb\n                FROM     information_schema.tables\n                GROUP BY table_schema;"},
		},
		{
			name:    "flushing-two-groups",
			pattern: "^\\S",
			in:      []string{"2021-05-31T11:15:26.043419Z INFO", "2021-05-31T11:15:26.043419Z WARN"},
			out:     []string{"2021-05-31T11:15:26.043419Z INFO", "2021-05-31T11:15:26.043419Z WARN"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := New([]string{tc.pattern}, nil)
			assert.NoError(t, err)

			outIdx := 0
			for _, line := range tc.in {
				res := m.ProcessLineString(line)
				if res != "" {
					assert.Equal(t, tc.out[outIdx], res)
					outIdx++
				}
			}

			assert.Equal(t, tc.out[outIdx], m.FlushString())
		})
	}
}

func TestMultilineMatchLimit(t *testing.T) {
	t.Run("buff-is-zero", func(t *testing.T) {
		patterns := []string{}
		m, err := New(patterns, nil)
		assert.NoError(t, err)

		assert.Equal(t, 0, m.BuffLength())

		// 当 buff 为空时，即使匹配失败也会 flush
		res := m.ProcessLineString("\tnomatched-head-is-space")
		assert.Equal(t, "\tnomatched-head-is-space", res)
	})

	t.Run("flush-duration", func(t *testing.T) {
		patterns := []string{"^\\S"}
		opt := &Option{
			MaxLifeDuration: time.Millisecond * 100,
		}

		m, err := New(patterns, opt)
		assert.NoError(t, err)

		_ = m.ProcessLineString("2021-05-31T11:15:26.043419Z INFO")

		time.Sleep(time.Millisecond * 150)

		res := m.ProcessLineString("\t1234567890-1234567890-1234567890")
		assert.Equal(t, "2021-05-31T11:15:26.043419Z INFO\n\t1234567890-1234567890-1234567890", res)
	})

	t.Run("max-length-50", func(t *testing.T) {
		patterns := []string{"^\\S"}
		opt := &Option{
			MaxLength: 50,
		}

		m, err := New(patterns, opt)
		assert.NoError(t, err)

		_ = m.ProcessLineString("2021-05-31T11:15:26.043419Z INFO")
		res := m.ProcessLineString("\t1234567890-1234567890-1234567890")
		assert.Equal(t, "2021-05-31T11:15:26.043419Z INFO\n\t1234567890-1234567890-1234567890", res)
	})
}

func TestNewMultiline(t *testing.T) {
	t.Run("ok-1", func(t *testing.T) {
		_, err := New(nil, nil)
		assert.NoError(t, err)
	})

	t.Run("ok-2", func(t *testing.T) {
		_, err := New([]string{"^\\S"}, nil)
		assert.NoError(t, err)
	})

	t.Run("ok-3", func(t *testing.T) {
		_, err := New([]string{"^\\S"}, &Option{MaxLength: 100})
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		_, err := New([]string{"(?!"}, nil)
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
			assert.Equal(t, TrimRightSpace(tc.in), tc.out)
		})
	}
}
