// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (

	//"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

type testCase struct {
	input    string
	expected []Item
	fail     bool
}

var cases = []struct {
	name  string
	tests []testCase
}{
	{
		name: "identifiers",
		tests: []testCase{

			{
				input: "中文 abc",
				expected: []Item{
					{ID, 0, "中文"},
					{ID, 7, "abc"},
				},
			},

			{
				input: "中文_abc",
				expected: []Item{
					{ID, 0, "中文_abc"},
				},
			},

			{
				input: "NAME👍",
				expected: []Item{
					{ID, 0, "NAME👍"},
				},
			},

			{
				input: "abc d",
				expected: []Item{
					{ID, 0, "abc"},
					{ID, 4, "d"},
				},
			},

			{
				input: "0a:bc",
				fail:  true,
			},
			{
				input: "M{cpu}",
				expected: []Item{
					{ID, 0, "M"},
					{LEFT_BRACE, 1, "{"},
					{ID, 2, "cpu"},
					{RIGHT_BRACE, 5, "}"},
				},
			},
			{
				input: "metric{cpu}",
				expected: []Item{
					{ID, 0, "metric"},
					{LEFT_BRACE, 6, "{"},
					{ID, 7, "cpu"},
					{RIGHT_BRACE, 10, "}"},
				},
			},
			{
				input: "metric{cpu(some_field, other_field)}",
				expected: []Item{
					{ID, 0, "metric"},
					{LEFT_BRACE, 6, "{"},
					{ID, 7, "cpu"},
					{LEFT_PAREN, 10, "("},
					{ID, 11, "some_field"},

					{COMMA, 21, ","},

					{ID, 23, "other_field"},
					{RIGHT_PAREN, 34, ")"},
					{RIGHT_BRACE, 35, "}"},
				},
			},
			{
				input: "{}",
				expected: []Item{
					{LEFT_BRACE, 0, "{"},
					{RIGHT_BRACE, 1, "}"},
				},
			},
			{
				input:    "logging",
				expected: []Item{{ID, 0, "logging"}},
			},
			{
				input:    "abc def",
				expected: []Item{{ID, 0, "abc"}, {ID, 4, "def"}},
			},
		},
	},

	{
		name: "common",
		tests: []testCase{
			{
				input:    ",",
				expected: []Item{{COMMA, 0, ","}},
			},
			{
				input:    "()",
				expected: []Item{{LEFT_PAREN, 0, `(`}, {RIGHT_PAREN, 1, `)`}},
			},
			{
				input:    "{}",
				expected: []Item{{LEFT_BRACE, 0, `{`}, {RIGHT_BRACE, 1, `}`}},
			},
			{
				input:    "\r\n\r",
				expected: []Item{},
			},
		},
	},
	{
		name: "numbers",
		tests: []testCase{
			{
				input:    "1",
				expected: []Item{{NUMBER, 0, "1"}},
			},
			{
				input:    "4.23",
				expected: []Item{{NUMBER, 0, "4.23"}},
			},
			/*{
				input:    ".3",
				expected: []Item{{NUMBER, 0, ".3"}},
			},
			{
				input:    "5.",
				expected: []Item{{NUMBER, 0, "5."}},
			}, */
			{
				input:    "NaN",
				expected: []Item{{NUMBER, 0, "NaN"}},
			},
			{
				input:    "nAN",
				expected: []Item{{NUMBER, 0, "nAN"}},
			},
			{
				input:    "NaN 123",
				expected: []Item{{NUMBER, 0, "NaN"}, {NUMBER, 4, "123"}},
			},
			{
				input:    "NaN123",
				expected: []Item{{ID, 0, "NaN123"}},
			},
			{
				input:    "iNf",
				expected: []Item{{NUMBER, 0, "iNf"}},
			},
			{
				input:    "Inf",
				expected: []Item{{NUMBER, 0, "Inf"}},
			},
			{
				input:    "+Inf",
				expected: []Item{{ADD, 0, "+"}, {NUMBER, 1, "Inf"}},
			},
			{
				input:    "+Inf 123",
				expected: []Item{{ADD, 0, "+"}, {NUMBER, 1, "Inf"}, {NUMBER, 5, "123"}},
			},
			{
				input:    "-Inf",
				expected: []Item{{SUB, 0, "-"}, {NUMBER, 1, "Inf"}},
			},
			{
				input:    "Infoo",
				expected: []Item{{ID, 0, "Infoo"}},
			},
			{
				input:    "-Infoo",
				expected: []Item{{SUB, 0, "-"}, {ID, 1, "Infoo"}},
			},
			{
				input:    "-Inf 123",
				expected: []Item{{SUB, 0, "-"}, {NUMBER, 1, "Inf"}, {NUMBER, 5, "123"}},
			},
			{
				input:    "0x123 0X123",
				expected: []Item{{NUMBER, 0, "0x123"}, {NUMBER, 6, "0X123"}},
			},
			{
				input:    "0123",
				expected: []Item{{NUMBER, 0, "0123"}},
			},
			{
				input:    "7.823E5",
				expected: []Item{{NUMBER, 0, "7.823E5"}},
			},
		},
	},
	{
		name: "comments",
		tests: []testCase{
			{
				input:    "# some comment",
				expected: []Item{{COMMENT, 0, "# some comment"}},
			}, {
				input: "5 # 1+1\n5",
				expected: []Item{
					{NUMBER, 0, "5"},
					{COMMENT, 2, "# 1+1"},
					{NUMBER, 8, "5"},
				},
			},
		},
	},

	{
		name: "operators",
		tests: []testCase{
			{
				input:    `=`,
				expected: []Item{{EQ, 0, `=`}},
			},
			{
				// Inside braces equality is a single '=' character but in terms of a token
				// it should be treated as ASSIGN.
				input:    `{=}`,
				expected: []Item{{LEFT_BRACE, 0, `{`}, {EQ, 1, `=`}, {RIGHT_BRACE, 2, `}`}},
			},
			{
				input:    `!=`,
				expected: []Item{{NEQ, 0, `!=`}},
			},
			{
				input:    `<`,
				expected: []Item{{LT, 0, `<`}},
			},
			{
				input:    `>`,
				expected: []Item{{GT, 0, `>`}},
			},
			{
				input:    `>=`,
				expected: []Item{{GTE, 0, `>=`}},
			},
			{
				input:    `<=`,
				expected: []Item{{LTE, 0, `<=`}},
			},
			{
				input:    `+`,
				expected: []Item{{ADD, 0, `+`}},
			},
			{
				input:    `-`,
				expected: []Item{{SUB, 0, `-`}},
			},
			{
				input:    `*`,
				expected: []Item{{MUL, 0, `*`}},
			},
			{
				input:    `/`,
				expected: []Item{{DIV, 0, `/`}},
			},
			{
				input:    `^`,
				expected: []Item{{POW, 0, `^`}},
			},
			{
				input:    `%`,
				expected: []Item{{MOD, 0, `%`}},
			},
			{
				input:    `&&`,
				expected: []Item{{AND, 0, `&&`}},
			},
			{
				input:    `||`,
				expected: []Item{{OR, 0, `||`}},
			},
		},
	},
	{
		name: "aggregators",
		tests: []testCase{
			{
				input:    `sum`,
				expected: []Item{{ID, 0, `sum`}},
			}, {
				input:    `AVG`,
				expected: []Item{{ID, 0, `AVG`}},
			}, {
				input:    `MAX`,
				expected: []Item{{ID, 0, `MAX`}},
			}, {
				input:    `min`,
				expected: []Item{{ID, 0, `min`}},
			}, {
				input:    `count`,
				expected: []Item{{ID, 0, `count`}},
			}, {
				input:    `stdvar`,
				expected: []Item{{ID, 0, `stdvar`}},
			}, {
				input:    `stddev`,
				expected: []Item{{ID, 0, `stddev`}},
			},
		},
	},
	{
		name: "selectors",
		tests: []testCase{
			{
				input: `台北`,
				// fail:  true,
				expected: []Item{
					{ID, 0, `台北`},
				},
			},
			{
				input: `{台北='a'}`,
				expected: []Item{
					{LEFT_BRACE, 0, `{`},
					{ID, 1, `台北`},
					{EQ, 7, `=`},
					{STRING, 8, `'a'`},
					{RIGHT_BRACE, 11, `}`},
				},
				// fail:  true,
			},
			{
				input: `{0a='a'}`,
				fail:  true,
			},
			{
				input: `{foo='bar'}`,
				expected: []Item{
					{LEFT_BRACE, 0, `{`},
					{ID, 1, `foo`},
					{EQ, 4, `=`},
					{STRING, 5, `'bar'`},
					{RIGHT_BRACE, 10, `}`},
				},
			},
			{
				input: `{foo="bar"}`,
				expected: []Item{
					{LEFT_BRACE, 0, `{`},
					{ID, 1, `foo`},
					{EQ, 4, `=`},
					{STRING, 5, `"bar"`},
					{RIGHT_BRACE, 10, `}`},
				},
			},
			{
				input: `{foo="bar\"bar"}`,
				expected: []Item{
					{LEFT_BRACE, 0, `{`},
					{ID, 1, `foo`},
					{EQ, 4, `=`},
					{STRING, 5, `"bar\"bar"`},
					{RIGHT_BRACE, 15, `}`},
				},
			},
			{
				input: `{NaN	!= "bar" }`,
				expected: []Item{
					{LEFT_BRACE, 0, `{`},
					{NUMBER, 1, `NaN`},
					{NEQ, 5, `!=`},
					{STRING, 8, `"bar"`},
					{RIGHT_BRACE, 14, `}`},
				},
			},
			{
				input: `{alert!#"bar"}`, fail: true,
			},
		},
	},
	{
		name: "common errors",
		tests: []testCase{
			{
				input: `!(`, fail: true,
			},
			{
				input: "1a", fail: true,
			},
		},
	},
	{
		name: "mismatched parentheses",
		tests: []testCase{
			{
				input: `(`, fail: true,
			}, {
				input: `())`, fail: true,
			}, {
				input: `(()`, fail: true,
			}, {
				input: `{`, fail: true,
			}, {
				input: `}`, fail: true,
			}, {
				input: "{{", fail: true,
			}, {
				input: `[`, fail: true,
			}, {
				input: `[[`, fail: true,
			}, {
				input: `[]]`, fail: true,
			}, {
				input: `]`, fail: true,
			},
		},
	},

	{
		name: "encoding issues",
		tests: []testCase{
			{
				input: "\"\xfff\"",
				expected: []Item{
					{STRING, 0, "\"\xfff\""},
				},
			},
			{
				input: "`\xff`", // invalid UTF-8 rune
				fail:  true,
			},

			{
				input: "\xff",
				expected: []Item{
					{ID, 0, "\xff"},
				},
			},
		},
	},
	{
		name: "strings",
		tests: []testCase{
			{
				input:    "\"test\\tsequence\"",
				expected: []Item{{STRING, 0, `"test\tsequence"`}},
			},
			{
				input:    "\"test\\\\.expression1\"",
				expected: []Item{{STRING, 0, `"test\\.expression1"`}},
			},
			{
				input: "\"test\\.expression2\"",
				expected: []Item{
					{ERROR, 0, "unknown escape sequence U+002E '.'"},
					{STRING, 0, `"test\.expression2"`},
				},
			},
			{
				input:    "`test\\.expression3`",
				expected: []Item{{QUOTED_STRING, 0, "`test\\.expression3`"}},
			},
			{
				input:    "\".٩\"",
				expected: []Item{{STRING, 0, `".٩"`}},
			},

			{
				input:    "\"中文\"",
				expected: []Item{{STRING, 0, `"中文"`}},
			},

			{
				input:    "`中文`",
				expected: []Item{{QUOTED_STRING, 0, "`中文`"}},
			},
			{
				input:    "''",
				expected: []Item{{STRING, 0, "''"}},
			},

			{
				input:    `""`,
				expected: []Item{{STRING, 0, `""`}},
			},

			{
				input: `'''this
is
'
string'''`,
				expected: []Item{
					{MULTILINE_STRING, 0, `'''this
is
'
string'''`},
				},
			},
		},
	},
}

