// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/parser"
)

func TestRuntime(t *testing.T) {
	pl := `
	b = 1 + 1
	a = (b + 2) == 4 || False
	c = a * 3 + +100 + -10 + 
		3/1.1
	d = 4 / 3
	e = "dwdw" + "3"
	add_key(e)

	1 != 2
	2 == 2
	2 && 2
	c = 1 / 2 + 1 - 2 * 1 + 3 % 2
	add_key(c)

	map_a = {"a": 1, "b" :2 , "1.1": 2, "nil": [1,2.1,"3"], "1": nil}

	f = map_a["nil"][-1]

	
	aaaa = 1.0 == (b = 1)

	a = v = a
	x7 = [1, 2.1, "3"]
	if b == 2 {
		x = 2
		for i = 1; i < 4; i = i+1 {
			x1 = 1 + x
			e = e + "1"
			if i == 2 {
				break
			}
			continue
			e = e + "2"
		}
	}
	ddd = "" 
	
	# 无序遍历 key
	# for x in {'a': 1, "b":2, "c":3} {
	# 	ddd = ddd + x
	# }

	# add_key(ddd)

	abc = {
		"a": [1,2,3],
		"d": "a",
		"1": 2,
		"d": nil
	}
	add_key(abc)
	abc["a"][-1] = 5
	add_key(abc)
` + ";`aa dw.` = abc;" + "add_key(`aa dw.`)" + `
for x in [1,2,3 ] {
	for y in [1,3,4] {
		if y == 3 {
			break
		}
		continue
	}
	break
}

a  = 1
for ;; {
	if a > 10 {
		break
	}
	a = a + 1
	continue
	a = a - 1
}

for ; a < 12; a = a + 1 {

}

for a; ; {
	if a > 15 {
		break
	}
	a = a + 1
}

for ; a; {
	add_key(a)
	if a > 10 {
		break
	}
} 

for ; ; a= 15 {
	if a > 10 {
		break
	}
} 

for a = 0; a < 12; a = a + 1 {
	if a > 5 {
	  add_key(ef, a)
	  break
	}
	continue
	a = a - 1
  }

add_key(len1, len([12,2]))
add_key(len2, len("123"))
`
	stmts, err := parseScript(pl)
	if err != nil {
		t.Fatal(err)
	}

	script := &Script{
		CallRef: nil,
		FuncCall: map[string]FuncCall{
			"test":    callexprtest,
			"add_key": addkeytest,
			"len":     lentest,
		},
		Name:      "abc",
		Namespace: "default",
		Category:  "",
		FilePath:  "",
		Ast:       stmts,
	}
	err = CheckScript(script, map[string]FuncCheck{
		"add_key": addkeycheck,
		"len":     lencheck,
	})
	if err != nil {
		t.Fatal(err)
	}

	m, tags, f, tn, drop, err := RunScript(script, "s", nil, nil, time.Now())
	t.Log(m, tags, tn, drop)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, map[string]any{
		"aa dw.": `{"1":2,"a":[1,2,5],"d":null}`,
		"abc":    `{"1":2,"a":[1,2,5],"d":null}`,
		"e":      "dwdw3",
		"ef":     int64(6),
		"a":      int64(16),
		"len1":   int64(2),
		"len2":   int64(3),
		"c":      int64(0),
	}, f)
}

func TestInitPoint(t *testing.T) {
	pt := &Point{
		Tags: map[string]string{
			"a1": "1",
		},
		Fields: map[string]any{
			"a": int32(1),
			"b": int64(2),
			"c": float32(3),
			"d": float64(4),
			"e": "5",
			"f": true,
			"g": nil,
		},
	}
	pt2 := &Point{}
	InitPt(pt2, "", pt.Tags, pt.Fields, time.Now())
	assert.Equal(t, &TFMeta{DType: ast.Int, PtFlag: PtField}, pt2.Meta["a"])
	assert.Equal(t, int64(1), pt2.Fields["a"])
	assert.Equal(t, &TFMeta{DType: ast.Int, PtFlag: PtField}, pt2.Meta["b"])
	assert.Equal(t, float64(3), pt2.Fields["c"])
	assert.Equal(t, &TFMeta{DType: ast.Float, PtFlag: PtField}, pt2.Meta["c"])
	assert.Equal(t, &TFMeta{DType: ast.Float, PtFlag: PtField}, pt2.Meta["d"])
	assert.Equal(t, &TFMeta{DType: ast.String, PtFlag: PtField}, pt2.Meta["e"])
	assert.Equal(t, &TFMeta{DType: ast.Bool, PtFlag: PtField}, pt2.Meta["f"])
	assert.Equal(t, &TFMeta{DType: ast.Nil, PtFlag: PtField}, pt2.Meta["g"])
	assert.Equal(t, &TFMeta{DType: ast.String, PtFlag: PtTag}, pt2.Meta["a1"])
}

