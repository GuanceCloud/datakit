package parser

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestExprConditions(t *testing.T) {
	cases := []struct {
		in     string
		source string
		tags   map[string]string
		fields map[string]interface{}
		pass   bool
	}{
		{
			in:     "{abc notin [1.1,1.2,1.3] and (a > 1 || c< 0)}",
			fields: map[string]interface{}{"abc": int64(4), "a": int64(-1), "c": int64(-2)},
			pass:   true,
		},

		{
			in:     "{a notin [1,2,3,4]}",
			fields: map[string]interface{}{"a": int64(4)},
			pass:   false,
		},

		{
			in:     "{abc notin [1,2,3]}",
			fields: map[string]interface{}{"abc": int64(4)},
			pass:   true,
		},

		{
			in:     ";;;{a > 1, b > 1 or c > 1, xx != 123 };;;; {xyz > 1};;;",
			fields: map[string]interface{}{"a": int64(2), "c": "xyz"},
			pass:   false,
		},

		{
			in:     "{a > 1, b > 1 or c > 1}",
			fields: map[string]interface{}{"a": int64(2), "c": "xyz"},
			pass:   false,
		},

		{
			in:     "{a > 1, b > 1 or c = 'xyz'}",
			fields: map[string]interface{}{"a": int64(2), "c": "xyz", "b": false},
			pass:   true,
		},

		{
			in:     "{xxx < 111}; {a > 1, b > 1 or c = 'xyz'}",
			fields: map[string]interface{}{"a": int64(2), "c": "xyz", "b": false},
			pass:   true,
		},

		{
			in:     `{host = re("^nginx_.*$")}`,
			fields: map[string]interface{}{"host": "nginx_abc"},
			pass:   true,
		},

		{
			in:     "{host = re(`nginx_*`)}",
			fields: map[string]interface{}{"host": "abcdef"},
			pass:   false,
		},

		// {
		// 	in:     "{host in [re(`mongo_.*`), re(`nginx_.*`), reg(`mysql_.*`)]}",
		// 	fields: map[string]interface{}{"host": "123abc"},
		// 	pass:   false,
		// },
	}

	for _, tc := range cases {
		conditions := GetConds(tc.in)
		tu.Assert(t, conditions != nil, "conditions should not nil")

		tu.Equals(t, tc.pass, conditions.Eval(tc.source, tc.tags, tc.fields))

		t.Logf("[ok] %s => %v, source: %s, tags: %+#v, fields: %+#v", tc.in, tc.pass, tc.source, tc.tags, tc.fields)
	}
}