// TestLexer tests basic functionality of the lexer. More elaborate tests are implemented
// for the parser to avoid duplicated effort.
func TestLexer(t *testing.T) {
	for _, typ := range cases {
		t.Run(typ.name, func(t *testing.T) {
			for i, test := range typ.tests {
				l := &Lexer{
					input: test.input,
				}

				t.Logf("%s[%d]: input %q", typ.name, i, test.input)

				var out []Item

				for l.state = lexStatements; l.state != nil; {
					out = append(out, Item{})

					l.NextItem(&out[len(out)-1])
				}

				lastItem := out[len(out)-1] // ignore last EOF item
				if test.fail {
					hasError := false
					for _, item := range out {
						if item.Typ == ERROR {
							hasError = true
						}
					}
					if !hasError {
						t.Fatalf("expected lexing error but did not fail: got\n%s", outStr(out))
					}
					continue
				}

				if lastItem.Typ == ERROR { // ERROR usually be the last item
					t.Logf("%d: input %q", i, test.input)
					t.Fatalf("unexpected lexing error at position %d: %s", lastItem.Pos, lastItem)
				}

				eofItem := Item{EOF, Pos(len(test.input)), ""}
				testutil.Equals(t, eofItem, lastItem)

				out = out[:len(out)-1]
				testutil.Equals(t, test.expected, out)
			}
		})
	}
}

func outStr(items []Item) string {
	strs := []string{}
	for _, item := range items {
		strs = append(strs, item.lexStr())
	}

	return strings.Join(strs, "\n")
}

func TestUTF8(t *testing.T) {
	cases := []struct {
		in []byte
	}{
		{
			in: []byte("abc中文def"),
		},
	}

	for _, tc := range cases {
		for len(tc.in) > 0 {
			r, n := utf8.DecodeRune(tc.in)
			t.Logf("r: %c(%02x, %d), n: %d", r, r, utf8.RuneLen(r), n)
			tc.in = tc.in[n:]
		}
	}
}