func TestCondTrue(t *testing.T) {
	cases := []struct {
		val   any
		dtype ast.DType
		ok    bool
	}{
		{int64(1), ast.Int, true},
		{int64(0), ast.Int, false},
		{int64(-1), ast.Int, true},
		{float64(-0.), ast.Float, false},
		{float64(0.0), ast.Float, false},
		{float64(0.00000000000001), ast.Float, true},
		{"", ast.String, false},
		{"a1", ast.String, true},
		{true, ast.Bool, true},
		{false, ast.Bool, false},
		{map[string]any{}, ast.Map, false},
		{map[string]any{"1": nil}, ast.Map, true},
		{[]any{}, ast.List, false},
		{[]any{1}, ast.List, true},
	}

	for i, v := range cases {
		if condTrue(v.val, v.dtype) != v.ok {
			t.Error("idx ", i, v, " ", " ", !v.ok)
		}
	}
}

func TestCondOp(t *testing.T) {
	cases := []struct {
		op         ast.Op
		lhs, rhs   any
		lhsT, rhsT ast.DType
		ok         bool
		fail       bool
	}{
		{
			op:  ast.EQEQ,
			lhs: nil, rhs: nil,
			lhsT: ast.Nil, rhsT: ast.Nil,
			ok: true,
		},
		{
			op:  ast.EQEQ,
			lhs: int64(1), rhs: nil,
			lhsT: ast.Int, rhsT: ast.Nil,
			ok: false,
		},
		{
			op:  ast.EQEQ,
			lhs: int64(1), rhs: float64(1.),
			lhsT: ast.Int, rhsT: ast.Float,
			ok: true,
		},
		{
			op:  ast.EQEQ,
			lhs: int64(1), rhs: true,
			lhsT: ast.Int, rhsT: ast.Bool,
			ok: true,
		},
		{
			op:  ast.EQEQ,
			lhs: true, rhs: true,
			lhsT: ast.Bool, rhsT: ast.Bool,
			ok: true,
		},
		{
			op:  ast.EQEQ,
			lhs: true, rhs: float64(1.),
			lhsT: ast.Bool, rhsT: ast.Float,
			ok: true,
		},
		{
			op:  ast.EQEQ,
			lhs: false, rhs: float64(1.),
			lhsT: ast.Int, rhsT: ast.Float,
			ok: false,
		},
		{
			op:  ast.EQEQ,
			lhs: int64(1), rhs: "",
			lhsT: ast.Int, rhsT: ast.String,
			ok: false,
		},
		{
			op:  ast.EQEQ,
			lhs: "", rhs: "",
			lhsT: ast.String, rhsT: ast.String,
			ok: true,
		},
		{
			op:  ast.EQEQ,
			lhs: "", rhs: true,
			lhsT: ast.String, rhsT: ast.Bool,
			ok: false,
		},
		{
			op:  ast.EQEQ,
			lhs: nil, rhs: true,
			lhsT: ast.Nil, rhsT: ast.Bool,
			ok: false,
		},
		{
			op:  ast.EQEQ,
			lhs: nil, rhs: nil,
			lhsT: ast.Nil, rhsT: ast.Nil,
			ok: true,
		},
		{
			op:   ast.EQEQ,
			lhs:  []any{"a", []any{1, 2, "3", nil}},
			rhs:  []any{"a", []any{1, 2, "3", nil}},
			lhsT: ast.List, rhsT: ast.List,
			ok: true,
		},
		{
			op:   ast.EQEQ,
			lhs:  []any{"a", []any{1, 2, "3", nil}},
			rhs:  []any{"a"},
			lhsT: ast.List, rhsT: ast.List,
			ok: false,
		},
		{
			op:  ast.NEQ,
			lhs: nil, rhs: nil,
			lhsT: ast.Nil, rhsT: ast.Nil,
			ok: false,
		},
		{
			op:  ast.NEQ,
			lhs: int64(1), rhs: nil,
			lhsT: ast.Int, rhsT: ast.Nil,
			ok: true,
		},
		{
			op:  ast.NEQ,
			lhs: int64(1), rhs: float64(1.),
			lhsT: ast.Int, rhsT: ast.Float,
			ok: false,
		},
		{
			op:  ast.NEQ,
			lhs: int64(1), rhs: true,
			lhsT: ast.Int, rhsT: ast.Bool,
			ok: false,
		},
		{
			op:  ast.NEQ,
			lhs: true, rhs: true,
			lhsT: ast.Bool, rhsT: ast.Bool,
			ok: false,
		},
		{
			op:  ast.NEQ,
			lhs: true, rhs: float64(1.),
			lhsT: ast.Bool, rhsT: ast.Float,
			ok: false,
		},
		{
			op:  ast.NEQ,
			lhs: false, rhs: float64(1.),
			lhsT: ast.Int, rhsT: ast.Float,
			ok: true,
		},
		{
			op:  ast.NEQ,
			lhs: int64(1), rhs: "",
			lhsT: ast.Int, rhsT: ast.String,
			ok: true,
		},
		{
			op:  ast.NEQ,
			lhs: "", rhs: "",
			lhsT: ast.String, rhsT: ast.String,
			ok: false,
		},
		{
			op:  ast.NEQ,
			lhs: "", rhs: true,
			lhsT: ast.String, rhsT: ast.Bool,
			ok: true,
		},
		{
			op:  ast.NEQ,
			lhs: nil, rhs: true,
			lhsT: ast.Nil, rhsT: ast.Bool,
			ok: true,
		},
		{
			op:  ast.NEQ,
			lhs: nil, rhs: nil,
			lhsT: ast.Nil, rhsT: ast.Nil,
			ok: false,
		},
		{
			op:   ast.NEQ,
			lhs:  []any{"a", []any{1, 2, "3", nil}},
			rhs:  []any{"a", []any{1, 2, "3", nil}},
			lhsT: ast.List, rhsT: ast.List,
			ok: false,
		},
		{
			op:   ast.NEQ,
			lhs:  []any{"a", []any{1, 2, "3", nil}},
			rhs:  []any{"a"},
			lhsT: ast.List, rhsT: ast.List,
			ok: true,
		},

		{
			op:  ast.AND,
			lhs: true, rhs: false,
			lhsT: ast.Bool, rhsT: ast.Bool,
			ok: false,
		},
		{
			op:  ast.AND,
			lhs: true, rhs: true,
			lhsT: ast.Bool, rhsT: ast.Bool,
			ok: true,
		},
		{
			op:  ast.AND,
			lhs: nil, rhs: true,
			lhsT: ast.Nil, rhsT: ast.Bool,
			// ok: false,
			fail: true,
		},
		{
			op:  ast.AND,
			lhs: false, rhs: nil,
			lhsT: ast.Bool, rhsT: ast.Nil,
			// ok: false,
			fail: true,
		},
		{
			op:  ast.OR,
			lhs: false, rhs: false,
			lhsT: ast.Bool, rhsT: ast.Bool,
			ok: false,
		},
		{
			op:  ast.OR,
			lhs: true, rhs: false,
			lhsT: ast.Bool, rhsT: ast.Bool,
			ok: true,
		},
		{
			op:  ast.OR,
			lhs: true, rhs: true,
			lhsT: ast.Bool, rhsT: ast.Bool,
			ok: true,
		},
		{
			op:  ast.OR,
			lhs: nil, rhs: true,
			lhsT: ast.Nil, rhsT: ast.Bool,
			// ok: true,
			fail: true,
		},
		{
			op:  ast.OR,
			lhs: false, rhs: nil,
			lhsT: ast.Bool, rhsT: ast.Nil,
			// ok: true,
			fail: true,
		},

		{
			op:  ast.LT,
			lhs: true, rhs: int64(0),
			lhsT: ast.Bool, rhsT: ast.Int,
			ok: false,
		},
		{
			op:  ast.LT,
			lhs: false, rhs: int64(0),
			lhsT: ast.Bool, rhsT: ast.Int,
			ok: false,
		},
		{
			op:  ast.LT,
			lhs: false, rhs: float64(1),
			lhsT: ast.Bool, rhsT: ast.Float,
			ok: true,
		},

		{
			op:  ast.LTE,
			lhs: true, rhs: int64(0),
			lhsT: ast.Bool, rhsT: ast.Int,
			ok: false,
		},
		{
			op:  ast.LTE,
			lhs: false, rhs: int64(0),
			lhsT: ast.Bool, rhsT: ast.Int,
			ok: true,
		},
		{
			op:  ast.LTE,
			lhs: false, rhs: float64(1),
			lhsT: ast.Bool, rhsT: ast.Float,
			ok: true,
		},

		{
			op:  ast.LTE,
			lhs: true, rhs: int64(0),
			lhsT: ast.Bool, rhsT: ast.Int,
			ok: false,
		},
		{
			op:  ast.LTE,
			lhs: false, rhs: int64(0),
			lhsT: ast.Bool, rhsT: ast.Int,
			ok: true,
		},
		{
			op:  ast.LTE,
			lhs: false, rhs: float64(1),
			lhsT: ast.Bool, rhsT: ast.Float,
			ok: true,
		},

		{
			op:  ast.GTE,
			lhs: true, rhs: int64(0),
			lhsT: ast.Bool, rhsT: ast.Int,
			ok: true,
		},
		{
			op:  ast.GTE,
			lhs: false, rhs: int64(0),
			lhsT: ast.Bool, rhsT: ast.Int,
			ok: true,
		},
		{
			op:  ast.GTE,
			lhs: false, rhs: float64(1),
			lhsT: ast.Bool, rhsT: ast.Float,
			ok: false,
		},

		{
			op:  ast.GT,
			lhs: true, rhs: int64(0),
			lhsT: ast.Bool, rhsT: ast.Int,
			ok: true,
		},
		{
			op:  ast.GT,
			lhs: false, rhs: int64(0),
			lhsT: ast.Bool, rhsT: ast.Int,
			ok: false,
		},
		{
			op:  ast.GT,
			lhs: false, rhs: float64(1),
			lhsT: ast.Bool, rhsT: ast.Float,
			ok: false,
		},

		{
			op:  ast.EQ,
			lhs: false, rhs: float64(1),
			lhsT: ast.Bool, rhsT: ast.Float,
			// ok: false
			fail: true,
		},
	}

	for k, v := range cases {
		val, dtype, err := condOp(v.lhs, v.rhs, v.lhsT, v.rhsT, v.op)
		if err != nil {
			if v.fail {
				continue
			}
			t.Error(err, v)
			continue
		}
		if condTrue(val, dtype) != v.ok {
			t.Error("idx", k, val, v)
		}
	}
}

