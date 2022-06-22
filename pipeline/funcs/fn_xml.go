// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/antchfx/xmlquery"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func XMLChecking(_ *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expects 3 args", funcExpr.Name)
	}

	// XML field key.
	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	// XPath expression.
	switch funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	// Field name.
	switch funcExpr.Param[2].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
	default:
		return fmt.Errorf("expect AttrExpr or Identifier or StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	return nil
}

func XML(ng *parser.EngineData, node parser.Node) interface{} {
	var (
		xmlKey, fieldName parser.Node
		xpathExpr         *parser.StringLiteral
	)
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expects 3 args", funcExpr.Name)
	}

	// XML field key.
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		xmlKey = v
	default:
		return fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	// XPath expression.
	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		xpathExpr = v
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	// Field name.
	switch v := funcExpr.Param[2].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		fieldName = v
	default:
		return fmt.Errorf("expect AttrExpr or Identifier or StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	cont, err := ng.GetContentStr(xmlKey)
	if err != nil {
		l.Debug(err)
		return nil
	}

	doc, err := xmlquery.Parse(strings.NewReader(cont))
	if err != nil {
		l.Warn(err)
		return nil
	}
	// xmlquery already caches the compiled expression for us.
	dest, err := xmlquery.Query(doc, xpathExpr.Val)
	if err != nil {
		l.Warn(err)
		return nil
	}
	if dest == nil {
		l.Warnf("can't find any XML Node that matches the XPath expr: %s", xpathExpr)
		return nil
	}

	if err := ng.SetContent(fieldName, dest.InnerText()); err != nil {
		l.Warn(err)
		return nil
	}

	return nil
}
