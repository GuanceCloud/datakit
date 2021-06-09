package dql

import (
	"testing"
)

func TestCheckAST(t *testing.T) {
	asts, err := Parse(`F::DQL:(sort(data=dql('M::mem {host="abc"}')))`, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(asts) != 1 {
		t.Fatal(err)
	}

	var fq fq
	if err := fq.checkFuncAST(asts[0]); err != nil {
		t.Error(err)
	}

	if err := fq.parseFuncArgDQL(asts[0]); err != nil {
		t.Error(err)
	}

	for _, fn := range fq.fnList {
		for _, arg := range fn.argList {
			t.Logf("innerQ: %s\nf: %s\nast: %+#v", arg.innerQ, arg.arg.String(), arg.ast)
		}
	}
}
