// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"regexp"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func MatchChecking(ng *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 2 {
		return fmt.Errorf("func %s expected 2", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
	default:
		return fmt.Errorf("expect AttrExpr , Identifier or StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
	default:
		return fmt.Errorf("expect AttrExpr , Identifier or StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}
	return nil
}

func Match(ng *parser.EngineData, node parser.Node) interface{} {
	funcExpr := fexpr(node)
	var cont string
	var err error
	var res bool

	var words, pattern parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		words = v
	default:
		l.Debugf("expect AttrExpr or Identifier or StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
		return false
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		pattern = v
	default:
		l.Debugf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
		return false
	}

	if words.String() == "_" {
		cont, err = ng.GetContentStr(words)
		if err != nil {
			l.Debug(err)
			return false
		}
	} else {
		cont = words.String()
	}

	res, err = isMatch(cont, pattern.String())
	if err != nil {
		l.Debug(err)
		return false
	}
	if res {
		return "true"
	} else {
		return "false"
	}
}

func isMatch(words string, pattern string) (bool, error) {
	isMatch, err := regexp.MatchString(pattern, "'"+words+"'")
	if err != nil {
		return false, err
	}
	return isMatch, err
}
