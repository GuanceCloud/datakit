// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parser

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestUnquote(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		expect string
		fail   bool
	}{
		{
			name:   "str",
			s:      `"abc"`,
			expect: "abc",
		},
		{
			name:   "str-with-special-char",
			s:      `"abc\""`,
			expect: "abc\"",
		},
		{
			name:   "str-with-unicode",
			s:      `"中\\abc"`,
			expect: "中\\abc",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			x, err := Unquote(tc.s)
			if tc.fail {
				tu.Assert(t, err != nil, "")
				return
			}

			tu.Assert(t, err == nil, "expect nil, got %s", err)
			tu.Equals(t, tc.expect, x)
		})
	}
}

func TestMultilineUnquote(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		expect string
		fail   bool
	}{
		{
			name:   "empty-multiline-str",
			s:      `""""""`,
			expect: "",
		},

		{
			name:   "unmatched-multiline-str",
			s:      `"""abc""`,
			expect: "",
			fail:   true,
		},

		{
			name:   "unmatched-multiline-str2",
			s:      `"""abc""'`,
			expect: "",
			fail:   true,
		},

		{
			name: "empty-multiline-str",
			s: `"""
abc
"""`,
			expect: `
abc
`,
		},

		{
			name:   "empty-multiline-str-single-quote",
			s:      `''''''`,
			expect: "",
		},

		{
			name:   "multiline-str-with-1-line",
			s:      `"""abc def"""`,
			expect: "abc def",
		},
		{
			name: "multiline-str-only-new-line",
			s: `"""
"""`,
			expect: "\n",
		},

		{
			name: "multiline-str",
			s: `"""this is
multiline
string"""`,
			expect: "this is\nmultiline\nstring",
		},
		{
			name: "multiline-str-with-special-chars",
			s: `"""this is
multiline ""'\
string"""`,
			expect: "this is\nmultiline \"\"'\\\nstring",
		},

		{
			name: "multiline-str-with-unicode",
			s: `"""中\n\n\n
文"""`,
			expect: "中\\n\\n\\n\n文",
		},

		{
			name:   "multiline-str-with-`",
			s:      "'''中`" + "\n" + "文'''",
			expect: "中`\n文",
		},

		{
			name: "multiline-str-with-escaped-char",
			s: `'''
\r\n
\b\c'''`,
			expect: `
\r\n
\b\c`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			x, err := UnquoteMultiline(tc.s)
			if tc.fail {
				tu.Assert(t, err != nil, "")
				t.Logf("[expect] %s => %s", tc.s, err)
				return
			}

			tu.Assert(t, err == nil, "expect nil, got %s", err)
			tu.Equals(t, tc.expect, x)
		})
	}
}
