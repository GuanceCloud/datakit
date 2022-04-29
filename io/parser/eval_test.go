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
			in:     "{ abc notmatch []}",
			fields: map[string]interface{}{"abc": "abc123"},
			pass:   false,
		},

		{
			in:     "{ abc match ['g(-z]+ng wrong regex']} # invalid regexp",
			fields: map[string]interface{}{"abc": "abc123"},
			pass:   false,
		},

		{
			in:     "{ abc match ['a.*']}",
			fields: map[string]interface{}{"abc": "abc123"},
			pass:   true,
		},

		{
			in:     "{ abc match ['a.*']}",
			fields: map[string]interface{}{"abc": "abc123"},
			pass:   true,
		},

		{
			in:     "{ source = re(`.*`) and (abc match ['a.*'])}",
			fields: map[string]interface{}{"abc": "abc123"},
			tags:   map[string]string{"source": "12345"},
			pass:   true,
		},

		{
			in:     "{ abc notmatch ['a.*'] or xyz match ['.*']}",
			fields: map[string]interface{}{"abc": "abc123"},
			tags:   map[string]string{"xyz": "def"},
			pass:   true,
		},

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
		t.Run(tc.in, func(t *testing.T) {
			conditions := GetConds(tc.in)

			tu.Equals(t, tc.pass, conditions.Eval(tc.tags, tc.fields))

			t.Logf("[ok] %s => %v, source: %s, tags: %+#v, fields: %+#v", tc.in, tc.pass, tc.source, tc.tags, tc.fields)
		})
	}
}

func TestConditions(t *testing.T) {
	cases := []struct {
		in     WhereConditions
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
			pass: true,
			tags: map[string]string{"source": "abc"},
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
			pass: true,
			tags: map[string]string{"source": "abc"},
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
			pass: false,
			tags: map[string]string{"source": "abc"},
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
		t.Run(tc.in.String(), func(t *testing.T) {
			tu.Equals(t, tc.pass, tc.in.Eval(tc.tags, tc.fields))
			t.Logf("[ok] %s => %v,  tags: %+#v, fields: %+#v", tc.in, tc.pass, tc.tags, tc.fields)
		})
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
			tags: map[string]string{"source": "abc"},
			cond: &BinaryExpr{
				Op:  EQ,
				LHS: &Identifier{Name: "source"},
				RHS: &StringLiteral{Val: "abc"},
			},
			pass: true,
		},

		{
			tags: map[string]string{"source": "abc"},
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
		t.Run(tc.cond.String(), func(t *testing.T) {
			t.Logf("[ok] %s => %v", tc.cond, tc.pass)
			tu.Equals(t, tc.pass, tc.cond.Eval(tc.tags, tc.fields))
		})
	}
}

func BenchmarkRegexp(b *testing.B) {
	// cliutils.CreateRandomString()
}
