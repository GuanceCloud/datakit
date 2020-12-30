package parser

import (
	"testing"

	//"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

type parseCase struct {
	in       string
	expected interface{}
	err      string
	fail     bool
}
var in = `count(avg(2,5));
count2(avg2(2,5));
count3(avg2(2,5));count4(avg2(2,5));
count6(avg2(2,5))
count5(avg2(21,51))
`
func TestParseFuncExpr(t *testing.T) {
	var funcExprCases = []*parseCase{
		{
			in: "#count(avg(2,5));count1(avg2(2,5))",
			expected: &Funcs{
				&FuncExpr{
					Name: "count",
					Param: []Node{
						&FuncExpr{
							Name: "avg",
							Param: []Node{
								&NumberLiteral{IsInt: true, Int: 2},
								&NumberLiteral{IsInt: true, Int: 5},
							},
						},
					},
				},
			},
		},

	}

	for idx, _ := range funcExprCases {
		_ = idx

		fexpr, err := ParseFuncExpr(in)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%v", len(fexpr))

		for _, fex := range fexpr {
			t.Logf("----------------------------\n")
			f, _ := fex.(*FuncExpr)
			t.Logf("%v", f.Name)
			ff, _ := f.Param[0].(*FuncExpr)
			t.Logf("%v", ff.Name)
			fp0 := ff.Param[0].(*NumberLiteral)
			t.Logf("%v", fp0.Int)
			fp1 := ff.Param[1].(*NumberLiteral)
			t.Logf("%v", fp1.Int)

		}

		//var x, y string

		//if err != nil {
		//	t.Logf("parse error: %s", err.Error())
		//} else {
		//	if tc.expected != nil {
		//		t.Logf("tc.expected: %+#v", tc.expected)
		//		expected, ok := tc.expected.(Funcs)
		//		if ok {
		//			x = expected.String()
		//		}
		//		y = fexpr.String()
		//	}
		//}

		//if !tc.fail {
		//	testutil.Ok(t, err)
		//	testutil.Equals(t, x, y)
		//	t.Logf("[%d] ok %s -> %s", idx, tc.in, y)
		//} else {
		//	t.Logf("[%d] %s -> expect fail: %v", idx, tc.in, err)
		//	testutil.NotOk(t, err, "")
		//}
	}
}
