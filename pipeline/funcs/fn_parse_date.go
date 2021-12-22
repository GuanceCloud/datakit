package funcs

import (
	"fmt"
	"reflect"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func ParseDateChecking(node parser.Node) error {
	funcExpr, ok := node.(*parser.FuncStmt)
	if !ok {
		return fmt.Errorf("expect function expr")
	}
	for _, p := range funcExpr.Param {
		switch p.(type) {
		case *parser.StringLiteral, *parser.Identifier, *parser.AssignmentStmt:
		default:
			return fmt.Errorf("param key expects StringLiteral or Identifier or AssignmentStmt, got %s",
				reflect.TypeOf(p).String())
		}
	}
	return nil
}

func ParseDate(ng *parser.Engine, node parser.Node) error {
	funcExpr, ok := node.(*parser.FuncStmt)
	if !ok {
		return fmt.Errorf("expect function expr")
	}

	now := time.Now()
	var yy, dd, hh, mi, ss, ms, us, ns int
	var mm time.Month
	var key string
	var zone *time.Location

	if funcExpr.KwParam == nil {
		funcExpr.KwParam = make(map[string]parser.Node)
		for _, p := range funcExpr.Param {
			if st, ok := p.(*parser.AssignmentStmt); ok {
				funcExpr.KwParam[st.LHS.String()] = st.RHS
			}
		}
	}

	if x, err := ng.GetFuncStrArg(funcExpr, 0, "key"); err != nil {
		return err
	} else {
		key = x
	}

	// year
	if x, err := ng.GetFuncIntArg(funcExpr, 1, "y"); err != nil {
		return err
	} else {
		yy, err = fixYear(now, x)
		if err != nil {
			return err
		}
	}

	// month
	if x, err := ng.GetFuncStrArg(funcExpr, 2, "m"); err != nil {
		return err
	} else {
		mm, err = fixMonth(now, x)
		if err != nil {
			return err
		}
	}

	// check day
	if x, err := ng.GetFuncIntArg(funcExpr, 3, "d"); err != nil {
		return err
	} else {
		dd, err = fixDay(now, x)
		if err != nil {
			return err
		}
	}

	// check hour
	if x, err := ng.GetFuncIntArg(funcExpr, 4, "h"); err != nil {
		return err
	} else {
		hh, err = fixHour(now, x)
		if err != nil {
			return err
		}
	}

	// check minute
	if x, err := ng.GetFuncIntArg(funcExpr, 5, "M"); err != nil {
		return err
	} else {
		mi, err = fixMinute(now, x)
		if err != nil {
			return err
		}
	}

	// check second
	if x, err := ng.GetFuncIntArg(funcExpr, 6, "s"); err != nil {
		return err
	} else {
		ss, err = fixSecond(now, x)
		if err != nil {
			return err
		}
	}

	if x, err := ng.GetFuncIntArg(funcExpr, 7, "ms"); err != nil {
		return err
	} else {
		ms = int(x)
		if x == DefaultInt {
			ms = 0
		}
	}

	if x, err := ng.GetFuncIntArg(funcExpr, 8, "us"); err != nil {
		return err
	} else {
		us = int(x)
		if x == DefaultInt {
			us = 0
		}
	}

	if x, err := ng.GetFuncIntArg(funcExpr, 9, "ns"); err != nil {
		return err
	} else {
		ns = int(x)
		if x == DefaultInt {
			ns = 0
		}
	}

	if x, err := ng.GetFuncStrArg(funcExpr, 10, "zone"); err != nil {
		return err
	} else {
		if x == "" {
			zone = time.UTC
		} else {
			zone, err = tz(x)
			if err != nil {
				return err
			}
		}
	}

	t := time.Date(yy, mm, dd, hh, mi, ss, ms*1000*1000+us*1000+ns, zone)
	ng.SetKey(key, t.UnixNano())
	return nil
}
