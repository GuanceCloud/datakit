package parser

import (
	"testing"
)

var in = `avg(x.y.z,1,2,3, p68, re("cd"), pqa);

f(g(f("abc"), 123),
1,2,3);


kkk(g(1,2),h(2,3),[5,6,7])
`

func FuncPrintHelp(f *FuncExpr, prev string, t *testing.T) {
	t.Logf("%v%v", prev, f.Name)

	for _, node := range f.Param {
		switch v := node.(type) {
		case *FuncExpr:
			FuncPrintHelp(v, prev+"    ", t)
		default:
			t.Logf("%v%v", prev+"    |", node)
		}
	}
}
func TestParseFuncExpr(t *testing.T) {
	fexpr, err := ParseFuncExpr(in)
	if err != nil {
		t.Error(err)
		return
	}

	for _, fex := range fexpr {
		f, _ := fex.(*FuncExpr)
		FuncPrintHelp(f, "", t)
	}
}
