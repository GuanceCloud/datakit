package parser

import (
	"reflect"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in       string
		expected interface{}
		fail     bool
	}{
		{
			in: `{a > 0, c < 5 || d < 2}`,
			// in: `{'http_dial_testing'  and ( aaaa in ['aaaa44', 'gaga']  and  city in ['北京'] )}`,
			expected: WhereCondition{
				conditions: []Node{
					&BinaryExpr{
						Op: AND,
						LHS: &BinaryExpr{
							Op:  GT,
							LHS: &Identifier{Name: "a"},
							RHS: &NumberLiteral{IsInt: true, Int: 0},
						},

						RHS: &BinaryExpr{
							Op: OR,
							LHS: &BinaryExpr{

								Op:  LT,
								LHS: &Identifier{Name: "c"},
								RHS: &NumberLiteral{IsInt: true, Int: 5},
							},

							RHS: &BinaryExpr{
								Op:  LT,
								LHS: &Identifier{Name: "d"},
								RHS: &NumberLiteral{IsInt: true, Int: 2},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		var err error
		p := newParser(tc.in)
		defer parserPool.Put(p)
		defer p.recover(&err)

		p.doParse()

		if tc.fail {
			tu.Assert(t, len(p.errs) > 0, "")
			continue
		}

		switch w := p.parseResult.(type) {
		case *WhereCondition:
			exp := tc.expected.(WhereCondition)
			x, y := exp.String(), w.String()
			tu.Equals(t, x, y)
			t.Logf("in: %s, exp: %s", x, y)
		default:
			t.Fatalf("should not been here: %s", reflect.TypeOf(p.parseResult).String())
		}
	}
}
