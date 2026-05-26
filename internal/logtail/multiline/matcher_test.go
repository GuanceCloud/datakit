// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package multiline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkMatch(b *testing.B) {
	content := "2022-08-15T15:04:05Z08:00  INFO  cmd/main.go  Running, Running, Running"
	m, _ := NewAutoMatcher(nil)

	for i := 0; i < b.N; i++ {
		_ = m.MatchString(content)
	}
}

func TestMatcherAutoPatterns(t *testing.T) {
	m, err := NewAutoMatcher([]string{`^\s*APPSTART\b`})
	assert.NoError(t, err)

	cases := []struct {
		line  string
		match bool
	}{
		{line: "2024-07-08 10:00:00 INFO started", match: true},
		{line: "1719820800 INFO unix epoch seconds", match: true},
		{line: "15:03:41.922 [http-nio-9002-exec-24] ERROR [bbs-mg-community] request failed", match: true},
		{line: "15:03:41,922 request failed", match: true},
		{line: "INFO [main] com.example.App - started", match: true},
		{line: "INFO | app.module:func:10 - started", match: true},
		{line: "level=info msg=\"started\"", match: true},
		{line: "START RequestId: 5c936f1c-8f4d-4c68-9b02-1a3a0d3d8d89 Version: $LATEST", match: true},
		{line: "[2020-12-03 11:36:20] INFO started", match: true},
		{line: "[2020-12-03] INFO started", match: true},
		{line: "[beat-logstash-some-name-832-2015.11.28] index not found", match: true},
		{line: `{"msg":"hello"}`, match: true},
		{line: "APPSTART custom format", match: true},
		{line: "\tAPPSTART custom format", match: true},
		{line: `  {"msg":"hello"}`, match: false},
		{line: "\tstack line", match: false},
		{line: "_ Jan 02 15:04:05 2006 continuation-looking line", match: false},
		{line: "[2020-12-03-not-a-date] continuation-looking line", match: false},
		{line: "[signal SIGSEGV: segmentation violation code=0x1]", match: false},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.match, m.MatchString(tc.line), tc.line)
	}
}

func TestMatcherManualPatterns(t *testing.T) {
	m, err := NewMatcher([]string{`^MANUAL\b`})
	assert.NoError(t, err)

	assert.True(t, m.MatchString("MANUAL first line"))
	assert.False(t, m.MatchString("2024-07-08 10:00:00 INFO default-looking line"))
	assert.False(t, m.MatchString("\tstack line"))
}

func TestMatcherDefaultFallback(t *testing.T) {
	m, err := NewMatcher(nil)
	assert.NoError(t, err)
	assert.True(t, m.MatchString("plain first line"))
	assert.False(t, m.MatchString("  continuation"))

	m, err = NewMatcher([]string{`^ERROR:`})
	assert.NoError(t, err)
	assert.True(t, m.MatchString("plain first line before any pattern match"))
	assert.False(t, m.MatchString("  continuation"))
	assert.True(t, m.MatchString("ERROR: matched line"))
	assert.False(t, m.MatchString("plain line after a pattern has matched"))
}

func TestNewMatcherValidation(t *testing.T) {
	_, err := NewMatcher([]string{`^\d{4}-\d{2}-\d{2}`})
	assert.NoError(t, err)

	_, err = NewAutoMatcher([]string{`^APPSTART\b`})
	assert.NoError(t, err)

	_, err = NewMatcher([]string{`(?!`})
	assert.Error(t, err)

	_, err = NewAutoMatcher([]string{`(?!`})
	assert.Error(t, err)
}
