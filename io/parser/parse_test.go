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
			in: `{abc notin [1.1,1.2,1.3] and (a > 1 || c< 0)}`,
			expected: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op: AND,
							RHS: &ParenExpr{
								Param: &BinaryExpr{
									Op: OR,
									LHS: &BinaryExpr{
										Op:  GT,
										LHS: &Identifier{Name: "a"},
										RHS: &NumberLiteral{IsInt: true, Int: 1},
									},
									RHS: &BinaryExpr{
										Op:  LT,
										LHS: &Identifier{Name: "c"},
										RHS: &NumberLiteral{IsInt: true, Int: 0},
									},
								},
							},
							LHS: &BinaryExpr{
								Op:  NOTIN,
								LHS: &Identifier{Name: "abc"},
								RHS: NodeList{
									&NumberLiteral{Float: 1.1},
									&NumberLiteral{Float: 1.2},
									&NumberLiteral{Float: 1.3},
								},
							},
						},
					},
				},
			},
		},

		{
			in: `{abc notin [1.1,1.2,1.3]}`,
			expected: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op:  NOTIN,
							LHS: &Identifier{Name: "abc"},
							RHS: NodeList{
								&NumberLiteral{Float: 1.1},
								&NumberLiteral{Float: 1.2},
								&NumberLiteral{Float: 1.3},
							},
						},
					},
				},
			},
		},

		{
			in:       `{};{};{}`,
			expected: WhereConditions{&WhereCondition{}, &WhereCondition{}, &WhereCondition{}},
		},

		{
			in: `;;;;;{a>1};;;; {b>1};;;;`, // multiple conditions
			expected: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op:  GT,
							LHS: &Identifier{Name: "a"},
							RHS: &NumberLiteral{IsInt: true, Int: 1},
						},
					},
				},
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op:  GT,
							LHS: &Identifier{Name: "b"},
							RHS: &NumberLiteral{IsInt: true, Int: 1},
						},
					},
				},
			},
		},

		{
			in: `{source = 'http_dial_testing' and ( aaaa in ['aaaa44', 'gaga']  and  city in ['北京'] )}`,
			expected: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op: AND,

							LHS: &BinaryExpr{
								Op:  EQ,
								LHS: &Identifier{Name: "source"},
								RHS: &StringLiteral{Val: "http_dial_testing"},
							},

							RHS: &ParenExpr{
								Param: &BinaryExpr{
									Op: AND,

									LHS: &BinaryExpr{
										Op:  IN,
										LHS: &Identifier{Name: "aaaa"},
										RHS: &NodeList{
											&StringLiteral{Val: "aaaa44"},
											&StringLiteral{Val: "gaga"},
										},
									},
									RHS: &BinaryExpr{
										Op:  IN,
										LHS: &Identifier{Name: "city"},
										RHS: &NodeList{
											&StringLiteral{Val: "北京"},
										},
									},
								},
							},
						},
					},
				},
			},
		},

		{
			in: `{source = 'http_dial_testing' and  aaaa in ['aaaa44', 'gaga']  and  city in ['北京'] }`,
			expected: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op: AND,
							LHS: &BinaryExpr{
								Op: AND,
								LHS: &BinaryExpr{
									Op:  EQ,
									LHS: &Identifier{Name: "source"},
									RHS: &StringLiteral{Val: "http_dial_testing"},
								},
								RHS: &BinaryExpr{
									Op:  IN,
									LHS: &Identifier{Name: "aaaa"},
									RHS: &NodeList{
										&StringLiteral{Val: "aaaa44"},
										&StringLiteral{Val: "gaga"},
									},
								},
							},
							RHS: &BinaryExpr{
								Op:  IN,
								LHS: &Identifier{Name: "city"},
								RHS: &NodeList{
									&StringLiteral{Val: "北京"},
								},
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

		switch v := p.parseResult.(type) {
		case WhereConditions:

			exp, ok := tc.expected.(WhereConditions)
			if !ok {
				t.Fatal("not WhereConditions")
			}

			x := exp.String()
			y := v.String()

			tu.Equals(t, x, y)
			t.Logf("[ok] in: %s, exp: %s", x, y)
		default:
			t.Fatalf("should not been here: %s", reflect.TypeOf(p.parseResult).String())
		}
	}
}
