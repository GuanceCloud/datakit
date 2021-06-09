package dql

import (
	"encoding/json"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
)

var (
	whereCases = []struct {
		in   parser.Node
		fail bool
	}{
		{
			in: &parser.BinaryExpr{
				LHS: &parser.Identifier{Name: "EQ1"},
				Op:  parser.EQ,
				RHS: &parser.Identifier{Name: "EQ2"},
			},
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.Identifier{Name: "NEQ1"},
				Op:  parser.NEQ,
				RHS: &parser.Identifier{Name: "NEQ2"},
			},
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.Identifier{Name: "LTE1"},
				Op:  parser.LTE,
				RHS: &parser.Identifier{Name: "LTE2"},
			},
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.Identifier{Name: "LT1"},
				Op:  parser.LT,
				RHS: &parser.Identifier{Name: "LT2"},
			},
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.Identifier{Name: "GTE1"},
				Op:  parser.GTE,
				RHS: &parser.Identifier{Name: "GTE2"},
			},
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.Identifier{Name: "GT1"},
				Op:  parser.GT,
				RHS: &parser.Identifier{Name: "GT2"},
			},
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.Identifier{Name: "SUB1"},
				Op:  parser.SUB,
				RHS: &parser.Identifier{Name: "SUB2"},
			},
			fail: true,
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.Identifier{Name: "MOD1"},
				Op:  parser.MOD,
				RHS: &parser.Identifier{Name: "MOD2"},
			},
			fail: true,
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.Identifier{Name: "OR1"},
				Op:  parser.OR,
				RHS: &parser.Identifier{Name: "OR2"},
			},
			fail: true,
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.StringLiteral{Val: "string1"},
				Op:  parser.EQ,
				RHS: &parser.StringLiteral{Val: "string2"},
			},
			fail: true,
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.BinaryExpr{
					LHS: &parser.Identifier{Name: "binary1"},
					Op:  parser.NEQ,
					RHS: &parser.NumberLiteral{IsInt: true, Int: 16},
				},
				Op:  parser.EQ,
				RHS: &parser.StringLiteral{Val: "binary2"},
			},
			fail: true,
		},
		{
			in: &parser.BinaryExpr{
				LHS: &parser.ParenExpr{Param: &parser.Identifier{Name: "paren1"}},
				Op:  parser.EQ,
				RHS: &parser.StringLiteral{Val: "binary2"},
			},
			fail: true,
		},
		{
			in:   &parser.StringLiteral{Val: "stringLiteral"},
			fail: true,
		},
	}
)

func TestCheckWithConditions(t *testing.T) {
	var err error

	for idx, c := range whereCases {
		lq := &lq{
			lambdaAST: &parser.Lambda{
				WhereCondition: []parser.Node{c.in},
			},
		}

		err = lq.checkWithExpr()
		if err != nil {
			if c.fail {
				t.Logf("[%d] OK -> CheckErr: %s\n", idx, err)
			} else {
				t.Fatalf("\033[1;31m[%d] Undefine Err: %s\033[0m\n", idx, err)
			}
		} else {

			t.Logf("[%d] OK -> Pass\n", idx)
		}
	}

	t.Log("END")
}

func TestContrast(t *testing.T) {
	var (
		tests = []struct {
			x, y     interface{}
			operator string
			fail     bool
		}{
			{
				x:        3.1415,
				operator: "=",
				y:        3.1415,
			},
			{
				x:        3.1415,
				operator: "=",
				y:        12.25,
			},
			{
				x:        3.1415,
				operator: "!=",
				y:        12.25,
			},
			{
				x:        3.1415,
				operator: ">",
				y:        12.25,
			},
			{
				x:        3.1415,
				operator: ">=",
				y:        12.25,
			},
			{
				x:        3.1415,
				operator: "<",
				y:        12.25,
			},
			{
				x:        3.1415,
				operator: "<=",
				y:        12.25,
			},
			{
				x:        3,
				operator: "<=",
				y:        12.25,
				fail:     true,
			},
			{
				x:        int64(3),
				operator: "<=",
				y:        12.25,
			},
			{
				x:        int64(3),
				operator: "!=",
				y:        12.25,
			},
			{
				x:        json.Number("10"),
				operator: "=",
				y:        json.Number("10.0"),
			},
			{
				x:        "ABCD",
				operator: "=",
				y:        "ABCD",
			},
			{
				x:        "ABCD",
				operator: "!=",
				y:        "ABCDEEEEEE",
			},
			{
				x:        "ABCD",
				operator: "<=",
				y:        "ABCD",
				fail:     true,
			},
			{
				x:        true,
				operator: "=",
				y:        true,
			},
			{
				x:        true,
				operator: "!=",
				y:        true,
			},
			{
				x:        true,
				operator: "<=",
				y:        false,
				fail:     true,
			},
		}
	)

	var b bool

	for idx, ts := range tests {
		b = contrast(ts.x, ts.operator, ts.y)
		if b {
			t.Logf("[%d] OK, pass: (%v %s %v)\n", idx, ts.x, ts.operator, ts.y)
		} else {
			t.Logf("[%d] OK, not:  (%v %s %v)\n", idx, ts.x, ts.operator, ts.y)
		}
	}

	t.Log("END")
}
