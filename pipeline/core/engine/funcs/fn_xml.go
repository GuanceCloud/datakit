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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func XMLChecking(_ *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expects 3 args", funcExpr.Name)
	}

	// XML field key.
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	// XPath expression.
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
	}

	// Field name.
	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
	default:
		return fmt.Errorf("expect AttrExpr or Identifier or StringLiteral, got %s",
			funcExpr.Param[2].NodeType)
	}

	return nil
}

func XML(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	var (
		xmlKey, fieldName string
		xpathExpr         string
	)

	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expects 3 args", funcExpr.Name)
	}

	// XML field key.
	xmlKey, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	// XPath expression
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		xpathExpr = funcExpr.Param[1].StringLiteral.Val
	default:
		return fmt.Errorf("expect StringLiteStringLiteral:ral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	// Field name.

	fieldName, err = getKeyName(funcExpr.Param[2])
	if err != nil {
		return err
	}

	cont, err := ctx.GetKeyConv2Str(xmlKey)
	if err != nil {
		l.Debug(err)
		return nil
	}

	doc, err := xmlquery.Parse(strings.NewReader(cont))
	if err != nil {
		l.Debug(err)
		return nil
	}
	// xmlquery already caches the compiled expression for us.
	dest, err := xmlquery.Query(doc, xpathExpr)
	if err != nil {
		l.Debug(err)
		return nil
	}
	if dest == nil {
		err = fmt.Errorf("can't find any XML Node that matches the XPath expr: %s", xpathExpr)
		l.Debug(err)
		return nil
	}

	if err := ctx.AddKey2PtWithVal(fieldName, dest.InnerText(), ast.String,
		runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}