func parseScript(content string) (ast.Stmts, error) {
	return parser.ParsePipeline(content)
}

func callexprtest(ctx *Context, callExpr *ast.CallExpr) PlPanic {
	return nil
}

func addkeytest(ctx *Context, callExpr *ast.CallExpr) PlPanic {
	key := callExpr.Param[0].Identifier.Name
	if len(callExpr.Param) > 1 {
		val, dtype, err := RunStmt(ctx, callExpr.Param[1])
		if err != nil {
			return err
		}
		return ctx.AddKey2PtWithVal(key, val, dtype, KindPtDefault)
	}
	return ctx.AddKey2Pt(key, KindPtDefault)
}

func addkeycheck(ctx *Context, callExpr *ast.CallExpr) error {
	return nil
}

func lentest(ctx *Context, callExpr *ast.CallExpr) PlPanic {
	val, dtype, err := RunStmt(ctx, callExpr.Param[0])
	if err != nil {
		return err
	}
	switch dtype { //nolint:exhaustive
	case ast.String:
		ctx.Regs.Append(int64(len(val.(string))), ast.Int)
	case ast.List:
		ctx.Regs.Append(int64(len(val.([]any))), ast.Int)
	case ast.Map:
		ctx.Regs.Append(int64(len(val.(map[string]any))), ast.Int)
	default:
		ctx.Regs.Append(int64(0), ast.Int)
	}
	return nil
}

func lencheck(ctx *Context, callexpr *ast.CallExpr) error {
	return nil
}
