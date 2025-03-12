// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/antchfx/xmlquery"
)

func XMLChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 3 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 3 args", funcExpr.Name), funcExpr.NamePos)
	}

	// XML field key.
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	// XPath expression.
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	// Field name.
	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect AttrExpr or Identifier or StringLiteral, got %s",
			funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
	}

	return nil
}

func XML(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	var (
		xmlKey, fieldName string
		xpathExpr         string
	)

	if len(funcExpr.Param) != 3 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 3 args", funcExpr.Name), funcExpr.NamePos)
	}

	// XML field key.
	xmlKey, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	// XPath expression
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		xpathExpr = funcExpr.Param[1].StringLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("expect StringLiteStringLiteral:ral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String()), funcExpr.Param[1].StartPos())
	}

	// Field name.

	fieldName, err = getKeyName(funcExpr.Param[2])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[2].StartPos())
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

	_ = addKey2PtWithVal(ctx.InData(), fieldName, dest.InnerText(),
		ast.String, ptinput.KindPtDefault)

	return nil
}
