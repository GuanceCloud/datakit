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
				input: "a:bc",
				expected: []Item{
					{ID, 0, "a"},
					{COLON, 1, ":"},
					{ID, 2, "bc"},
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
				input: ":bc",
				expected: []Item{
					{COLON, 0, ":"},
					{ID, 1, "bc"},
				},
			},
			{
				input: "0a:bc",
				fail:  true,
			},
			{
				input: "metric_a:field_x",
				expected: []Item{
					{ID, 0, "metric_a"},
					{COLON, 8, ":"},
					{ID, 9, "field_x"},
				},
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
				input: "metric{cpu:some_field}, metric{cpu:other_field}",
				expected: []Item{
					{ID, 0, "metric"},
					{LEFT_BRACE, 6, "{"},
					{ID, 7, "cpu"},
					{COLON, 10, ":"},
					{ID, 11, "some_field"},
					{RIGHT_BRACE, 21, "}"},

					{COMMA, 22, ","},

					{ID, 24, "metric"},
					{LEFT_BRACE, 30, "{"},
					{ID, 31, "cpu"},
					{COLON, 34, ":"},
					{ID, 35, "other_field"},
					{RIGHT_BRACE, 46, "}"},
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
				input:    "abc:def",
				expected: []Item{{ID, 0, "abc"}, {COLON, 3, ":"}, {ID, 4, "def"}},
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
				input: "[5m]",
				expected: []Item{
					{LEFT_BRACKET, 0, `[`},
					{DURATION, 1, `5m`},
					{RIGHT_BRACKET, 3, `]`},
				},
			},
			{
				input: "[ 5m]",
				expected: []Item{
					{LEFT_BRACKET, 0, `[`},
					{DURATION, 2, `5m`},
					{RIGHT_BRACKET, 4, `]`},
				},
			},
			{
				input: "[  5m]",
				expected: []Item{
					{LEFT_BRACKET, 0, `[`},
					{DURATION, 3, `5m`},
					{RIGHT_BRACKET, 5, `]`},
				},
			},
			{
				input: "[  5m ]",
				expected: []Item{
					{LEFT_BRACKET, 0, `[`},
					{DURATION, 3, `5m`},
					{RIGHT_BRACKET, 6, `]`},
				},
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
				// See https://github.com/prometheus/prometheus/issues/939.
				input: ".٩",
				fail:  true,
			},
		},
	},
	{
		name: "durations",
		tests: []testCase{
			{
				input: "[5m]",
				expected: []Item{
					{LEFT_BRACKET, 0, `[`},
					{DURATION, 1, `5m`},
					{RIGHT_BRACKET, 3, `]`},
				},
			},
			{
				input: "[10m:5m:30s]",
				expected: []Item{
					{LEFT_BRACKET, 0, `[`},
					{DURATION, 1, `10m`},
					{COLON, 4, `:`},
					{DURATION, 5, `5m`},
					{COLON, 7, `:`},
					{DURATION, 8, `30s`},
					{RIGHT_BRACKET, 11, `]`},
				},
			},
			{
				input:    "5s",
				expected: []Item{{DURATION, 0, "5s"}},
			}, {
				input:    "123m",
				expected: []Item{{DURATION, 0, "123m"}},
			}, {
				input:    "1h",
				expected: []Item{{DURATION, 0, "1h"}},
			}, {
				input:    "3w",
				expected: []Item{{DURATION, 0, "3w"}},
			}, {
				input:    "1y",
				expected: []Item{{DURATION, 0, "1y"}},
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
			}, {
				input:    `<`,
				expected: []Item{{LT, 0, `<`}},
			}, {
				input:    `>`,
				expected: []Item{{GT, 0, `>`}},
			}, {
				input:    `>=`,
				expected: []Item{{GTE, 0, `>=`}},
			}, {
				input:    `<=`,
				expected: []Item{{LTE, 0, `<=`}},
			}, {
				input:    `+`,
				expected: []Item{{ADD, 0, `+`}},
			}, {
				input:    `-`,
				expected: []Item{{SUB, 0, `-`}},
			}, {
				input:    `*`,
				expected: []Item{{MUL, 0, `*`}},
			}, {
				input:    `/`,
				expected: []Item{{DIV, 0, `/`}},
			}, {
				input:    `^`,
				expected: []Item{{POW, 0, `^`}},
			}, {
				input:    `%`,
				expected: []Item{{MOD, 0, `%`}},
			}, {
				input:    `AND`,
				expected: []Item{{AND, 0, `AND`}},
			}, {
				input:    `or`,
				expected: []Item{{OR, 0, `or`}},
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
		name: "keywords",
		tests: []testCase{
			{
				input:    "offset",
				expected: []Item{{OFFSET, 0, "offset"}},
			},
			{
				input:    "by",
				expected: []Item{{BY, 0, "by"}},
			},
			{
				input:    "as",
				expected: []Item{{AS, 0, "as"}},
			},
			{
				input:    "asc",
				expected: []Item{{ASC, 0, "asc"}},
			},
			{
				input: "a in b",
				expected: []Item{
					{ID, 0, "a"},
					{IN, 2, "in"},
					{ID, 5, "b"},
				},
			},
		},
	},
	{
		name: "selectors",
		tests: []testCase{
			{
				input: `台北`,
				fail:  true,
			}, {
				input: `{台北='a'}`,
				fail:  true,
			}, {
				input: `{0a='a'}`,
				fail:  true,
			}, {
				input: `{foo='bar'}`,
				expected: []Item{
					{LEFT_BRACE, 0, `{`},
					{ID, 1, `foo`},
					{EQ, 4, `=`},
					{STRING, 5, `'bar'`},
					{RIGHT_BRACE, 10, `}`},
				},
			}, {
				input: `{foo="bar"}`,
				expected: []Item{
					{LEFT_BRACE, 0, `{`},
					{ID, 1, `foo`},
					{EQ, 4, `=`},
					{STRING, 5, `"bar"`},
					{RIGHT_BRACE, 10, `}`},
				},
			}, {
				input: `{foo="bar\"bar"}`,
				expected: []Item{
					{LEFT_BRACE, 0, `{`},
					{ID, 1, `foo`},
					{EQ, 4, `=`},
					{STRING, 5, `"bar\"bar"`},
					{RIGHT_BRACE, 15, `}`},
				},
			}, {
				input: `{NaN	!= "bar" }`,
				expected: []Item{
					{LEFT_BRACE, 0, `{`},
					{NUMBER, 1, `NaN`},
					{NEQ, 5, `!=`},
					{STRING, 8, `"bar"`},
					{RIGHT_BRACE, 14, `}`},
				},
			}, {
				input: `{alert!#"bar"}`, fail: true,
			},
			{
				input: `{foo:a="bar"}`,
				expected: []Item{
					{LEFT_BRACE, 0, "{"},
					{ID, 1, "foo"},
					{COLON, 4, ":"},
					{ID, 5, "a"},
					{EQ, 6, "="},
					{STRING, 7, "\"bar\""},
					{RIGHT_BRACE, 12, "}"},
				},
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
			//{
			//	input: "\"\xfff\"", fail: true,
			//},
			{
				input: "`\xff`", fail: true,
			},
			{
				input: "\xff", fail: true,
			},
		},
	},
	{
		name: "subqueries",
		tests: []testCase{
			{
				input: `test_name{no!="bar"}[4m:4s]`,
				expected: []Item{
					{ID, 0, `test_name`},
					{LEFT_BRACE, 9, `{`},
					{ID, 10, `no`},
					{NEQ, 12, `!=`},
					{STRING, 14, `"bar"`},
					{RIGHT_BRACE, 19, `}`},
					{LEFT_BRACKET, 20, `[`},
					{DURATION, 21, `4m`},
					{COLON, 23, `:`},
					{DURATION, 24, `4s`},
					{RIGHT_BRACKET, 26, `]`},
				},
			},
			{
				input: `test:name{no!="bar"}[4m:4s]`,
				expected: []Item{
					{ID, 0, `test`},
					{COLON, 4, `:`},
					{ID, 5, `name`},
					{LEFT_BRACE, 9, `{`},
					{ID, 10, `no`},
					{NEQ, 12, `!=`},
					{STRING, 14, `"bar"`},
					{RIGHT_BRACE, 19, `}`},
					{LEFT_BRACKET, 20, `[`},
					{DURATION, 21, `4m`},
					{COLON, 23, `:`},
					{DURATION, 24, `4s`},
					{RIGHT_BRACKET, 26, `]`},
				},
			},
			{
				input: `test:name{no!="b:ar"}[4m:4s]`,
				expected: []Item{
					{ID, 0, `test`},
					{COLON, 4, `:`},
					{ID, 5, `name`},
					{LEFT_BRACE, 9, `{`},
					{ID, 10, `no`},
					{NEQ, 12, `!=`},
					{STRING, 14, `"b:ar"`},
					{RIGHT_BRACE, 20, `}`},
					{LEFT_BRACKET, 21, `[`},
					{DURATION, 22, `4m`},
					{COLON, 24, `:`},
					{DURATION, 25, `4s`},
					{RIGHT_BRACKET, 27, `]`},
				},
			},
			{
				input: `test:name{no!="b:ar"}[4m:]`,
				expected: []Item{
					{ID, 0, `test`},
					{COLON, 4, `:`},
					{ID, 5, `name`},
					{LEFT_BRACE, 9, `{`},
					{ID, 10, `no`},
					{NEQ, 12, `!=`},
					{STRING, 14, `"b:ar"`},
					{RIGHT_BRACE, 20, `}`},
					{LEFT_BRACKET, 21, `[`},
					{DURATION, 22, `4m`},
					{COLON, 24, `:`},
					{RIGHT_BRACKET, 25, `]`},
				},
			},
			{ // Nested Subquery.
				input: `min_over_time(rate(foo{bar="baz"}[2s])[5m:])[4m:3s]`,
				expected: []Item{
					{ID, 0, `min_over_time`},
					{LEFT_PAREN, 13, `(`},
					{ID, 14, `rate`},
					{LEFT_PAREN, 18, `(`},
					{ID, 19, `foo`},
					{LEFT_BRACE, 22, `{`},
					{ID, 23, `bar`},
					{EQ, 26, `=`},
					{STRING, 27, `"baz"`},
					{RIGHT_BRACE, 32, `}`},
					{LEFT_BRACKET, 33, `[`},
					{DURATION, 34, `2s`},
					{RIGHT_BRACKET, 36, `]`},
					{RIGHT_PAREN, 37, `)`},
					{LEFT_BRACKET, 38, `[`},
					{DURATION, 39, `5m`},
					{COLON, 41, `:`},
					{RIGHT_BRACKET, 42, `]`},
					{RIGHT_PAREN, 43, `)`},
					{LEFT_BRACKET, 44, `[`},
					{DURATION, 45, `4m`},
					{COLON, 47, `:`},
					{DURATION, 48, `3s`},
					{RIGHT_BRACKET, 50, `]`},
				},
			},
			// Subquery with offset.
			{
				input: `test:name{no!="b:ar"}[4m:4s] offset 10m`,
				expected: []Item{
					{ID, 0, `test`},
					{COLON, 4, `:`},
					{ID, 5, `name`},
					{LEFT_BRACE, 9, `{`},
					{ID, 10, `no`},
					{NEQ, 12, `!=`},
					{STRING, 14, `"b:ar"`},
					{RIGHT_BRACE, 20, `}`},
					{LEFT_BRACKET, 21, `[`},
					{DURATION, 22, `4m`},
					{COLON, 24, `:`},
					{DURATION, 25, `4s`},
					{RIGHT_BRACKET, 27, `]`},
					{OFFSET, 29, "offset"},
					{DURATION, 36, "10m"},
				},
			},
			{
				input: `min_over_time(rate(foo{bar="baz"}[2s])[5m:] offset 6m)[4m:3s]`,
				expected: []Item{

					{ID, 0, `min_over_time`},
					{LEFT_PAREN, 13, `(`},
					{ID, 14, `rate`},
					{LEFT_PAREN, 18, `(`},
					{ID, 19, `foo`},
					{LEFT_BRACE, 22, `{`},
					{ID, 23, `bar`},
					{EQ, 26, `=`},
					{STRING, 27, `"baz"`},
					{RIGHT_BRACE, 32, `}`},
					{LEFT_BRACKET, 33, `[`},
					{DURATION, 34, `2s`},
					{RIGHT_BRACKET, 36, `]`},
					{RIGHT_PAREN, 37, `)`},
					{LEFT_BRACKET, 38, `[`},
					{DURATION, 39, `5m`},
					{COLON, 41, `:`},
					{RIGHT_BRACKET, 42, `]`},
					{OFFSET, 44, `offset`},
					{DURATION, 51, `6m`},
					{RIGHT_PAREN, 53, `)`},
					{LEFT_BRACKET, 54, `[`},
					{DURATION, 55, `4m`},
					{COLON, 57, `:`},
					{DURATION, 58, `3s`},
					{RIGHT_BRACKET, 60, `]`},
				},
			},
			{
				input: `test:name[ 5m]`,
				expected: []Item{
					{ID, 0, `test`},
					{COLON, 4, `:`},
					{ID, 5, `name`},
					{LEFT_BRACKET, 9, `[`},
					{DURATION, 11, `5m`},
					{RIGHT_BRACKET, 13, `]`},
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
				testutil.Equals(t, lastItem, eofItem)

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
