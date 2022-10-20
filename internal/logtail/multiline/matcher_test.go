// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package multiline wrap regexp/match functions
package multiline

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	type in struct {
		content string
		match   bool
	}

	cases := []struct {
		name     string
		ins      []in
		topScore int
	}{
		{
			name: "time.RFC3339 1st",
			ins: []in{
				{
					// topScore +1
					"2022-08-15T15:04:05Z08:00  INFO  cmd/main.go  Running",
					true,
				},
			},
			topScore: 1,
		},
		{
			name: "time.RFC3339 2st",
			ins: []in{
				{
					// topScore +1
					"2022-08-15T15:04:05Z08:00  INFO  cmd/main.go  Running",
					true,
				},
				{
					// topScore +1
					"2022-08-15T15:04:25Z08:00  WARN  pkg/test.go  error",
					true,
				},
			},
			topScore: 2,
		},
		{
			name: "time.RFC3339 and time.RubyDate 1st",
			ins: []in{
				{
					// topScore +1
					"2022-08-15T15:04:05Z08:00  INFO  cmd/main.go  Running",
					true,
				},
				{
					// topScore +1
					"2022-08-15T15:04:25Z08:00  WARN  pkg/test.go  error",
					true,
				},
				{
					"Mon Jan 02 15:04:05 +0800 2022  INFO  cmd/main.go  Running",
					true,
				},
			},
			topScore: 2,
		},
		{
			name: "mixing",
			ins: []in{
				// 在此之前没有任何一条匹配成功，所以此条使用的默认匹配行为
				{
					"panic: runtime error: invalid memory address or nil pointer dereference",
					true,
				},
				// 同上
				{
					"[signal SIGSEGV: segmentation violation code=0x1 addr=0x90 pc=0xa2ed0]",
					true,
				},
				// 匹配成功
				{
					// topScore +1
					"2022-08-15T15:04:25Z08:00  WARN  pkg/test.go  error",
					true,
				},
				// 已经存在匹配成功，不再使用默认匹配行为
				{
					"panic: runtime error: invalid memory address or nil pointer dereference",
					false,
				},
				{
					// topScore +1
					"2022-08-15T15:04:25Z08:00  WARN  pkg/test.go  error",
					true,
				},
				{
					"[signal SIGSEGV: segmentation violation code=0x1 addr=0x90 pc=0xa2ed0]",
					false,
				},
			},
			topScore: 2,
		},
		{
			name: "empty-string",
			ins: []in{
				// 空字符串没有匹配意义，所以匹配失败
				{
					"",
					false,
				},
			},
			topScore: 0,
		},
	}

	for _, tc := range cases {
		m, err := NewMatcher(GlobalPatterns)
		assert.NoError(t, err)

		t.Run(tc.name, func(t *testing.T) {
			for _, in := range tc.ins {
				match := m.Match([]byte(in.content))
				assert.Equal(t, in.match, match)
			}

			assert.Equal(t, tc.topScore, m.patterns[0].score)
		})
	}
}

func TestMatchNoPattern(t *testing.T) {
	cases := []struct {
		in      string
		matched bool
	}{
		{
			in:      "2022-08-15T15:04:05Z08:00  INFO  cmd/main.go  Running",
			matched: true,
		},
		{
			in:      "  INFO  cmd/main.go  Running",
			matched: false,
		},
	}

	for _, tc := range cases {
		m, err := NewMatcher(nil)
		assert.NoError(t, err)

		t.Run("", func(t *testing.T) {
			match := m.Match([]byte(tc.in))
			assert.Equal(t, tc.matched, match)
		})
	}
}

func TestNewMatcher(t *testing.T) {
	t.Run("ok-patterns", func(t *testing.T) {
		patterns := []string{
			`^\d+-\d+-\d+T\d+:\d+:\d+(\.\d+)?(Z\d*:?\d*)?`,
			`^[A-Za-z_]+ [A-Za-z_]+ +\d+ \d+:\d+:\d+ \d+`,
		}
		_, err := NewMatcher(patterns)
		assert.NoError(t, err)
	})

	t.Run("ok-global-patterns", func(t *testing.T) {
		patterns := GlobalPatterns
		_, err := NewMatcher(patterns)
		assert.NoError(t, err)
	})

	t.Run("ok-nopattern", func(t *testing.T) {
		patterns := []string{}
		_, err := NewMatcher(patterns)
		assert.NoError(t, err)
	})

	t.Run("ok-nopattern-2", func(t *testing.T) {
		_, err := NewMatcher(nil)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		patterns := []string{
			`^\d+-\d+-\d+T\d+:\d+:\d+(\.\d+)?(Z\d*:?\d*)?`,
			`(?!`, // error
		}
		_, err := NewMatcher(patterns)
		assert.Error(t, err)
	})
}

func TestScorePatternString(t *testing.T) {
	t.Run("", func(t *testing.T) {
		sc := &scoredPattern{
			score:  1,
			regexp: regexp.MustCompile(`^\S`),
		}
		assert.Equal(t, "score:1, regexp:^\\S", sc.String())
	})
}
