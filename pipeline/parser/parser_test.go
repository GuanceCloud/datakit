package parser

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

type parseCase struct {
	in       string
	expected *Ast
	err      string
	fail     bool
}

var parseCases = []*parseCase{

	// case:
	{
		in: `f([2].x[3])`,
		expected: &Ast{
			Functions: []*FuncExpr{
				{
					Name: `f`,
					Param: []Node{
						&AttrExpr{
							Obj: &IndexExpr{Index: []int64{2}},
							Attr: &IndexExpr{
								Obj:   &Identifier{Name: "x"},
								Index: []int64{3},
							},
						},
					},
				},
			},
		},
	},

	// case: multi-dim arr
	{
		in:   `f(x.y[2.5])`,
		fail: true,
	},

	{
		in: `f(x.y[1][2].z)`,
		expected: &Ast{
			Functions: []*FuncExpr{
				{
					Name: `f`,
					Param: []Node{
						&AttrExpr{
							Obj: &Identifier{Name: "x"},
							Attr: &AttrExpr{
								Obj: &IndexExpr{
									Obj:   &Identifier{Name: "y"},
									Index: []int64{1, 2},
								},
								Attr: &Identifier{Name: "z"},
							},
						},
					},
				},
			},
		},
	},

	// case: multiple functions
	{
		in: `f1()
		f2()
		f3()`,

		expected: &Ast{
			Functions: []*FuncExpr{
				{
					Name: `f1`,
				},
				{
					Name: `f2`,
				},
				{
					Name: `f3`,
				},
			},
		},
	},

	// case: embeded functions
	{
		in: `f1(g(f2("abc"), 123), 1,2,3)`,

		expected: &Ast{
			Functions: []*FuncExpr{
				{
					Name: "f1",
					Param: []Node{
						&FuncExpr{
							Name: "g",
							Param: []Node{
								&FuncExpr{
									Name:  "f2",
									Param: []Node{&StringLiteral{Val: "abc"}},
								},
								&NumberLiteral{IsInt: true, Int: 123},
							},
						},

						&NumberLiteral{IsInt: true, Int: 1},
						&NumberLiteral{IsInt: true, Int: 2},
						&NumberLiteral{IsInt: true, Int: 3},
					},
				},
			},
		},
	},

	// case: attr syntax in function arg
	{
		in: `avg(x.y.z, 1,2,3, p68, re("cd"), pqa)`,
		expected: &Ast{
			Functions: []*FuncExpr{
				{
					Name: "avg",
					Param: []Node{

						&AttrExpr{
							Obj: &Identifier{Name: "x"},
							Attr: &AttrExpr{
								Obj:  &Identifier{Name: "y"},
								Attr: &Identifier{Name: "z"},
							},
						},

						&NumberLiteral{IsInt: true, Int: 1},
						&NumberLiteral{IsInt: true, Int: 2},
						&NumberLiteral{IsInt: true, Int: 3},

						&Identifier{Name: "p68"},
						&Regex{Regex: "cd"},
						&Identifier{Name: "pqa"},
					},
				},
			},
		},
	},

	// case: attr syntax with index syntax in function arg
	{
		in: `json(_, x.y[1].z)`,
		expected: &Ast{
			Functions: []*FuncExpr{
				{
					Name: "json",
					Param: []Node{
						&Identifier{Name: "_"},
						&AttrExpr{
							Obj: &Identifier{Name: "x"},
							Attr: &AttrExpr{
								Obj: &IndexExpr{
									Obj:   &Identifier{Name: "y"},
									Index: []int64{1},
								},
								Attr: &Identifier{Name: "z"},
							},
						},
					},
				},
			},
		},
	},

	// case: simple attr syntax
	{
		in: `json(_, x.y.z)`,
		expected: &Ast{
			Functions: []*FuncExpr{
				{
					Name: "json",
					Param: []Node{
						&Identifier{Name: "_"},
						&AttrExpr{
							Obj: &Identifier{Name: "x"},
							Attr: &AttrExpr{
								Obj:  &Identifier{Name: "y"},
								Attr: &Identifier{Name: "z"},
							},
						},
					},
				},
			},
		},
	},
}

func TestParser(t *testing.T) {
	runCases(t, parseCases)
}

func runCases(t *testing.T, cases []*parseCase) {
	for idx := len(cases) - 1; idx >= 0; idx-- {
		tc := cases[idx]
		node, err := ParsePipeline(tc.in)
		if err != nil {
			t.Log(err)
		}

		var ast *Ast

		switch v := node.(type) {
		case *Ast:
			ast = v
		default:
			t.Fatal("should not been here")
		}

		if !tc.fail {
			var x, y string
			x = tc.expected.String()
			y = ast.String()
			testutil.Ok(t, err)
			testutil.Equals(t, x, y)
			t.Logf("[%d] ok %s -> %s", idx, tc.in, y)
		} else {
			t.Logf("[%d] %s -> expect fail: %v", idx, tc.in, err)
			testutil.NotOk(t, err, "")
		}
	}
}
