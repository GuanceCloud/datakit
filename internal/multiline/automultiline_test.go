// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package multiline wrap regexp/match functions
package multiline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoMultilineProcess(t *testing.T) {
	type in struct {
		content string
		match   bool
	}

	cases := []struct {
		name             string
		ins              []in
		topScore         int
		topScoreNotEqual bool
	}{
		{
			name: "time.RFC3339 1st",
			ins: []in{
				{
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
					"2022-08-15T15:04:05Z08:00  INFO  cmd/main.go  Running",
					true,
				},
				{
					"2022-08-15T15:04:25Z08:00  WARN  pkg/test.go  error",
					true,
				},
			},
			topScore:         3,
			topScoreNotEqual: true,
		},
		{
			name: "time.RFC3339 and time.RubyDate 1st",
			ins: []in{
				{
					"2022-08-15T15:04:05Z08:00  INFO  cmd/main.go  Running",
					true,
				},
				{
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
			name: "nomatch->match",
			ins: []in{
				{
					"panic: runtime error: invalid memory address or nil pointer dereference",
					true,
				},
				{
					"[signal SIGSEGV: segmentation violation code=0x1 addr=0x90 pc=0xa2ed0]",
					true,
				},
				{
					"2022-08-15T15:04:25Z08:00  WARN  pkg/test.go  error",
					true,
				},
				{
					"panic: runtime error: invalid memory address or nil pointer dereference",
					false,
				},
				{
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
	}

	for _, tc := range cases {
		m, err := NewAutoMultiline(GlobalPatterns)
		assert.NoError(t, err)

		t.Run(tc.name, func(t *testing.T) {
			for _, in := range tc.ins {
				match := m.Match([]byte(in.content))
				assert.Equal(t, in.match, match)
			}

			if tc.topScoreNotEqual {
				assert.NotEqual(t, tc.topScore, m.patterns[0].score)
			} else {
				assert.Equal(t, tc.topScore, m.patterns[0].score)
			}
		})
	}
}

func TestNewAutoMultiline(t *testing.T) {
	t.Run("ok 1", func(t *testing.T) {
		patterns := []string{
			`^\d+-\d+-\d+T\d+:\d+:\d+(\.\d+)?(Z\d*:?\d*)?`,
			`^[A-Za-z_]+ [A-Za-z_]+ +\d+ \d+:\d+:\d+ \d+`,
		}
		_, err := NewAutoMultiline(patterns)
		assert.NoError(t, err)
	})

	t.Run("ok 2", func(t *testing.T) {
		patterns := GlobalPatterns
		_, err := NewAutoMultiline(patterns)
		assert.NoError(t, err)
	})

	t.Run("error 1", func(t *testing.T) {
		patterns := []string{}
		_, err := NewAutoMultiline(patterns)
		assert.Error(t, err)
	})

	t.Run("error 2", func(t *testing.T) {
		patterns := []string{
			`^\d+-\d+-\d+T\d+:\d+:\d+(\.\d+)?(Z\d*:?\d*)?`,
			`(?!`, // error
		}
		_, err := NewAutoMultiline(patterns)
		assert.Error(t, err)
	})
}
