package testutil

import (
	"reflect"
)

type TB interface {
	Helper()
	Fatalf(string, ...interface{})
}

func Assert(tb TB, condition bool, fmt string, a ...interface{}) {
	tb.Helper()
	if !condition {
		tb.Fatalf("\033[31m"+fmt+"\033[39m\n", a...)
	}
}

func Ok(tb TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("\033[31munexpected error: %v\033[39m\n", err)
	}
}

func NotOk(tb TB, err error, fmt string, a ...interface{}) {
	tb.Helper()
	if err == nil {
		if len(a) != 0 {
			tb.Fatalf("\033[31m"+fmt+": expected error, got none\033[39m", a...)
		}
		tb.Fatalf("\033[31mexpected error, got none\033[39m")
	}
}

func Equals(tb TB, exp, act interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(exp, act) {
		tb.Fatalf("\033[31m\nexp: %#v\n\ngot: %#v\033[39m\n", exp, act)
	}
}
