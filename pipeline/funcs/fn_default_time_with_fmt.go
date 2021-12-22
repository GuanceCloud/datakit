package funcs

import (
	"fmt"
	"reflect"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func DefaultTimeWithFmtChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 2 {
		return fmt.Errorf("func %s expected more than 2 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
	default:
		return fmt.Errorf("param key expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	if len(funcExpr.Param) > 2 {
		switch funcExpr.Param[2].(type) {
		case *parser.StringLiteral:
		default:
			return fmt.Errorf("param key expect StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[2]).String())
		}
	}

	return nil
}

func DefaultTimeWithFmt(ng *parser.Engine, node parser.Node) error {
	var err error
	var goTimeFmt string
	var tz string
	var t time.Time
	timezone := time.Local

	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 2 {
		return fmt.Errorf("func %s expected more than 2 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		goTimeFmt = v.Val
	default:
		return fmt.Errorf("param key expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	if len(funcExpr.Param) > 2 {
		switch v := funcExpr.Param[2].(type) {
		case *parser.StringLiteral:
			tz = v.Val
		default:
			return fmt.Errorf("param key expect StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[2]).String())
		}
	}

	timeStr, err := ng.GetContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if tz != "" {
		timezone, err = time.LoadLocation(tz)
	}

	if err == nil {
		t, err = time.ParseInLocation(goTimeFmt, timeStr, timezone)
	}

	if err != nil {
		l.Debugf("time string: %s, time format: %s, timezone: %s, error msg: %s",
			timeStr, goTimeFmt, tz, err)
		return err
	} else {
		if err := ng.SetContent(key, t.UnixNano()); err != nil {
			l.Warn(err)
			return nil
		}

		return nil
	}
}