func TestConditions(t *testing.T) {
	cases := []struct {
		in     WhereConditions
		source string
		tags   map[string]string
		fields map[string]interface{}
		pass   bool
	}{
		{ // multi conditions
			fields: map[string]interface{}{"a": int64(2), "c": "xyz"},
			in: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op:  GT,
							LHS: &Identifier{Name: "b"},
							RHS: &NumberLiteral{IsInt: true, Int: int64(1)},
						},
					},
				},

				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op:  GT,
							LHS: &Identifier{Name: "d"},
							RHS: &NumberLiteral{IsInt: true, Int: int64(1)},
						},
					},
				},
			},
			pass: false,
		},

		{
			fields: map[string]interface{}{"a": int64(2)},
			in: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op:  GT,
							LHS: &Identifier{Name: "a"},
							RHS: &NumberLiteral{IsInt: true, Int: int64(1)},
						},
					},
				},
			},
			pass: true,
		},

		{
			pass:   true,
			fields: map[string]interface{}{"a": "abc"},
			in: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op:  IN,
							LHS: &Identifier{Name: "a"},
							RHS: NodeList{
								&StringLiteral{Val: "123"},
								&StringLiteral{Val: "abc"},
								&NumberLiteral{Float: 123.0},
							},
						},
					},
				},
			},
		},

		{
			pass:   true,
			source: "abc",
			in: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op:  IN,
							LHS: &Identifier{Name: "source"},
							RHS: NodeList{
								&StringLiteral{Val: "xyz"},
								&StringLiteral{Val: "abc"},
								&NumberLiteral{Float: 123.0},
							},
						},
					},
				},
			},
		},

		{
			pass:   true,
			source: "abc",
			in: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op:  NEQ,
							LHS: &Identifier{Name: "source"},
							RHS: &StringLiteral{Val: "xyz"},
						},
					},
				},
			},
		},

		{
			pass:   false,
			source: "abc",
			in: WhereConditions{
				&WhereCondition{
					conditions: []Node{
						&BinaryExpr{
							Op:  EQ,
							LHS: &Identifier{Name: "source"},
							RHS: &StringLiteral{Val: "xyz"},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		tu.Equals(t, tc.pass, tc.in.Eval(tc.source, tc.tags, tc.fields))
		t.Logf("[ok] %s => %v, source: %s, tags: %+#v, fields: %+#v", tc.in, tc.pass, tc.source, tc.tags, tc.fields)
	}
}

func TestBinEval(t *testing.T) {
	cases := []struct {
		op   ItemType
		lhs  interface{}
		rhs  interface{}
		pass bool
	}{
		{
			op:   GT,
			lhs:  int64(4),
			rhs:  int64(3),
			pass: true,
		},

		{
			op:   GTE,
			lhs:  int64(4),
			rhs:  int64(3),
			pass: true,
		},

		{
			op:   GTE,
			lhs:  int64(4),
			rhs:  int64(4),
			pass: true,
		},

		{
			op:   EQ,
			lhs:  int64(3),
			rhs:  int64(3),
			pass: true,
		},

		{
			op:   EQ,
			lhs:  "abc",
			rhs:  "def",
			pass: false,
		},

		{
			op:   LT,
			lhs:  "abc",
			rhs:  "def",
			pass: true,
		},

		{
			op:   LT,
			lhs:  "abc",
			rhs:  123.4,
			pass: false,
		},

		{
			op:   LT,
			lhs:  123.4,
			rhs:  "abc",
			pass: false,
		},

		{
			op:   EQ,
			lhs:  123.4,
			rhs:  "abc",
			pass: false,
		},

		{
			op:   NEQ,
			lhs:  123.4,
			rhs:  "abc",
			pass: false,
		},

		{
			op:   IN,
			lhs:  123.4,
			rhs:  "abc",
			pass: false,
		},
	}

	for _, tc := range cases {
		tu.Equals(t, tc.pass, binEval(tc.op, tc.lhs, tc.rhs))
		t.Logf("[ok] %v %s %v => %v", tc.lhs, tc.op, tc.rhs, tc.pass)
	}
}

func TestEval(t *testing.T) {
	cases := []struct {
		cond   *BinaryExpr
		source string
		tags   map[string]string
		fields map[string]interface{}
		pass   bool
	}{
		{
			fields: map[string]interface{}{"a": int64(3)},
			cond: &BinaryExpr{
				Op:  GTE,
				LHS: &Identifier{Name: "a"},
				RHS: &NumberLiteral{IsInt: true, Int: int64(3)},
			},
			pass: true,
		},

		{
			fields: map[string]interface{}{"a": float64(3.14)},
			cond: &BinaryExpr{
				Op:  GT,
				LHS: &Identifier{Name: "a"},
				RHS: &NumberLiteral{Float: float64(1.1)},
			},
			pass: true,
		},

		{
			source: "abc",
			cond: &BinaryExpr{
				Op:  EQ,
				LHS: &Identifier{Name: "source"},
				RHS: &StringLiteral{Val: "abc"},
			},
			pass: true,
		},

		{
			source: "abc",
			cond: &BinaryExpr{
				Op:  IN,
				LHS: &Identifier{Name: "source"},
				RHS: NodeList{
					&StringLiteral{Val: "abc123"},
					&NumberLiteral{Float: 3.14},
					&StringLiteral{Val: "xyz"},
				},
			},
			pass: false,
		},

		{
			tags: map[string]string{"a": "xyz"},
			cond: &BinaryExpr{
				Op:  IN,
				LHS: &Identifier{Name: "a"},
				RHS: NodeList{
					&StringLiteral{Val: "abc123"},
					&NumberLiteral{Float: 3.14},
					&StringLiteral{Val: "xyz"},
				},
			},
			pass: true,
		},
	}

	for _, tc := range cases {
		t.Logf("[ok] %s => %v", tc.cond, tc.pass)
		tu.Equals(t, tc.pass, tc.cond.Eval(tc.source, tc.tags, tc.fields))
	}
}
