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

func TestMultilineEndToEndFlow(t *testing.T) {
	t.Run("auto-default-rules-and-extra-patterns", func(t *testing.T) {
		m, err := NewAuto([]string{`^APPSTART\b`})
		assert.NoError(t, err)

		out := collectStringLines(m, []string{
			"APPSTART custom entry",
			"  custom continuation",
			"2024-07-08 10:00:00 INFO default timestamp entry",
			"  timestamp continuation",
			"INFO [main] default level entry",
		})

		assert.Equal(t, []string{
			"APPSTART custom entry\n  custom continuation",
			"2024-07-08 10:00:00 INFO default timestamp entry\n  timestamp continuation",
			"INFO [main] default level entry",
		}, out)
	})

	t.Run("manual-pattern-does-not-use-auto-defaults", func(t *testing.T) {
		m, err := New([]string{`^MANUAL\b`})
		assert.NoError(t, err)

		out := collectStringLines(m, []string{
			"MANUAL first entry",
			"2024-07-08 10:00:00 INFO default-looking continuation",
			"MANUAL second entry",
		})

		assert.Equal(t, []string{
			"MANUAL first entry\n2024-07-08 10:00:00 INFO default-looking continuation",
			"MANUAL second entry",
		}, out)
	})

	t.Run("auto-keeps-stack-continuations", func(t *testing.T) {
		m, err := NewAuto(nil)
		assert.NoError(t, err)

		out := collectStringLines(m, []string{
			`[2020-12-03 11:36:23] ERROR in app: request failed on /error [GET]`,
			`Traceback (most recent call last):`,
			`  File "/app.py", line 1, in <module>`,
			`[2020-12-03 11:36:24] "GET /health HTTP/1.1" 200 -`,
		})

		assert.Equal(t, []string{
			"[2020-12-03 11:36:23] ERROR in app: request failed on /error [GET]\n" +
				"Traceback (most recent call last):\n" +
				`  File "/app.py", line 1, in <module>`,
			`[2020-12-03 11:36:24] "GET /health HTTP/1.1" 200 -`,
		}, out)
	})

	t.Run("auto-fractional-time-prefix-keeps-continuations", func(t *testing.T) {
		m, err := NewAuto(nil)
		assert.NoError(t, err)

		out := collectStringLines(m, []string{
			`15:03:41.922 [http-nio-9002-exec-24] ERROR [bbs-mg-community] request failed`,
			`request detail: user isn't exist`,
			`       stack frame: Author.getBookIds(Author.java:38)`,
			`15:03:42.922 [http-nio-9002-exec-24] ERROR [bbs-mg-community] next request failed`,
			`operation detail: failed`,
		})

		assert.Equal(t, []string{
			"15:03:41.922 [http-nio-9002-exec-24] ERROR [bbs-mg-community] request failed\n" +
				"request detail: user isn't exist\n" +
				"       stack frame: Author.getBookIds(Author.java:38)",
			"15:03:42.922 [http-nio-9002-exec-24] ERROR [bbs-mg-community] next request failed\n" +
				"operation detail: failed",
		}, out)
	})

	t.Run("auto-filebeat-style-bracket-starts", func(t *testing.T) {
		m, err := NewAuto(nil)
		assert.NoError(t, err)

		out := collectStringLines(m, []string{
			`[beat-logstash-some-name-832-2015.11.28] index not found`,
			`    at org.elasticsearch.cluster.metadata.IndexNameExpressionResolver.resolve(IndexNameExpressionResolver.java:566)`,
			`[2015-08-24 11:49:14,389][INFO ][env] [Letha] using [1] data paths, mounts [[/`,
			`(/dev/disk1)]], net usable_space [34.5gb]`,
		})

		assert.Equal(t, []string{
			"[beat-logstash-some-name-832-2015.11.28] index not found\n" +
				"    at org.elasticsearch.cluster.metadata.IndexNameExpressionResolver.resolve(IndexNameExpressionResolver.java:566)",
			"[2015-08-24 11:49:14,389][INFO ][env] [Letha] using [1] data paths, mounts [[/\n" +
				"(/dev/disk1)]], net usable_space [34.5gb]",
		}, out)
	})
}

func collectStringLines(m *Multiline, lines []string) []string {
	var out []string
	for _, line := range lines {
		text, _ := m.ProcessLineString(line)
		if text != "" {
			out = append(out, text)
		}
	}
	if text := m.FlushString(); text != "" {
		out = append(out, text)
	}
	return out
}

func TestMultilineBoundaries(t *testing.T) {
	t.Run("no-context", func(t *testing.T) {
		m, err := New([]string{`^BEGIN`})
		assert.NoError(t, err)

		res, state := m.ProcessLineString("\tcontinuation without first line")
		assert.Equal(t, "\tcontinuation without first line", res)
		assert.Equal(t, NoContext, state)
	})

	t.Run("max-length-flushes-partial", func(t *testing.T) {
		m, err := New([]string{`^BEGIN`}, WithMaxLength(20))
		assert.NoError(t, err)

		_, state := m.ProcessLineString("BEGIN first")
		assert.Equal(t, NewMultiline, state)

		res, state := m.ProcessLineString("  long continuation")
		assert.Equal(t, "BEGIN first\n  long continuation", res)
		assert.Equal(t, FlushPartial, state)
	})
}

func TestNewMultilineValidation(t *testing.T) {
	_, err := New([]string{`^BEGIN`})
	assert.NoError(t, err)

	_, err = NewAuto([]string{`^APPSTART`})
	assert.NoError(t, err)

	_, err = New([]string{"(?!"})
	assert.Error(t, err)

	_, err = NewAuto([]string{"(?!"})
	assert.Error(t, err)
}

func TestTrimRightSpace(t *testing.T) {
	cases := []struct {
		in, out string
	}{
		{in: "", out: ""},
		{in: "123\n", out: "123"},
		{in: "123\r\n", out: "123"},
		{in: "\t123\t\r\n", out: "\t123"},
		{in: "\t123\t456\r\n", out: "\t123\t456"},
	}

	for _, tc := range cases {
		assert.Equal(t, []byte(tc.out), TrimRightSpace([]byte(tc.in)), tc.in)
	}
}
